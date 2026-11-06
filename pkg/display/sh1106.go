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

	halted bool
}

// Options defines the configuration options for the SH1106 device
type Options struct {
	// Width and Height are the panel's size: 128 x 64 for the Waveshare
	// 1.3inch OLED HAT. Height is a multiple of 8, the panel's page height.
	Width  int
	Height int
}

// NewSH1106SPI creates a new SH1106 display driver for SPI communication.
//
// cs may be nil when the SPI controller drives chip select in hardware (the
// spidev CE line); the driver then never touches it. Pass a pin only when chip
// select is wired to a plain GPIO.
func NewSH1106SPI(p spi.Port, dc, rst, cs gpio.PinOut, opts *Options) (*SH1106, error) {
	if dc == nil || rst == nil {
		return nil, errors.New("display: dc and rst pins are required")
	}

	if opts == nil || opts.Width <= 0 || opts.Height <= 0 {
		return nil, errors.New("display: width and height must be greater than 0")
	}

	if opts.Height%8 != 0 {
		return nil, errors.New("display: height must be a multiple of 8")
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
	sh1106.Clear()

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
		0x80, // Contrast value (power-on default)
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
func (d *SH1106) sendCommand(cmd ...byte) error {
	return d.tx(gpio.Low, cmd)
}

// sendData sends data to the display
func (d *SH1106) sendData(data []byte) error {
	return d.tx(gpio.High, data)
}

// tx sets DC to the given level and writes b, framing it with the software
// chip select when one was given.
func (d *SH1106) tx(dc gpio.Level, b []byte) error {
	if err := d.dc.Out(dc); err != nil {
		return err
	}

	if d.cs != nil {
		if err := d.cs.Out(gpio.Low); err != nil {
			return err
		}
		defer d.cs.Out(gpio.High)
	}

	return d.c.Tx(b, nil)
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

// setPixel sets a pixel in the buffer. A set bit is black: display inverts
// the buffer, and the SH1106 lights a pixel for a set bit.
func (d *SH1106) setPixel(x, y int, black bool) {
	if x < 0 || x >= d.rect.Dx() || y < 0 || y >= d.rect.Dy() {
		return
	}

	page := y / 8
	bit := y % 8
	index := page*d.rect.Dx() + x

	if black {
		d.buffer[index] |= (1 << bit)
	} else {
		d.buffer[index] &^= (1 << bit)
	}
}

// Clear clears the display buffer to black
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

// Halt turns off the display. The panel stops driving the OLED but keeps its
// RAM, so Wake shows the last frame again without a redraw. The next update
// wakes the panel; Wake does it ahead of time.
func (d *SH1106) Halt() error {
	if d.halted {
		return nil
	}

	if err := d.sendCommand(0xAE); err != nil {
		return err
	}
	d.halted = true

	return nil
}

// Wake turns the display back on after Halt.
func (d *SH1106) Wake() error {
	if !d.halted {
		return nil
	}

	if err := d.sendCommand(0xAF); err != nil {
		return err
	}
	d.halted = false

	return nil
}

// SetContrast sets the segment drive current, 0x00 to 0xFF (0x80 after init).
// Lower values dim the panel and slow OLED wear.
func (d *SH1106) SetContrast(level byte) error {
	return d.sendCommand(0x81, level)
}

// Draw implements display.Drawer: it draws src, from sp, into r of the
// display, then shows the buffer.
func (d *SH1106) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	r = r.Intersect(d.rect)

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p := sp.Add(image.Pt(x-r.Min.X, y-r.Min.Y))
			if !p.In(src.Bounds()) {
				continue
			}

			// Pixels darker than 50% are black
			gray := color.GrayModel.Convert(src.At(p.X, p.Y)).(color.Gray)
			d.setPixel(x, y, gray.Y < 128)
		}
	}

	return d.Update()
}

// SetPixel directly sets a pixel on the display buffer
func (d *SH1106) SetPixel(x, y int, black bool) {
	d.setPixel(x, y, black)
}

// Update displays the current buffer to the screen, waking the panel from
// Halt
func (d *SH1106) Update() error {
	if err := d.Wake(); err != nil {
		return err
	}

	return d.display()
}
