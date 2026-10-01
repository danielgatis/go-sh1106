package display

import (
	"bytes"
	"image"
	"testing"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
	"periph.io/x/conn/v3/spi/spitest"
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

type levelPin struct {
	gpiotest.Pin
	levels []gpio.Level
}

func (p *levelPin) Out(l gpio.Level) error {
	p.levels = append(p.levels, l)

	return p.Pin.Out(l)
}

func newTestDisplay(t *testing.T, cs gpio.PinOut) (*SH1106, *spitest.Record) {
	t.Helper()

	port := &spitest.Record{}
	dc := &gpiotest.Pin{N: "DC"}
	rst := &gpiotest.Pin{N: "RST"}

	dev, err := NewSH1106SPI(port, dc, rst, cs, &Options{Width: 128, Height: 64})
	if err != nil {
		t.Fatalf("NewSH1106SPI: %v", err)
	}

	return dev, port
}

func writes(port *spitest.Record) [][]byte {
	out := make([][]byte, 0, len(port.Ops))
	for _, op := range port.Ops {
		out = append(out, op.W)
	}

	return out
}

func TestNewSH1106SPI_NilCSUsesHardwareChipSelect(t *testing.T) {
	_, port := newTestDisplay(t, nil)

	if len(port.Ops) == 0 {
		t.Fatal("expected init commands to be written")
	}
}

func TestNewSH1106SPI_RequiresDCAndRST(t *testing.T) {
	_, err := NewSH1106SPI(&spitest.Record{}, nil, &gpiotest.Pin{}, nil, &Options{Width: 128, Height: 64})
	if err == nil {
		t.Fatal("expected an error without dc")
	}
}

func TestNewSH1106SPI_SoftwareCSFramesEveryTransfer(t *testing.T) {
	cs := &levelPin{Pin: gpiotest.Pin{N: "CS"}}

	_, port := newTestDisplay(t, cs)

	if got, want := len(cs.levels), 2*len(port.Ops); got != want {
		t.Fatalf("cs toggled %d times, want %d", got, want)
	}

	for i := 0; i < len(cs.levels); i += 2 {
		if cs.levels[i] != gpio.Low || cs.levels[i+1] != gpio.High {
			t.Fatalf("transfer %d not framed low/high: %v", i/2, cs.levels[i:i+2])
		}
	}
}

func TestInit_ContrastCommandCarriesItsValue(t *testing.T) {
	_, port := newTestDisplay(t, nil)

	w := writes(port)
	for i, b := range w {
		if len(b) == 1 && b[0] == 0x81 {
			if i+1 >= len(w) || !bytes.Equal(w[i+1], []byte{0x80}) {
				t.Fatalf("0x81 not followed by its contrast value: %v", w[i+1:])
			}

			return
		}
	}

	t.Fatal("contrast command not sent")
}

func TestSetContrast(t *testing.T) {
	dev, port := newTestDisplay(t, nil)
	port.Ops = nil

	if err := dev.SetContrast(0x40); err != nil {
		t.Fatalf("SetContrast: %v", err)
	}

	if w := writes(port); len(w) != 1 || !bytes.Equal(w[0], []byte{0x81, 0x40}) {
		t.Fatalf("got %v, want [[0x81 0x40]]", w)
	}
}

func TestHaltAndWake(t *testing.T) {
	dev, port := newTestDisplay(t, nil)
	port.Ops = nil

	if err := dev.Halt(); err != nil {
		t.Fatalf("Halt: %v", err)
	}

	if err := dev.Wake(); err != nil {
		t.Fatalf("Wake: %v", err)
	}

	if w := writes(port); len(w) != 2 || w[0][0] != 0xAE || w[1][0] != 0xAF {
		t.Fatalf("got %v, want [[0xAE] [0xAF]]", w)
	}
}
