package httpapi

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

const thumbnailMaxSize = 300

func generateThumbnail(srcPath string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(srcPath))
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	default:
		return "", nil // not a supported image format, skip
	}
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= thumbnailMaxSize && h <= thumbnailMaxSize {
		return "", nil // already small enough
	}

	ratio := float64(thumbnailMaxSize) / float64(max(w, h))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	thumb := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.ApproxBiLinear.Scale(thumb, thumb.Bounds(), img, bounds, draw.Over, nil)

	thumbPath := strings.TrimSuffix(srcPath, ext) + "_thumb" + ext
	var buf bytes.Buffer
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80})
	case ".png":
		err = png.Encode(&buf, thumb)
	}
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(thumbPath, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return thumbPath, nil
}
