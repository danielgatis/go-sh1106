package main

import (
	"fmt"
	"image"
	"log"
	"time"

	"github.com/danielgatis/go-sh1106/pkg/display"
	"github.com/danielgatis/go-sh1106/pkg/text"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

func main() {
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

	// Animation loop
	for i := 0; i < 100; i++ {
		// Clear the display
		dev.Clear()

		// Create animated text
		textRenderer.SetTexts([]string{
			"SH1106 Display",
			fmt.Sprintf("Progress: %d%%", i),
			"Loading...",
			"",
			"Status: OK",
			fmt.Sprintf("Time: %02d:%02d", i/60, i%60),
		})

		// Draw the text to the display
		if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
			log.Printf("Error drawing: %v", err)
		}

		dev.Update()

		time.Sleep(200 * time.Millisecond)
	}

	// Show final message
	textRenderer.SetTexts([]string{
		"Animation",
		"Complete!",
		"",
		"SH1106 Display",
		"Library Demo",
		"Finished",
	})

	dev.Clear()
	if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
		log.Printf("Error drawing final message: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Turn off the display
	dev.Halt()
}
