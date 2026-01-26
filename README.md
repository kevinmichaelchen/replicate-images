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

![Architecture Diagram](diagram.webp)

## Installation

### Script (macOS/Linux/FreeBSD)

```bash
curl -fsSL https://raw.githubusercontent.com/kevinmichaelchen/replicate-images/main/scripts/install.sh | bash
```

### Go

Requires Go 1.21+:

```bash
go install github.com/kevinmichaelchen/replicate-images/cmd/replicate-images@latest
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

## Supported Models

| Model                            | Best For                                          |
| -------------------------------- | ------------------------------------------------- |
| `black-forest-labs/flux-schnell` | Fast, high-quality generations (default)          |
| `black-forest-labs/flux-1.1-pro` | Higher quality, slower. Best for final outputs    |
| `stability-ai/sdxl`              | Classic model with wide style range               |
| `google/nano-banana-pro`         | Text rendering, diagrams, technical illustrations |

Other models may work but are untested. List supported models:

```bash
replicate-images supported-models
replicate-images supported-models --json  # For agents
```

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

### Discover Supported Models

Query available models programmatically:

```bash
replicate-images supported-models --json
[{"id":"black-forest-labs/flux-schnell","name":"FLUX Schnell","description":"Fast, high-quality generations. Great default choice.","default":true},...]
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

Releases are fully automated via
[release-please](https://github.com/googleapis/release-please):

1. Push commits to `main` using
   [conventional commits](https://conventionalcommits.org/)
2. release-please creates/updates a Release PR with changelog
3. Merge the PR when ready to release
4. GitHub Actions automatically builds binaries and publishes the release

**Commit prefixes and version bumps:**

| Prefix                         | Version Bump | Example                             |
| ------------------------------ | ------------ | ----------------------------------- |
| `fix:`                         | Patch        | `fix: handle empty response`        |
| `feat:`                        | Minor        | `feat: add retry logic`             |
| `feat!:` or `BREAKING CHANGE:` | Major        | `feat!: change API response format` |

[Replicate]: https://replicate.com
[nativewebp]: https://github.com/HugoSmits86/nativewebp
