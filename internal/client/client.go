package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/kevinmichaelchen/replicate-images/internal/models"
	"github.com/replicate/replicate-go"
)

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
func (c *Client) GenerateImage(ctx context.Context, modelID, prompt string) ([]byte, string, error) {
	input := replicate.PredictionInput{
		"prompt": prompt,
	}

	// Apply model-specific defaults
	if model, ok := models.Get(modelID); ok {
		for k, v := range model.Defaults {
			input[k] = v
		}
	}

	output, err := c.r.Run(ctx, modelID, input, nil)
	if err != nil {
		return nil, "", fmt.Errorf("prediction failed: %w", err)
	}

	// Extract image URL from output - format varies by model
	imageURL, err := extractImageURL(output)
	if err != nil {
		return nil, "", err
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

	var results []ModelInfo
	for _, m := range page.Results {
		results = append(results, ModelInfo{
			Owner:       m.Owner,
			Name:        m.Name,
			Description: m.Description,
			RunCount:    m.RunCount,
		})
	}
	return results, nil
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

// extractImageURL handles various output formats from different Replicate models.
// Known formats:
//   - string: direct URL (e.g., "https://...")
//   - []any: array of URLs, take first (e.g., ["https://..."])
//   - map[string]any: object with URL field (e.g., {"url": "https://..."})
func extractImageURL(output any) (string, error) {
	switch v := output.(type) {
	case string:
		return v, nil
	case []any:
		if len(v) == 0 {
			return "", fmt.Errorf("empty output array from model")
		}
		// First element could be string or map
		return extractImageURL(v[0])
	case map[string]any:
		// Try common field names
		for _, key := range []string{"url", "image", "output", "uri"} {
			if val, ok := v[key]; ok {
				if s, ok := val.(string); ok {
					return s, nil
				}
			}
		}
		return "", fmt.Errorf("no image URL found in output object: %v", v)
	default:
		return "", fmt.Errorf("unexpected output format %T: %v", output, output)
	}
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
