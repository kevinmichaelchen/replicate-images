package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/kevinmichaelchen/replicate-images/internal/cache"
	"github.com/kevinmichaelchen/replicate-images/internal/client"
	"github.com/kevinmichaelchen/replicate-images/internal/convert"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Exit codes for agent-friendly operation.
const (
	ExitSuccess       = 0 // All operations succeeded
	ExitPartialFail   = 1 // Some generations failed
	ExitTotalFail     = 2 // All generations failed
	ExitInvalidInput  = 3 // Invalid input (bad YAML, missing file, etc.)
)

var (
	flagModel       string
	flagOutput      string
	flagNoCache     bool
	flagConcurrency int
	flagJSON        bool
	flagDryRun      bool
	flagQuiet       bool
)

// GenerateResult represents the JSON output for a single generation.
type GenerateResult struct {
	Status     string `json:"status"`
	Prompt     string `json:"prompt"`
	Model      string `json:"model"`
	Hash       string `json:"hash"`
	OutputFile string `json:"output_file,omitempty"`
	Cached     bool   `json:"cached"`
	Error      string `json:"error,omitempty"`
}

// DryRunResult represents the JSON output for a dry-run.
type DryRunResult struct {
	ToGenerate int              `json:"to_generate"`
	Cached     int              `json:"cached"`
	Prompts    []DryRunPrompt   `json:"prompts"`
}

// DryRunPrompt represents a single prompt in dry-run output.
type DryRunPrompt struct {
	Prompt     string `json:"prompt"`
	Model      string `json:"model"`
	Hash       string `json:"hash"`
	Status     string `json:"status"`
	OutputFile string `json:"output_file,omitempty"`
}

// ExitError represents an error with a specific exit code.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return e.Message
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if exitErr, ok := err.(*ExitError); ok {
			if exitErr.Message != "" && !flagJSON {
				fmt.Fprintln(os.Stderr, exitErr.Message)
			}
			os.Exit(exitErr.Code)
		}
		if !flagJSON {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(ExitPartialFail)
	}
}

var rootCmd = &cobra.Command{
	Use:           "replicate-images [prompt]",
	Short:         "Generate images from text prompts using Replicate",
	SilenceErrors: true,
	SilenceUsage:  true,
	Long: `A CLI tool that generates images from text prompts using Replicate's API.

Images are cached based on prompt+model hash to avoid regenerating duplicates.
Output files are saved as WEBP in the output directory.`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

var modelsCmd = &cobra.Command{
	Use:   "models [query]",
	Short: "Search for popular text-to-image models",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runModels,
}

var batchCmd = &cobra.Command{
	Use:   "batch <prompts.yaml>",
	Short: "Generate images from a YAML file of prompts",
	Long: `Process a YAML file containing multiple prompt/model combinations.

Example prompts.yaml:
  prompts:
    - prompt: "a cat in space"
      model: black-forest-labs/flux-schnell
    - prompt: "a dog on the moon"
    - prompt: "a bird underwater"
      model: stability-ai/sdxl

Prompts without a model use the default or --model flag value.
Existing cached images are skipped unless --no-cache is set.`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

var validateCmd = &cobra.Command{
	Use:   "validate <prompts.yaml>",
	Short: "Validate a prompts YAML file without generating",
	Long: `Check a prompts YAML file for syntax errors and structural issues.

Validates:
  - YAML syntax
  - Required fields (prompt)
  - Empty prompts
  - Duplicate prompt/model combinations`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "./generated-images", "Output directory")
	rootCmd.PersistentFlags().BoolVar(&flagNoCache, "no-cache", false, "Force regeneration, ignore cache")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output results as JSON (JSONL for batch)")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be generated without executing")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Replicate model to use")

	batchCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Default model for prompts without one")
	batchCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 3, "Number of concurrent generations")

	validateCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Default model for prompts without one")

	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(batchCmd)
	rootCmd.AddCommand(validateCmd)
}

// shouldOutput returns true if human-readable output should be shown.
func shouldOutput() bool {
	return !flagJSON && !flagQuiet
}

