package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kevinmichaelchen/replicate-images/internal/cache"
	"github.com/kevinmichaelchen/replicate-images/internal/client"
	"github.com/kevinmichaelchen/replicate-images/internal/convert"
	"github.com/spf13/cobra"
)

var (
	flagModel   string
	flagOutput  string
	flagNoCache bool
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

func init() {
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", client.DefaultModel, "Replicate model to use")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "./generated-images", "Output directory")
	rootCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "Force regeneration, ignore cache")

	rootCmd.AddCommand(modelsCmd)
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
	c.Add(prompt, flagModel, filename)
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
