# replicate-images

A CLI that generates images from text prompts using [Replicate].

## Features

- Text-to-image generation via [Replicate] API
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
```

## Configuration

| Flag             | Default                          | Description        |
| ---------------- | -------------------------------- | ------------------ |
| `--model`, `-m`  | `black-forest-labs/flux-schnell` | Model to use       |
| `--output`, `-o` | `./generated-images`             | Output directory   |
| `--no-cache`     | `false`                          | Force regeneration |

[Replicate]: https://replicate.com
[nativewebp]: https://github.com/HugoSmits86/nativewebp