func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	prompt := args[0]

	hash := cache.Hash(prompt, flagModel)

	// For dry-run, we only need to check the cache
	if flagDryRun {
		c, err := cache.Load(flagOutput)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load cache: %w", err)
		}

		status := "pending"
		var outputFile string
		if !flagNoCache && c != nil {
			if entry := c.Lookup(hash); entry != nil {
				outputPath := filepath.Join(flagOutput, entry.OutputFile)
				if _, err := os.Stat(outputPath); err == nil {
					status = "cached"
					outputFile = outputPath
				}
			}
		}

		result := DryRunResult{
			ToGenerate: 0,
			Cached:     0,
			Prompts: []DryRunPrompt{{
				Prompt:     prompt,
				Model:      flagModel,
				Hash:       hash,
				Status:     status,
				OutputFile: outputFile,
			}},
		}
		if status == "cached" {
			result.Cached = 1
		} else {
			result.ToGenerate = 1
		}

		if flagJSON {
			outputJSON(result)
		} else if shouldOutput() {
			fmt.Printf("Dry run: %s\n", prompt)
			fmt.Printf("  Model:  %s\n", flagModel)
			fmt.Printf("  Hash:   %s\n", hash)
			fmt.Printf("  Status: %s\n", status)
			if outputFile != "" {
				fmt.Printf("  File:   %s\n", outputFile)
			}
		}
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(flagOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load cache
	c, err := cache.Load(flagOutput)
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	// Check cache
	if !flagNoCache {
		if entry := c.Lookup(hash); entry != nil {
			outputPath := filepath.Join(flagOutput, entry.OutputFile)
			if _, err := os.Stat(outputPath); err == nil {
				if flagJSON {
					outputJSON(GenerateResult{
						Status:     "cached",
						Prompt:     prompt,
						Model:      flagModel,
						Hash:       hash,
						OutputFile: outputPath,
						Cached:     true,
					})
				} else if shouldOutput() {
					fmt.Printf("Using cached image: %s\n", outputPath)
				}
				return nil
			}
		}
	}

	// Create client
	rc, err := client.New()
	if err != nil {
		return err
	}

	if shouldOutput() {
		fmt.Printf("Generating image with %s...\n", flagModel)
	}

	// Generate image
	data, url, err := rc.GenerateImage(ctx, flagModel, prompt)
	if err != nil {
		if flagJSON {
			outputJSON(GenerateResult{
				Status: "error",
				Prompt: prompt,
				Model:  flagModel,
				Hash:   hash,
				Error:  err.Error(),
			})
			return nil
		}
		return err
	}

	if shouldOutput() {
		fmt.Printf("Downloaded from: %s\n", url)
	}

	// Convert to WEBP and save
	filename := hash + ".webp"
	outputPath := filepath.Join(flagOutput, filename)

	if err := convert.SaveWebP(data, outputPath); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Update cache
	c.Upsert(prompt, flagModel, filename)
	if err := c.Save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	if flagJSON {
		outputJSON(GenerateResult{
			Status:     "generated",
			Prompt:     prompt,
			Model:      flagModel,
			Hash:       hash,
			OutputFile: outputPath,
			Cached:     false,
		})
	} else if shouldOutput() {
		fmt.Printf("Saved: %s\n", outputPath)
	}
	return nil
}

func outputJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(v)
}

