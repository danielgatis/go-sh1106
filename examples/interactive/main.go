package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/danielgatis/go-sh1106/pkg/display"
	"github.com/danielgatis/go-sh1106/pkg/joystick"
	"github.com/danielgatis/go-sh1106/pkg/text"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

type MenuState struct {
	items         []string
	selectedIndex int
	mu            sync.Mutex
}

func NewMenuState(items []string) *MenuState {
	return &MenuState{
		items:         items,
		selectedIndex: 0,
	}
}

func (m *MenuState) Up() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.selectedIndex > 0 {
		m.selectedIndex--
	}
}

func (m *MenuState) Down() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.selectedIndex < len(m.items)-1 {
		m.selectedIndex++
	}
}

func (m *MenuState) GetDisplayLines() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	lines := make([]string, 6)
	lines[0] = "=== MENU ==="

	// Show 4 items at a time, centered around selection
	startIdx := m.selectedIndex - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(m.items)-4 {
		startIdx = len(m.items) - 4
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := 0; i < 4; i++ {
		lineIdx := i + 1
		itemIdx := startIdx + i

		if itemIdx >= len(m.items) {
			lines[lineIdx] = ""
			continue
		}

		if itemIdx == m.selectedIndex {
			lines[lineIdx] = "> " + m.items[itemIdx]
		} else {
			lines[lineIdx] = "  " + m.items[itemIdx]
		}
	}

	lines[5] = fmt.Sprintf("Item %d/%d", m.selectedIndex+1, len(m.items))

	return lines
}

func (m *MenuState) GetSelected() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.items[m.selectedIndex]
}

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

	// Configure display GPIO pins
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

	// Configure joystick GPIO pins
	upPin := gpioreg.ByName("GPIO5")
	downPin := gpioreg.ByName("GPIO6")
	leftPin := gpioreg.ByName("GPIO13")
	rightPin := gpioreg.ByName("GPIO19")
	button1Pin := gpioreg.ByName("GPIO20")
	button2Pin := gpioreg.ByName("GPIO21")
	button3Pin := gpioreg.ByName("GPIO26")

	// Configure all pins as inputs with pull-up resistors
	configurePinAsInput(upPin)
	configurePinAsInput(downPin)
	configurePinAsInput(leftPin)
	configurePinAsInput(rightPin)
	configurePinAsInput(button1Pin)
	configurePinAsInput(button2Pin)
	configurePinAsInput(button3Pin)

	// Create menu state
	menuItems := []string{
		"Option 1",
		"Option 2",
		"Option 3",
		"Option 4",
		"Option 5",
		"Settings",
		"About",
		"Exit",
	}
	menu := NewMenuState(menuItems)

	// Create joystick instance
	joy := joystick.NewJoystick(upPin, downPin, leftPin, rightPin, button1Pin, button2Pin, button3Pin)

	// Function to update display
	updateDisplay := func() {
		textRenderer.SetTexts(menu.GetDisplayLines())
		if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
			log.Printf("Error drawing: %v", err)
		}
	}

	// Register joystick callbacks
	joy.OnClickUp(func() {
		menu.Up()
		updateDisplay()
		fmt.Printf("Selected: %s\n", menu.GetSelected())
	})

	joy.OnClickDown(func() {
		menu.Down()
		updateDisplay()
		fmt.Printf("Selected: %s\n", menu.GetSelected())
	})

	joy.OnClickButton1(func() {
		selected := menu.GetSelected()
		fmt.Printf("Confirmed: %s\n", selected)

		// Show confirmation on display
		textRenderer.SetTexts([]string{
			"SELECTED:",
			"",
			selected,
			"",
			"Press B2 to go",
			"back to menu",
		})
		if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
			log.Printf("Error drawing: %v", err)
		}
	})

	joy.OnClickButton2(func() {
		fmt.Println("Back to menu")
		updateDisplay()
	})

	joy.OnClickButton3(func() {
		fmt.Println("Exit requested")
		textRenderer.SetTexts([]string{
			"",
			"",
			"  GOODBYE!",
			"",
			"",
			"",
		})
		if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
			log.Printf("Error drawing: %v", err)
		}
		time.Sleep(2 * time.Second)
		os.Exit(0)
	})

	// Hold callbacks for faster navigation
	joy.OnHoldUp(func() {
		menu.Up()
		updateDisplay()
		fmt.Printf("Selected: %s\n", menu.GetSelected())
	})

	joy.OnHoldDown(func() {
		menu.Down()
		updateDisplay()
		fmt.Printf("Selected: %s\n", menu.GetSelected())
	})

	// Show welcome screen
	textRenderer.SetTexts([]string{
		"",
		"  INTERACTIVE",
		"    MENU",
		"",
		"Press any key",
		"to start",
	})
	if err := dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{}); err != nil {
		log.Fatal(err)
	}

	// Wait for any button press to start
	started := false
	var removeCallbacks []joystick.RemoveCallbackFunc

	startMenu := func() {
		if started {
			return
		}
		started = true
		fmt.Println("Starting interactive menu...")

		// Remove startup callbacks
		for _, remove := range removeCallbacks {
			remove()
		}

		// Show initial menu
		updateDisplay()
	}

	removeCallbacks = append(removeCallbacks, joy.OnClickUp(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickDown(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickLeft(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickRight(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickButton1(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickButton2(startMenu))
	removeCallbacks = append(removeCallbacks, joy.OnClickButton3(startMenu))

	// Start joystick polling
	joy.Start()
	fmt.Println("Interactive menu started. Press Ctrl+C to exit.")
	fmt.Println("Controls:")
	fmt.Println("  UP/DOWN    - Navigate menu")
	fmt.Println("  BUTTON 1   - Select item")
	fmt.Println("  BUTTON 2   - Back to menu")
	fmt.Println("  BUTTON 3   - Exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	fmt.Println("\nStopping...")
	joy.Stop()
	textRenderer.SetTexts([]string{
		"",
		"",
		"  SHUTDOWN",
		"",
		"",
		"",
	})
	dev.Draw(textRenderer.Bounds(), textRenderer.Image(), image.Point{})
	time.Sleep(1 * time.Second)
	fmt.Println("Goodbye!")
}

func configurePinAsInput(pin gpio.PinIn) {
	if pin == nil {
		return
	}
	if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Printf("Failed to configure pin %s: %v", pin.Name(), err)
	}
}
