package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/replicate/replicate-go"
)

const DefaultModel = "black-forest-labs/flux-schnell"

type Client struct {
	r *replicate.Client
}

// New creates a new Replicate client using REPLICATE_API_TOKEN from environment.
func New() (*Client, error) {
	r, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		return nil, fmt.Errorf("failed to create replicate client: %w", err)
	}
	return &Client{r: r}, nil
}

// GenerateImage runs a text-to-image model and returns the image data.
func (c *Client) GenerateImage(ctx context.Context, model, prompt string) ([]byte, string, error) {
	input := replicate.PredictionInput{
		"prompt": prompt,
	}

	output, err := c.r.Run(ctx, model, input, nil)
	if err != nil {
		return nil, "", fmt.Errorf("prediction failed: %w", err)
	}

	// Output is typically a slice containing image URL(s)
	outputs, ok := output.([]any)
	if !ok || len(outputs) == 0 {
		return nil, "", fmt.Errorf("unexpected output format: %T", output)
	}

	imageURL, ok := outputs[0].(string)
	if !ok {
		return nil, "", fmt.Errorf("expected string URL, got: %T", outputs[0])
	}

	// Download the image
	data, err := downloadImage(ctx, imageURL)
	if err != nil {
		return nil, "", err
	}

	return data, imageURL, nil
}

// SearchModels searches for models by query and returns them sorted by popularity.
func (c *Client) SearchModels(ctx context.Context, query string) ([]ModelInfo, error) {
	page, err := c.r.SearchModels(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var models []ModelInfo
	for _, m := range page.Results {
		models = append(models, ModelInfo{
			Owner:       m.Owner,
			Name:        m.Name,
			Description: m.Description,
			RunCount:    m.RunCount,
		})
	}
	return models, nil
}

type ModelInfo struct {
	Owner       string
	Name        string
	Description string
	RunCount    int
}

func (m ModelInfo) FullName() string {
	return fmt.Sprintf("%s/%s", m.Owner, m.Name)
}

func downloadImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
