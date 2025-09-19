// Package text provides text rendering capabilities for OLED displays.
package text

import (
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/zachomedia/go-bdf"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Renderer handles text rendering on an image canvas
type Renderer struct {
	img        *image.RGBA
	face       font.Face
	width      int
	height     int
	lineCount  int
	lineHeight int
	lines      []string
}

// Config holds configuration for the text renderer
type Config struct {
	Width     int
	Height    int
	LineCount int
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

// NewRenderer creates a new text renderer with the specified configuration
func NewRenderer(fontPath string, config *Config) (*Renderer, error) {
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	return newRendererFromBytes(fontBytes, config)
}

// NewRendererWithEmbeddedFont creates a new text renderer using the embedded font
func NewRendererWithEmbeddedFont(config *Config) (*Renderer, error) {
	return newRendererFromBytes(embeddedFontData, config)
}

// newRendererFromBytes creates a new text renderer from font bytes
func newRendererFromBytes(fontBytes []byte, config *Config) (*Renderer, error) {
	// Validate configuration
	if config.Width <= 0 {
		return nil, &ConfigError{Field: "Width", Value: config.Width, Message: "width must be greater than 0"}
	}
	if config.Height <= 0 {
		return nil, &ConfigError{Field: "Height", Value: config.Height, Message: "height must be greater than 0"}
	}
	if config.LineCount < 0 {
		return nil, &ConfigError{Field: "LineCount", Value: config.LineCount, Message: "line count must be non-negative"}
	}

	bdfFont, err := bdf.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face := bdfFont.NewFace()

	rect := image.Rect(0, 0, config.Width, config.Height)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	lineHeight := 1
	if config.LineCount > 0 {
		lineHeight = config.Height / config.LineCount
	}

	return &Renderer{
		img:        img,
		face:       face,
		width:      config.Width,
		height:     config.Height,
		lineCount:  config.LineCount,
		lineHeight: lineHeight,
		lines:      make([]string, config.LineCount),
	}, nil
}

// SetText sets the text for a specific line
func (r *Renderer) SetText(text string, line int) {
	if line < 0 || line >= r.lineCount {
		return
	}
	r.lines[line] = text
	r.redraw()
}

// SetTexts sets multiple lines of text at once
func (r *Renderer) SetTexts(texts []string) {
	for i, text := range texts {
		if i < r.lineCount {
			r.lines[i] = text
		}
	}
	r.redraw()
}

// Clear clears all text from the renderer
func (r *Renderer) Clear() {
	for i := range r.lines {
		r.lines[i] = ""
	}
	r.redraw()
}

// redraw redraws the image with all lines
func (r *Renderer) redraw() {
	// Clear the canvas
	draw.Draw(r.img, r.img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  r.img,
		Src:  image.NewUniform(color.White),
		Face: r.face,
	}

	for row, line := range r.lines {
		if line == "" {
			continue
		}

		d.Dot = fixed.Point26_6{
			X: fixed.I(0),
			Y: fixed.I((row+1)*r.lineHeight - 1), // Adjust baseline
		}
		d.DrawString(line)
	}
}

// Image returns the rendered image
func (r *Renderer) Image() image.Image {
	return r.img
}

// Bounds returns the bounds of the rendered image
func (r *Renderer) Bounds() image.Rectangle {
	return r.img.Bounds()
}