func runModels(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	query := "text to image"
	if len(args) > 0 {
		query = args[0]
	}

	rc, err := client.New()
	if err != nil {
		return err
	}

	models, err := rc.SearchModels(ctx, query)
	if err != nil {
		return err
	}

	// Sort by run count (popularity)
	sort.Slice(models, func(i, j int) bool {
		return models[i].RunCount > models[j].RunCount
	})

	fmt.Printf("Popular models for %q:\n\n", query)
	for i, m := range models {
		if i >= 10 {
			break
		}
		fmt.Printf("  %s\n", m.FullName())
		fmt.Printf("    Runs: %d\n", m.RunCount)
		if m.Description != "" {
			desc := m.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
		fmt.Println()
	}

	return nil
}

// PromptFile represents the YAML structure for batch processing.
type PromptFile struct {
	Prompts []PromptEntry `yaml:"prompts"`
}

// PromptEntry represents a single prompt/model combination.
type PromptEntry struct {
	Prompt string `yaml:"prompt"`
	Model  string `yaml:"model,omitempty"`
}

func runBatch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Read and parse YAML file
	data, err := os.ReadFile(args[0])
	if err != nil {
		return &ExitError{Code: ExitInvalidInput, Message: fmt.Sprintf("failed to read file: %v", err)}
	}

	var pf PromptFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return &ExitError{Code: ExitInvalidInput, Message: fmt.Sprintf("failed to parse YAML: %v", err)}
	}

	if len(pf.Prompts) == 0 {
		return &ExitError{Code: ExitInvalidInput, Message: "no prompts found in file"}
	}

	// Load cache (don't create output dir for dry-run)
	var c *cache.Cache
	if flagDryRun {
		c, err = cache.Load(flagOutput)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load cache: %w", err)
		}
		if c == nil {
			c = &cache.Cache{}
		}
	} else {
		// Ensure output directory exists
		if err := os.MkdirAll(flagOutput, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		c, err = cache.Load(flagOutput)
		if err != nil {
			return fmt.Errorf("failed to load cache: %w", err)
		}
	}

	// Categorize prompts
	var (
		toGenerate []PromptEntry
		dryPrompts []DryRunPrompt
		cachedCount int
	)

	for _, p := range pf.Prompts {
		model := p.Model
		if model == "" {
			model = flagModel
		}

		hash := cache.Hash(p.Prompt, model)
		isCached := false

		if !flagNoCache {
			if entry := c.Lookup(hash); entry != nil {
				outputPath := filepath.Join(flagOutput, entry.OutputFile)
				if _, err := os.Stat(outputPath); err == nil {
					isCached = true
					cachedCount++

					if flagDryRun {
						dryPrompts = append(dryPrompts, DryRunPrompt{
							Prompt:     p.Prompt,
							Model:      model,
							Hash:       hash,
							Status:     "cached",
							OutputFile: outputPath,
						})
					} else if flagJSON {
						outputJSON(GenerateResult{
							Status:     "cached",
							Prompt:     p.Prompt,
							Model:      model,
							Hash:       hash,
							OutputFile: outputPath,
							Cached:     true,
						})
					} else if shouldOutput() {
						fmt.Printf("Cached: %s\n", p.Prompt)
					}
				}
			}
		}

		if !isCached {
			toGenerate = append(toGenerate, PromptEntry{Prompt: p.Prompt, Model: model})
			if flagDryRun {
				dryPrompts = append(dryPrompts, DryRunPrompt{
					Prompt: p.Prompt,
					Model:  model,
					Hash:   hash,
					Status: "pending",
				})
			}
		}
	}

	// Handle dry-run output
	if flagDryRun {
		result := DryRunResult{
			ToGenerate: len(toGenerate),
			Cached:     cachedCount,
			Prompts:    dryPrompts,
		}

		if flagJSON {
			outputJSON(result)
		} else if shouldOutput() {
			fmt.Printf("Dry run summary:\n")
			fmt.Printf("  To generate: %d\n", result.ToGenerate)
			fmt.Printf("  Cached:      %d\n", result.Cached)
			fmt.Printf("  Total:       %d\n\n", len(pf.Prompts))

			for _, p := range dryPrompts {
				fmt.Printf("  [%s] %s\n", p.Status, p.Prompt)
				fmt.Printf("         Model: %s\n", p.Model)
				fmt.Printf("         Hash:  %s\n", p.Hash)
				if p.OutputFile != "" {
					fmt.Printf("         File:  %s\n", p.OutputFile)
				}
			}
		}
		return nil
	}

	if len(toGenerate) == 0 {
		if shouldOutput() {
			fmt.Println("All images already cached.")
		}
		return nil
	}

	// Create client (only needed if actually generating)
	rc, err := client.New()
	if err != nil {
		return err
	}

	if shouldOutput() {
		fmt.Printf("Generating %d images (concurrency: %d)...\n\n", len(toGenerate), flagConcurrency)
	}

	// Process with concurrency limit
	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, flagConcurrency)
		mu      sync.Mutex
		errored int
	)

	for _, p := range toGenerate {
		wg.Add(1)
		sem <- struct{}{}

		go func(prompt, model string) {
			defer wg.Done()
			defer func() { <-sem }()

			hash := cache.Hash(prompt, model)
			filename := hash + ".webp"
			outputPath := filepath.Join(flagOutput, filename)

			data, _, err := rc.GenerateImage(ctx, model, prompt)
			if err != nil {
				mu.Lock()
				if flagJSON {
					outputJSON(GenerateResult{
						Status: "error",
						Prompt: prompt,
						Model:  model,
						Hash:   hash,
						Error:  err.Error(),
					})
				} else {
					fmt.Printf("Error [%s]: %v\n", prompt, err)
				}
				errored++
				mu.Unlock()
				return
			}

			if err := convert.SaveWebP(data, outputPath); err != nil {
				mu.Lock()
				if flagJSON {
					outputJSON(GenerateResult{
						Status: "error",
						Prompt: prompt,
						Model:  model,
						Hash:   hash,
						Error:  err.Error(),
					})
				} else {
					fmt.Printf("Error saving [%s]: %v\n", prompt, err)
				}
				errored++
				mu.Unlock()
				return
			}

			mu.Lock()
			c.Upsert(prompt, model, filename)
			if flagJSON {
				outputJSON(GenerateResult{
					Status:     "generated",
					Prompt:     prompt,
					Model:      model,
					Hash:       hash,
					OutputFile: outputPath,
					Cached:     false,
				})
			} else if shouldOutput() {
				fmt.Printf("Generated: %s -> %s\n", prompt, filename)
			}
			mu.Unlock()
		}(p.Prompt, p.Model)
	}

	wg.Wait()

	// Save cache
	if err := c.Save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	if errored > 0 {
		msg := fmt.Sprintf("%d generation(s) failed", errored)
		if errored == len(toGenerate) {
			return &ExitError{Code: ExitTotalFail, Message: msg}
		}
		return &ExitError{Code: ExitPartialFail, Message: msg}
	}

	if shouldOutput() {
		fmt.Printf("\nDone. Generated %d images.\n", len(toGenerate))
	}
	return nil
}

