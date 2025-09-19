package display

import (
	"image"
	"testing"
)

func TestSH1106Options(t *testing.T) {
	opts := &Options{
		Width:  128,
		Height: 64,
	}

	if opts.Width != 128 {
		t.Errorf("Expected width 128, got %d", opts.Width)
	}

	if opts.Height != 64 {
		t.Errorf("Expected height 64, got %d", opts.Height)
	}
}

func TestSH1106Bounds(t *testing.T) {
	// Create a mock SH1106 with specific dimensions
	sh1106 := &SH1106{
		rect: image.Rect(0, 0, 128, 64),
	}

	bounds := sh1106.Bounds()
	expected := image.Rect(0, 0, 128, 64)

	if bounds != expected {
		t.Errorf("Expected bounds %v, got %v", expected, bounds)
	}
}

func TestSH1106String(t *testing.T) {
	sh1106 := &SH1106{
		rect: image.Rect(0, 0, 128, 64),
	}

	str := sh1106.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}
}
