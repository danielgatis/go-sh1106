# SH1106 Go Driver

[![Go Report Card](https://goreportcard.com/badge/github.com/danielgatis/go-sh1106?style=flat-square)](https://goreportcard.com/report/github.com/danielgatis/go-sh1106)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/danielgatis/go-sh1106/master/LICENSE)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/danielgatis/go-sh1106)
[![Release](https://img.shields.io/github/release/danielgatis/go-sh1106.svg?style=flat-square)](https://github.com/danielgatis/go-sh1106/releases/latest)



A Go library for controlling SH1106 OLED displays with text rendering capabilities.

[<video src="demo.mp4" controls width="600"></video>](https://github.com/user-attachments/assets/6c657a5c-2fcd-41a2-b957-916cb8a23586)

https://www.waveshare.com/1.3inch-oled-hat.htm

## Features

- **SH1106 Display Driver**: Full support for SH1106 OLED displays via SPI
- **Text Rendering**: BDF font support with embedded font option
- **Joystick Support**: Complete joystick/button handling with callbacks
- **Easy Integration**: Simple API for quick integration
- **Examples**: Complete examples showing different usage patterns

## Installation

```bash
go get github.com/danielgatis/go-sh1106
```

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/danielgatis/go-sh1106/pkg/display"
    "github.com/danielgatis/go-sh1106/pkg/text"
    
    "periph.io/x/conn/v3/gpio/gpioreg"
    "periph.io/x/conn/v3/spi/spireg"
    "periph.io/x/host/v3"
)

func main() {
    // Initialize periph.io
    host.Init()
    
    // Open SPI bus
    bus, _ := spireg.Open("")
    
    // Configure GPIO pins
    dc := gpioreg.ByName("GPIO24")
    rst := gpioreg.ByName("GPIO25")
    cs := gpioreg.ByName("GPIO8")
    
    // Create display
    dev, _ := display.NewSH1106SPI(bus, dc, rst, cs, &display.Options{
        Width:  128,
        Height: 64,
    })
    defer dev.Halt()
    
    // Create text renderer with embedded font
    textRenderer, _ := text.NewRendererWithEmbeddedFont(&text.Config{
        Width:     128,
        Height:    64,
        LineCount: 6,
    })
    
    // Set text and display
    textRenderer.SetTexts([]string{
        "Hello World!",
        "SH1106 Display",
        "Go Library",
    })
    
    dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{})
}
```

### Enable SPI on Raspberry Pi

```bash
sudo raspi-config
# Navigate to: Interfacing Options > SPI > Enable
```

## Packages

### Display Package (`pkg/display`)
SH1106 OLED display driver with SPI support.

### Text Package (`pkg/text`)
Text rendering with BDF font support and embedded font option.

```go
// With embedded font (no external file needed)
renderer, _ := text.NewRendererWithEmbeddedFont(&text.Config{
    Width:     128,
    Height:    64,
    LineCount: 6,
})

// Or with custom BDF font file
renderer, _ := text.NewRenderer("path/to/font.bdf", &text.Config{
    Width:     128,
    Height:    64,
    LineCount: 6,
})
```

### Joystick Package (`pkg/joystick`)
Complete joystick/button event handling with multiple callback support.

```go
import "github.com/danielgatis/go-sh1106/pkg/joystick"

// Create joystick
joy := joystick.NewJoystick(upPin, downPin, leftPin, rightPin, btn1, btn2, btn3)

// Register multiple callbacks (returns remove function)
remove1 := joy.OnClickUp(func() {
    fmt.Println("UP clicked - Callback 1")
})

remove2 := joy.OnClickUp(func() {
    fmt.Println("UP clicked - Callback 2")
})

// Configure
joy.SetHoldDuration(500 * time.Millisecond)
joy.SetPollInterval(50 * time.Millisecond)

// Start polling
joy.Start()
defer joy.Stop()

// Remove specific callback
remove1()
```

**Supported Events:**
- `OnClick[Button]()` - Triggered on button press
- `OnHold[Button]()` - Triggered when button held (default 500ms)
- `OnRelease[Button]()` - Triggered on button release

**Available Buttons:** `Up`, `Down`, `Left`, `Right`, `Button1`, `Button2`, `Button3`

## Examples

### Basic Display Example
```bash
cd examples/basic
go run main.go "YOUR MESSAGE"
```

### Animation Example
```bash
cd examples/animation
go run main.go
```

### Joystick Example
```bash
cd examples/joystick
go run main.go
```

### Interactive Menu Example
Complete example combining display, text rendering, and joystick navigation:
```bash
cd examples/interactive
go run main.go
```

This example demonstrates:
- Interactive menu navigation with joystick
- Real-time display updates
- Multiple callback handling
- State management

## Building for Raspberry Pi

```bash
# Build for Raspberry Pi
GOOS=linux GOARCH=arm go build -o basic examples/basic/main.go

# Copy to Raspberry Pi
scp basic pi@raspberrypi.local:.
```

### License

Copyright (c) 2025-present [Daniel Gatis](https://github.com/danielgatis)

Licensed under [MIT License](./LICENSE)

### Buy me a coffee
Liked some of my work? Buy me a coffee (or more likely a beer)

<a href="https://www.buymeacoffee.com/danielgatis" target="_blank"><img src="https://bmc-cdn.nyc3.digitaloceanspaces.com/BMC-button-images/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;"></a>