// ValidationResult represents the JSON output for validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []string          `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
	Summary  ValidationSummary `json:"summary"`
}

// ValidationSummary provides counts for validation.
type ValidationSummary struct {
	TotalPrompts   int `json:"total_prompts"`
	UniquePrompts  int `json:"unique_prompts"`
	Duplicates     int `json:"duplicates"`
	EmptyPrompts   int `json:"empty_prompts"`
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Read file
	data, err := os.ReadFile(args[0])
	if err != nil {
		if flagJSON {
			outputJSON(ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("failed to read file: %v", err)},
			})
			return nil
		}
		return &ExitError{Code: ExitInvalidInput, Message: fmt.Sprintf("failed to read file: %v", err)}
	}

	// Parse YAML
	var pf PromptFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		if flagJSON {
			outputJSON(ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("invalid YAML syntax: %v", err)},
			})
			return nil
		}
		return &ExitError{Code: ExitInvalidInput, Message: fmt.Sprintf("invalid YAML syntax: %v", err)}
	}

	var (
		errors   []string
		warnings []string
		seen     = make(map[string]int)
		empty    int
	)

	// Check for empty prompts array
	if len(pf.Prompts) == 0 {
		errors = append(errors, "no prompts found in file")
	}

	// Validate each prompt
	for i, p := range pf.Prompts {
		model := p.Model
		if model == "" {
			model = flagModel
		}

		// Check for empty prompt
		if p.Prompt == "" {
			errors = append(errors, fmt.Sprintf("prompt %d: empty prompt text", i+1))
			empty++
			continue
		}

		// Check for duplicates
		key := p.Prompt + "|" + model
		if prev, exists := seen[key]; exists {
			warnings = append(warnings, fmt.Sprintf("prompt %d: duplicate of prompt %d (same prompt+model)", i+1, prev))
		} else {
			seen[key] = i + 1
		}
	}

	result := ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Summary: ValidationSummary{
			TotalPrompts:  len(pf.Prompts),
			UniquePrompts: len(seen),
			Duplicates:    len(pf.Prompts) - len(seen) - empty,
			EmptyPrompts:  empty,
		},
	}

	if flagJSON {
		outputJSON(result)
		if !result.Valid {
			return &ExitError{Code: ExitInvalidInput}
		}
		return nil
	}

	// Human-readable output
	if shouldOutput() {
		if result.Valid {
			fmt.Println("✓ Valid")
		} else {
			fmt.Println("✗ Invalid")
		}

		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Total prompts:  %d\n", result.Summary.TotalPrompts)
		fmt.Printf("  Unique prompts: %d\n", result.Summary.UniquePrompts)
		if result.Summary.Duplicates > 0 {
			fmt.Printf("  Duplicates:     %d\n", result.Summary.Duplicates)
		}
		if result.Summary.EmptyPrompts > 0 {
			fmt.Printf("  Empty prompts:  %d\n", result.Summary.EmptyPrompts)
		}

		if len(errors) > 0 {
			fmt.Printf("\nErrors:\n")
			for _, e := range errors {
				fmt.Printf("  • %s\n", e)
			}
		}

		if len(warnings) > 0 {
			fmt.Printf("\nWarnings:\n")
			for _, w := range warnings {
				fmt.Printf("  • %s\n", w)
			}
		}
	}

	if !result.Valid {
		return &ExitError{Code: ExitInvalidInput}
	}
	return nil
}
