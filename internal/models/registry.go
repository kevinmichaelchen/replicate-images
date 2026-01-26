// Package models defines supported image generation models and their configurations.
package models

// Model represents a supported image generation model.
type Model struct {
	ID          string         // e.g., "black-forest-labs/flux-schnell"
	Name        string         // Human-friendly name
	Description string         // What the model is good at
	Defaults    map[string]any // Default input parameters beyond prompt
}

// Supported models registry.
var Supported = []Model{
	{
		ID:          "black-forest-labs/flux-schnell",
		Name:        "FLUX Schnell",
		Description: "Fast, high-quality generations. Great default choice.",
		Defaults:    nil,
	},
	{
		ID:          "black-forest-labs/flux-1.1-pro",
		Name:        "FLUX 1.1 Pro",
		Description: "Higher quality than Schnell, slower. Best for final outputs.",
		Defaults:    nil,
	},
	{
		ID:          "stability-ai/sdxl",
		Name:        "Stable Diffusion XL",
		Description: "Classic model with wide style range and community support.",
		Defaults:    nil,
	},
	{
		ID:          "google/nano-banana-pro",
		Name:        "Nano Banana Pro",
		Description: "Excellent for text rendering, diagrams, and technical illustrations.",
		Defaults: map[string]any{
			"aspect_ratio": "1:1",
		},
	},
}

// Default is the model used when none is specified.
const Default = "black-forest-labs/flux-schnell"

// registry provides O(1) lookup by ID.
var registry = make(map[string]Model)

func init() {
	for _, m := range Supported {
		registry[m.ID] = m
	}
}

// Get returns a model by ID and whether it's supported.
func Get(id string) (Model, bool) {
	m, ok := registry[id]
	return m, ok
}

// IsSupported returns true if the model is in the supported list.
func IsSupported(id string) bool {
	_, ok := registry[id]
	return ok
}

// List returns all supported model IDs.
func List() []string {
	ids := make([]string, len(Supported))
	for i, m := range Supported {
		ids[i] = m.ID
	}
	return ids
}
