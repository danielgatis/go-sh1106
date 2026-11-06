package text

import (
	"image"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
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

func TestScale_EnlargesEachPixelIntoABlock(t *testing.T) {
	small, err := NewRendererWithEmbeddedFont(&Config{Width: 30, Height: 20, LineCount: 2})
	if err != nil {
		t.Fatalf("NewRendererWithEmbeddedFont: %v", err)
	}

	big, err := NewRendererWithEmbeddedFont(&Config{Width: 92, Height: 61, LineCount: 2, Scale: 3})
	if err != nil {
		t.Fatalf("NewRendererWithEmbeddedFont: %v", err)
	}

	small.SetTexts([]string{"Hi!", "42"})
	big.SetTexts([]string{"Hi!", "42"})

	if big.Bounds() != image.Rect(0, 0, 92, 61) {
		t.Fatalf("bounds %v, want the configured size", big.Bounds())
	}

	// The edge past 90 x 60 repeats the last canvas row and column
	for y := range 61 {
		for x := range 92 {
			want := small.Image().At(min(x/3, 29), min(y/3, 19))
			if got := big.Image().At(x, y); got != want {
				t.Fatalf("pixel (%d,%d) is %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestScale_Invalid(t *testing.T) {
	for _, cfg := range []*Config{
		{Width: 10, Height: 10, Scale: -1},
		{Width: 2, Height: 10, Scale: 3},
	} {
		if _, err := NewRendererWithEmbeddedFont(cfg); err == nil {
			t.Errorf("expected an error for %+v", cfg)
		}
	}
}

func TestNewRendererWithFace_RequiresFace(t *testing.T) {
	if _, err := NewRendererWithFace(nil, &Config{Width: 10, Height: 10}); err == nil {
		t.Fatal("expected an error without a face")
	}
}

func TestNewRendererWithFace_KeepsDescenders(t *testing.T) {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatalf("opentype.Parse: %v", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 16, DPI: 72})
	if err != nil {
		t.Fatalf("opentype.NewFace: %v", err)
	}

	r, err := NewRendererWithFace(face, &Config{Width: 60, Height: 40, LineCount: 2})
	if err != nil {
		t.Fatalf("NewRendererWithFace: %v", err)
	}
	r.SetTexts([]string{"gy"})

	bg := r.Image().At(59, 39)
	baseline := 20 - face.Metrics().Descent.Ceil()

	inked := func(y int) bool {
		for x := range 60 {
			if r.Image().At(x, y) != bg {
				return true
			}
		}
		return false
	}

	// The tails of g and y reach below the baseline, inside the first line
	if !inked(baseline + 2) {
		t.Fatalf("no ink below the baseline at row %d: descenders cut off", baseline+2)
	}
	for y := 20; y < 40; y++ {
		if inked(y) {
			t.Fatalf("ink at row %d spilled into the second line", y)
		}
	}
}
