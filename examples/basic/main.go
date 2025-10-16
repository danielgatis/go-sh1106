package main

import (
	"image"
	"log"
	"os"
	"time"

	"github.com/danielgatis/go-sh1106/pkg/display"
	"github.com/danielgatis/go-sh1106/pkg/text"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

func main() {
	// Get message from command line argument
	message := ""
	if len(os.Args) > 1 {
		message = os.Args[1]
	}

	// Load all the drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a handle to the first available SPI bus
	bus, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}

	dc := gpioreg.ByName("GPIO24")
	if dc == nil {
		log.Fatal("GPIO24 not available")
	}

	rst := gpioreg.ByName("GPIO25")
	if rst == nil {
		log.Fatal("GPIO25 not available")
	}

	cs := gpioreg.ByName("GPIO8")
	if cs == nil {
		log.Fatal("GPIO8 not available")
	}

	// Create SH1106 display driver
	dev, err := display.NewSH1106SPI(bus, dc, rst, cs, &display.Options{
		Width:  128,
		Height: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Halt()

	// Create text renderer with embedded font
	textRenderer, err := text.NewRendererWithEmbeddedFont(&text.Config{
		Width:     128,
		Height:    64,
		LineCount: 6,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set some sample text
	textRenderer.SetTexts([]string{
		message,
		message,
		message,
		message,
		message,
		message,
	})

	// Draw the text to the display
	if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
		log.Fatal(err)
	}

	// Keep the display on for 20 seconds
	time.Sleep(30 * time.Second)

	// Turn off the display
	dev.Halt()
}
