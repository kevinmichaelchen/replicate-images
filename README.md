# replicate-images

[![Go Report Card](https://goreportcard.com/badge/github.com/kevinmichaelchen/replicate-images)](https://goreportcard.com/report/github.com/kevinmichaelchen/replicate-images)

A CLI that generates images from text prompts using [Replicate].

## Features

- Text-to-image generation via [Replicate] API
- Batch processing from YAML files
- Caching based on prompt+model hash (avoids duplicate generations)
- Automatic WEBP conversion using [nativewebp]
- Model search by popularity
- Agent-friendly: JSON output, dry-run, structured exit codes

## Installation

### Homebrew (macOS/Linux)

```bash
brew install kevinmichaelchen/tap/replicate-images
```

### Go

Requires Go 1.21+:

```bash
go install github.com/kevinmichaelchen/replicate-images/cmd/replicate-images@latest
```

### Download Binary

Download pre-built binaries from the
[GitHub Releases](https://github.com/kevinmichaelchen/replicate-images/releases)
page.

```bash
# macOS (Apple Silicon)
curl -Lo replicate-images https://github.com/kevinmichaelchen/replicate-images/releases/latest/download/replicate-images_darwin_arm64
chmod +x replicate-images
sudo mv replicate-images /usr/local/bin/

# macOS (Intel)
curl -Lo replicate-images https://github.com/kevinmichaelchen/replicate-images/releases/latest/download/replicate-images_darwin_amd64
chmod +x replicate-images
sudo mv replicate-images /usr/local/bin/

# Linux (amd64)
curl -Lo replicate-images https://github.com/kevinmichaelchen/replicate-images/releases/latest/download/replicate-images_linux_amd64
chmod +x replicate-images
sudo mv replicate-images /usr/local/bin/
```

## Usage

```bash
export REPLICATE_API_TOKEN="r8_..."

# Generate an image
replicate-images "a cat wearing a hat"

# Use a different model
replicate-images --model stability-ai/sdxl "a sunset over mountains"

# Custom output directory
replicate-images --output ./my-art "abstract painting"

# Skip cache
replicate-images --no-cache "a cat wearing a hat"

# Search for models
replicate-images models "anime"

# Batch process from YAML
replicate-images batch prompts.yaml

# Validate YAML before processing
replicate-images validate prompts.yaml
```

### Batch File Format

```yaml
prompts:
  - prompt: "a cat in space"
    model: black-forest-labs/flux-schnell
  - prompt: "a dog on the moon"
  - prompt: "a bird underwater"
    model: stability-ai/sdxl
```

Prompts without a `model` use the default or `--model` flag value.

## Configuration

| Flag                  | Default                          | Description                    |
| --------------------- | -------------------------------- | ------------------------------ |
| `--model`, `-m`       | `black-forest-labs/flux-schnell` | Model to use                   |
| `--output`, `-o`      | `./generated-images`             | Output directory               |
| `--no-cache`          | `false`                          | Force regeneration             |
| `--concurrency`, `-c` | `3`                              | Concurrent generations (batch) |
| `--json`              | `false`                          | Output as JSON/JSONL           |
| `--dry-run`           | `false`                          | Preview without generating     |
| `--quiet`, `-q`       | `false`                          | Suppress output, use exit code |

## Agent-Friendly Features

### JSON Output

```bash
# Single prompt
replicate-images --json "a cat in space"
{"status":"generated","prompt":"a cat in space","model":"black-forest-labs/flux-schnell","hash":"f417c5f0015e36af","output_file":"./generated-images/f417c5f0015e36af.webp","cached":false}

# Batch (JSONL - one JSON per line)
replicate-images batch --json prompts.yaml
```

### Dry Run

Preview what would be generated without making API calls:

```bash
replicate-images --dry-run --json "a cat in space"
{"to_generate":1,"cached":0,"prompts":[{"prompt":"a cat in space","model":"black-forest-labs/flux-schnell","hash":"f417c5f0015e36af","status":"pending"}]}
```

### Validation

Check YAML syntax and detect issues before processing:

```bash
replicate-images validate prompts.yaml
replicate-images validate --json prompts.yaml
```

### Exit Codes

| Code | Meaning                                |
| ---- | -------------------------------------- |
| 0    | Success (all generated or cached)      |
| 1    | Partial failure (some failed)          |
| 2    | Total failure (all failed)             |
| 3    | Invalid input (bad YAML, missing file) |

### Quiet Mode

Suppress all output; rely on exit codes:

```bash
replicate-images -q "a cat in space" && echo "Success" || echo "Failed"
```

## Development

### Prerequisites

- Go 1.21+
- [golangci-lint](https://golangci-lint.run/welcome/install/)
- [pre-commit](https://pre-commit.com/) (optional)

### Setup

```bash
# Clone the repo
git clone https://github.com/kevinmichaelchen/replicate-images.git
cd replicate-images

# Install pre-commit hooks (optional)
pre-commit install

# Build
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt
```

### Releasing

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

1. Create and push a tag:

   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically build and publish:
   - Binaries for macOS (arm64, amd64), Linux (amd64, arm64), and Windows
   - Homebrew formula to `kevinmichaelchen/homebrew-tap`

### Homebrew Tap Setup

To enable Homebrew installation, create a tap repository:

1. Create a new repo: `kevinmichaelchen/homebrew-tap`
2. GoReleaser will automatically push formula updates on each release

[Replicate]: https://replicate.com
[nativewebp]: https://github.com/HugoSmits86/nativewebp
