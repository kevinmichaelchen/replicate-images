# replicate-images

A CLI that generates images from text prompts using [Replicate].

## Features

- Text-to-image generation via [Replicate] API
- Batch processing from YAML files
- Caching based on prompt+model hash (avoids duplicate generations)
- Automatic WEBP conversion using [nativewebp]
- Model search by popularity

## Installation

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

[Replicate]: https://replicate.com
[nativewebp]: https://github.com/HugoSmits86/nativewebp
