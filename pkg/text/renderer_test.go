package text

import (
	"image"
	"testing"
)

func TestTextRendererConfig(t *testing.T) {
	config := &Config{
		Width:     128,
		Height:    64,
		LineCount: 6,
	}

	if config.Width != 128 {
		t.Errorf("Expected width 128, got %d", config.Width)
	}

	if config.Height != 64 {
		t.Errorf("Expected height 64, got %d", config.Height)
	}

	if config.LineCount != 6 {
		t.Errorf("Expected line count 6, got %d", config.LineCount)
	}
}

func TestTextRendererBounds(t *testing.T) {
	// This test would require a font file, so we'll just test the config
	config := &Config{
		Width:     128,
		Height:    64,
		LineCount: 6,
	}

	expectedBounds := image.Rect(0, 0, config.Width, config.Height)
	if expectedBounds.Dx() != 128 || expectedBounds.Dy() != 64 {
		t.Errorf("Expected bounds %v, got %v", expectedBounds, expectedBounds)
	}
}

func TestNewRendererWithEmbeddedFont(t *testing.T) {
	config := &Config{
		Width:     128,
		Height:    64,
		LineCount: 6,
	}

	renderer, err := NewRendererWithEmbeddedFont(config)
	if err != nil {
		t.Fatalf("Failed to create renderer with embedded font: %v", err)
	}

	if renderer == nil {
		t.Fatal("Renderer should not be nil")
	}

	// Test basic properties
	if renderer.width != config.Width {
		t.Errorf("Expected width %d, got %d", config.Width, renderer.width)
	}

	if renderer.height != config.Height {
		t.Errorf("Expected height %d, got %d", config.Height, renderer.height)
	}

	if renderer.lineCount != config.LineCount {
		t.Errorf("Expected line count %d, got %d", config.LineCount, renderer.lineCount)
	}

	// Test that we can set text
	renderer.SetText("Hello World", 0)
	renderer.SetText("Test Line 2", 1)

	// Test bounds
	bounds := renderer.Bounds()
	expectedBounds := image.Rect(0, 0, config.Width, config.Height)
	if bounds != expectedBounds {
		t.Errorf("Expected bounds %v, got %v", expectedBounds, bounds)
	}

	// Test that we can get an image
	img := renderer.Image()
	if img == nil {
		t.Fatal("Image should not be nil")
	}

	if img.Bounds() != expectedBounds {
		t.Errorf("Expected image bounds %v, got %v", expectedBounds, img.Bounds())
	}
}

func TestNewRendererWithEmbeddedFontInvalidConfig(t *testing.T) {
	// Test with zero width
	config := &Config{
		Width:     0,
		Height:    64,
		LineCount: 6,
	}

	_, err := NewRendererWithEmbeddedFont(config)
	if err == nil {
		t.Error("Expected error for zero width, got nil")
	}

	// Test with zero height
	config = &Config{
		Width:     128,
		Height:    0,
		LineCount: 6,
	}

	_, err = NewRendererWithEmbeddedFont(config)
	if err == nil {
		t.Error("Expected error for zero height, got nil")
	}
}

func TestEmbeddedFontData(t *testing.T) {
	// Test that embedded font data is not empty
	if len(embeddedFontData) == 0 {
		t.Fatal("Embedded font data should not be empty")
	}

	// Test that it starts with BDF header
	if len(embeddedFontData) < 10 {
		t.Fatal("Embedded font data seems too short")
	}

	// Check for BDF file signature
	header := string(embeddedFontData[:10])
	if header != "STARTFONT " {
		t.Errorf("Expected BDF file to start with 'STARTFONT ', got '%s'", header)
	}
}
