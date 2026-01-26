package main

import (
	"context"
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

var (
	flagModel       string
	flagOutput      string
	flagNoCache     bool
	flagConcurrency int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "replicate-images [prompt]",
	Short: "Generate images from text prompts using Replicate",
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "./generated-images", "Output directory")
	rootCmd.PersistentFlags().BoolVar(&flagNoCache, "no-cache", false, "Force regeneration, ignore cache")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Replicate model to use")

	batchCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Default model for prompts without one")
	batchCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 3, "Number of concurrent generations")

	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(batchCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	prompt := args[0]

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
	hash := cache.Hash(prompt, flagModel)
	if !flagNoCache {
		if entry := c.Lookup(hash); entry != nil {
			outputPath := filepath.Join(flagOutput, entry.OutputFile)
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("Using cached image: %s\n", outputPath)
				return nil
			}
		}
	}

	// Create client
	rc, err := client.New()
	if err != nil {
		return err
	}

	fmt.Printf("Generating image with %s...\n", flagModel)

	// Generate image
	data, url, err := rc.GenerateImage(ctx, flagModel, prompt)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded from: %s\n", url)

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

	fmt.Printf("Saved: %s\n", outputPath)
	return nil
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
		return fmt.Errorf("failed to read file: %w", err)
	}

	var pf PromptFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(pf.Prompts) == 0 {
		return fmt.Errorf("no prompts found in file")
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

	// Create client
	rc, err := client.New()
	if err != nil {
		return err
	}

	// Filter to only prompts that need generation
	var toGenerate []PromptEntry
	for _, p := range pf.Prompts {
		model := p.Model
		if model == "" {
			model = flagModel
		}

		hash := cache.Hash(p.Prompt, model)
		if !flagNoCache {
			if entry := c.Lookup(hash); entry != nil {
				outputPath := filepath.Join(flagOutput, entry.OutputFile)
				if _, err := os.Stat(outputPath); err == nil {
					fmt.Printf("Cached: %s\n", p.Prompt)
					continue
				}
			}
		}
		toGenerate = append(toGenerate, PromptEntry{Prompt: p.Prompt, Model: model})
	}

	if len(toGenerate) == 0 {
		fmt.Println("All images already cached.")
		return nil
	}

	fmt.Printf("Generating %d images (concurrency: %d)...\n\n", len(toGenerate), flagConcurrency)

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
				fmt.Printf("Error [%s]: %v\n", prompt, err)
				mu.Lock()
				errored++
				mu.Unlock()
				return
			}

			if err := convert.SaveWebP(data, outputPath); err != nil {
				fmt.Printf("Error saving [%s]: %v\n", prompt, err)
				mu.Lock()
				errored++
				mu.Unlock()
				return
			}

			mu.Lock()
			c.Upsert(prompt, model, filename)
			mu.Unlock()

			fmt.Printf("Generated: %s -> %s\n", prompt, filename)
		}(p.Prompt, p.Model)
	}

	wg.Wait()

	// Save cache
	if err := c.Save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	if errored > 0 {
		return fmt.Errorf("%d generation(s) failed", errored)
	}

	fmt.Printf("\nDone. Generated %d images.\n", len(toGenerate))
	return nil
}
