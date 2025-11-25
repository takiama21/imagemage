package filehandler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/image/draw"
)

// SaveImage saves base64 encoded image data to a file
func SaveImage(imageData, outputPath string) error {
	// Decode base64 image data
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fmt.Errorf("failed to decode image data: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write image to file
	if err := os.WriteFile(outputPath, decoded, 0644); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// GenerateFilename creates a descriptive filename from a prompt
func GenerateFilename(prompt, prefix string, count int) string {
	// Clean the prompt to make it filename-friendly
	cleaned := cleanPrompt(prompt)

	// Truncate if too long
	maxLen := 50
	if len(cleaned) > maxLen {
		cleaned = cleaned[:maxLen]
	}

	// Build filename
	var filename string
	if prefix != "" {
		filename = fmt.Sprintf("%s_%s", prefix, cleaned)
	} else {
		filename = cleaned
	}

	// Add counter if specified
	if count > 0 {
		filename = fmt.Sprintf("%s_%d", filename, count)
	}

	return filename + ".png"
}

// cleanPrompt converts a prompt into a filename-safe string
func cleanPrompt(prompt string) string {
	// Convert to lowercase
	s := strings.ToLower(prompt)

	// Remove special characters and replace spaces with underscores
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")

	// Replace multiple spaces with single underscore
	reg = regexp.MustCompile(`\s+`)
	s = reg.ReplaceAllString(s, "_")

	// Remove leading/trailing underscores
	s = strings.Trim(s, "_")

	return s
}

// EnsureUniqueFilename checks if a file exists and adds a counter if needed
func EnsureUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// File exists, add counter
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	counter := 1
	for {
		newPath := fmt.Sprintf("%s_%d%s", base, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// LoadImageAsBase64 loads an image file and returns it as base64
func LoadImageAsBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// ResizeAndSaveImage decodes base64 image data, resizes it to the target size, and saves to outputPath
func ResizeAndSaveImage(imageData string, size int, outputPath string) error {
	// Decode base64 image data
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fmt.Errorf("failed to decode image data: %w", err)
	}

	// Decode image
	src, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Create destination image with target size (square for icons)
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// Use high-quality CatmullRom interpolation for resizing
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Encode as PNG
	if err := png.Encode(f, dst); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
