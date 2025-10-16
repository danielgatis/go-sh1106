package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgatis/go-sh1106/pkg/joystick"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func main() {
	// Load all the drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Configure GPIO pins for joystick
	// Adjust these pin numbers according to your hardware setup
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

	// Create joystick instance
	joy := joystick.NewJoystick(upPin, downPin, leftPin, rightPin, button1Pin, button2Pin, button3Pin)

	// Register multiple click callbacks (each returns a remove function)
	removeUp1 := joy.OnClickUp(func() {
		fmt.Println("UP clicked - Callback 1")
	})

	joy.OnClickUp(func() {
		fmt.Println("UP clicked - Callback 2")
	})

	joy.OnClickDown(func() {
		fmt.Println("DOWN clicked")
	})

	joy.OnClickLeft(func() {
		fmt.Println("LEFT clicked")
	})

	joy.OnClickRight(func() {
		fmt.Println("RIGHT clicked")
	})

	joy.OnClickButton1(func() {
		fmt.Println("BUTTON 1 clicked")
	})

	joy.OnClickButton2(func() {
		fmt.Println("BUTTON 2 clicked")
	})

	removeButton3 := joy.OnClickButton3(func() {
		fmt.Println("BUTTON 3 clicked - This will be removed after 5 seconds")
	})

	// Remove callback after 5 seconds (demonstrating callback removal)
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("[Removing BUTTON 3 and UP Callback 1...]")
		removeButton3()
		removeUp1()
		fmt.Println("[Callbacks removed]")
	}()

	// Register hold callbacks
	joy.OnHoldUp(func() {
		fmt.Println("UP held")
	})

	joy.OnHoldDown(func() {
		fmt.Println("DOWN held")
	})

	joy.OnHoldLeft(func() {
		fmt.Println("LEFT held")
	})

	joy.OnHoldRight(func() {
		fmt.Println("RIGHT held")
	})

	joy.OnHoldButton1(func() {
		fmt.Println("BUTTON 1 held")
	})

	joy.OnHoldButton2(func() {
		fmt.Println("BUTTON 2 held")
	})

	joy.OnHoldButton3(func() {
		fmt.Println("BUTTON 3 held")
	})

	// Register release callbacks
	joy.OnReleaseUp(func() {
		fmt.Println("UP released")
	})

	joy.OnReleaseDown(func() {
		fmt.Println("DOWN released")
	})

	joy.OnReleaseLeft(func() {
		fmt.Println("LEFT released")
	})

	joy.OnReleaseRight(func() {
		fmt.Println("RIGHT released")
	})

	joy.OnReleaseButton1(func() {
		fmt.Println("BUTTON 1 released")
	})

	joy.OnReleaseButton2(func() {
		fmt.Println("BUTTON 2 released")
	})

	joy.OnReleaseButton3(func() {
		fmt.Println("BUTTON 3 released")
	})

	// Start joystick polling
	joy.Start()
	fmt.Println("Joystick started. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	fmt.Println("\nStopping joystick...")
	joy.Stop()
	fmt.Println("Joystick stopped. Goodbye!")
}

func configurePinAsInput(pin gpio.PinIn) {
	if pin == nil {
		return
	}
	if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Printf("Failed to configure pin %s: %v", pin.Name(), err)
	}
}
