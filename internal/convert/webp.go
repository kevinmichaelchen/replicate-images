package convert

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/HugoSmits86/nativewebp"
)

// ToWebP converts image data to WEBP format if needed.
// Returns the converted data and true if conversion occurred.
func ToWebP(data []byte) ([]byte, bool, error) {
	contentType := http.DetectContentType(data)

	// Already WEBP, no conversion needed
	if strings.Contains(contentType, "webp") {
		return data, false, nil
	}

	// Decode the source image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode image: %w", err)
	}

	// Encode to WEBP
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, nil); err != nil {
		return nil, false, fmt.Errorf("failed to encode webp: %w", err)
	}

	return buf.Bytes(), true, nil
}

// SaveWebP saves image data as WEBP to the specified path.
// Converts if necessary.
func SaveWebP(data []byte, path string) error {
	converted, _, err := ToWebP(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, converted, 0644)
}

// IsWebP checks if the data is already in WEBP format.
func IsWebP(r io.Reader) bool {
	buf := make([]byte, 12)
	n, err := r.Read(buf)
	if err != nil || n < 12 {
		return false
	}
	// WEBP files start with "RIFF" and contain "WEBP"
	return string(buf[0:4]) == "RIFF" && string(buf[8:12]) == "WEBP"
}
