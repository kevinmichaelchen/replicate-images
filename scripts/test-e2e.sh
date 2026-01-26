#!/bin/bash
# End-to-end test script. Run manually when you want to verify things work.
# Costs: 1 image generation (only if not cached)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BIN="$ROOT_DIR/replicate-images"
FIXTURES="$ROOT_DIR/test/fixtures"
TEST_PROMPT="e2e test fixture image"

cd "$ROOT_DIR"

echo "=== Building ==="
go build -o "$BIN" ./cmd/replicate-images

echo ""
echo "=== Test: validate (valid YAML) ==="
cat > "$FIXTURES/valid.yaml" << 'EOF'
prompts:
  - prompt: "test prompt one"
  - prompt: "test prompt two"
EOF
"$BIN" validate "$FIXTURES/valid.yaml"
echo "✓ Passed"

echo ""
echo "=== Test: validate (invalid YAML) ==="
cat > "$FIXTURES/invalid.yaml" << 'EOF'
prompts:
  - prompt: ""
  - prompt: "duplicate"
  - prompt: "duplicate"
EOF
if "$BIN" validate "$FIXTURES/invalid.yaml" 2>/dev/null; then
  echo "✗ Should have failed" && exit 1
fi
echo "✓ Passed (correctly rejected)"

echo ""
echo "=== Test: dry-run ==="
"$BIN" --dry-run "$TEST_PROMPT" > /dev/null
echo "✓ Passed"

echo ""
echo "=== Test: generate image (may use cache) ==="
OUTPUT=$("$BIN" --json -o "$FIXTURES" "$TEST_PROMPT")
echo "$OUTPUT" | head -1
if echo "$OUTPUT" | grep -q '"cached":true'; then
  echo "✓ Passed (from cache - \$0 spent)"
else
  echo "✓ Passed (generated new - \$\$ spent)"
fi

echo ""
echo "=== Test: cache hit ==="
OUTPUT=$("$BIN" --json -o "$FIXTURES" "$TEST_PROMPT")
if echo "$OUTPUT" | grep -q '"cached":true'; then
  echo "✓ Passed"
else
  echo "✗ Cache miss when hit expected" && exit 1
fi

echo ""
echo "=== All tests passed ==="
