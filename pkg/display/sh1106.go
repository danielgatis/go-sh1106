// Package display provides drivers for various OLED displays.
package display

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/display"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

var _ display.Drawer = (*SH1106)(nil)

// SH1106 driver for OLED displays
type SH1106 struct {
	c   conn.Conn
	dc  gpio.PinOut
	rst gpio.PinOut
	cs  gpio.PinOut

	rect   image.Rectangle
	buffer []byte
}

// Options defines the configuration options for the SH1106 device
type Options struct {
	Width  int
	Height int
}

// NewSH1106SPI creates a new SH1106 display driver for SPI communication
func NewSH1106SPI(p spi.Port, dc, rst, cs gpio.PinOut, opts *Options) (*SH1106, error) {
	if dc == nil || rst == nil || cs == nil {
		return nil, errors.New("display: dc, rst, and cs pins are required")
	}

	speed := physic.Frequency(1.95 * float64(physic.MegaHertz)) // 1.95MHz
	c, err := p.Connect(speed, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}

	sh1106 := &SH1106{
		c:      c,
		dc:     dc,
		rst:    rst,
		cs:     cs,
		rect:   image.Rect(0, 0, opts.Width, opts.Height),
		buffer: make([]byte, (opts.Width*opts.Height)/8),
	}

	// Initialize display
	if err := sh1106.init(); err != nil {
		return nil, err
	}

	return sh1106, nil
}

// init initializes the SH1106 display with the proper command sequence
func (d *SH1106) init() error {
	// Hardware reset sequence
	d.rst.Out(gpio.High)
	time.Sleep(1 * time.Millisecond)
	d.rst.Out(gpio.Low)
	time.Sleep(1 * time.Millisecond)
	d.rst.Out(gpio.High)
	time.Sleep(1 * time.Millisecond)

	// Send initialization commands
	commands := []byte{
		0xAE, // Turn off OLED panel
		0x02, // Set low column address
		0x10, // Set high column address
		0x40, // Set start line address
		0x81, // Set contrast control register
		0xA0, // Set SEG/Column mapping
		0xC0, // Set COM/Row scan direction
		0xA6, // Set normal display
		0xA8, // Set multiplex ratio (1 to 64)
		0x3F, // 1/64 duty
		0xD3, // Set display offset
		0x00, // Not offset
		0xD5, // Set display clock divide ratio/oscillator frequency
		0x80, // Set divide ratio, Set Clock as 100 Frames/Sec
		0xD9, // Set pre-charge period
		0xF1, // Set Pre-Charge as 15 Clocks & Discharge as 1 Clock
		0xDA, // Set com pins hardware configuration
		0x12,
		0xDB, // Set vcomh
		0x40, // Set VCOM Deselect Level
		0x20, // Set Page Addressing Mode
		0x02,
		0xA4, // Disable Entire Display On
		0xA6, // Disable Inverse Display On
		0xAF, // Turn on OLED panel
	}

	for _, cmd := range commands {
		if err := d.sendCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}

// sendCommand sends a command to the display
func (d *SH1106) sendCommand(cmd byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	if err := d.cs.Out(gpio.Low); err != nil {
		return err
	}
	defer d.cs.Out(gpio.High)

	return d.c.Tx([]byte{cmd}, nil)
}

// sendData sends data to the display
func (d *SH1106) sendData(data []byte) error {
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	if err := d.cs.Out(gpio.Low); err != nil {
		return err
	}
	defer d.cs.Out(gpio.High)

	return d.c.Tx(data, nil)
}

// display sends the buffer to the display
func (d *SH1106) display() error {
	w := d.rect.Dx()
	h := d.rect.Dy()
	pages := h / 8

	for page := range pages {
		// Set page address
		if err := d.sendCommand(0xB0 + byte(page)); err != nil {
			return err
		}
		// Set low column address
		if err := d.sendCommand(0x02); err != nil {
			return err
		}
		// Set high column address
		if err := d.sendCommand(0x10); err != nil {
			return err
		}

		// Send page data
		start := page * w
		pageData := make([]byte, w)
		for i := range w {
			pageData[i] = ^d.buffer[start+i] // Invert for SH1106
		}
		if err := d.sendData(pageData); err != nil {
			return err
		}
	}

	return nil
}

// setPixel sets a pixel in the buffer
func (d *SH1106) setPixel(x, y int, on bool) {
	if x < 0 || x >= d.rect.Dx() || y < 0 || y >= d.rect.Dy() {
		return
	}

	page := y / 8
	bit := y % 8
	index := page*d.rect.Dx() + x

	if on {
		d.buffer[index] |= (1 << bit)
	} else {
		d.buffer[index] &^= (1 << bit)
	}
}

// Clear clears the display buffer
func (d *SH1106) Clear() {
	for i := range d.buffer {
		d.buffer[i] = 0xFF
	}
}

// String implements fmt.Stringer
func (d *SH1106) String() string {
	return fmt.Sprintf("SH1106{%s, %s}", d.rect.Max, d.c)
}

// Bounds returns the display bounds
func (d *SH1106) Bounds() image.Rectangle {
	return d.rect
}

// ColorModel returns the color model (monochrome)
func (d *SH1106) ColorModel() color.Model {
	return color.GrayModel
}

// Halt turns off the display
func (d *SH1106) Halt() error {
	return d.sendCommand(0xAE)
}

// Draw implements display.Drawer
func (d *SH1106) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	bounds := src.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			srcX := x - bounds.Min.X + sp.X
			srcY := y - bounds.Min.Y + sp.Y

			if srcX >= 0 && srcX < d.rect.Dx() && srcY >= 0 && srcY < d.rect.Dy() {
				// Convert to grayscale
				r, g, b, _ := src.At(x, y).RGBA()
				gray := uint8((r + g + b) / 3 >> 8)

				// Set pixel based on threshold
				// Pixels darker than 50% are on
				on := gray < 128
				d.setPixel(srcX, srcY, on)
			}
		}
	}

	return d.display()
}

// SetPixel directly sets a pixel on the display
func (d *SH1106) SetPixel(x, y int, on bool) {
	d.setPixel(x, y, on)
}

// Update displays the current buffer to the screen
func (d *SH1106) Update() error {
	return d.display()
}
