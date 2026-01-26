package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const CacheFileName = "cache.json"

type Entry struct {
	Hash       string    `json:"hash"`
	Prompt     string    `json:"prompt"`
	Model      string    `json:"model"`
	OutputFile string    `json:"output_file"`
	CreatedAt  time.Time `json:"created_at"`
}

type Cache struct {
	Entries []Entry `json:"entries"`
	path    string
}

// Load reads the cache from the output directory, creating it if it doesn't exist.
func Load(outputDir string) (*Cache, error) {
	path := filepath.Join(outputDir, CacheFileName)
	c := &Cache{
		Entries: []Entry{},
		path:    path,
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	c.path = path
	return c, nil
}

// Save writes the cache to disk.
func (c *Cache) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// Hash generates a unique hash for a prompt+model combination.
func Hash(prompt, model string) string {
	h := sha256.New()
	h.Write([]byte(prompt))
	h.Write([]byte(model))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Lookup finds an existing cache entry by hash.
func (c *Cache) Lookup(hash string) *Entry {
	for i := range c.Entries {
		if c.Entries[i].Hash == hash {
			return &c.Entries[i]
		}
	}
	return nil
}

// Add creates a new cache entry.
func (c *Cache) Add(prompt, model, outputFile string) *Entry {
	entry := Entry{
		Hash:       Hash(prompt, model),
		Prompt:     prompt,
		Model:      model,
		OutputFile: outputFile,
		CreatedAt:  time.Now(),
	}
	c.Entries = append(c.Entries, entry)
	return &entry
}
