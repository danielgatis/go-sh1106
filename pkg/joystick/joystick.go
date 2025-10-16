package joystick

import (
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
)

// ButtonCallback is a function that is called when a button event occurs
type ButtonCallback func()

// RemoveCallbackFunc is a function that removes a callback
type RemoveCallbackFunc func()

// ButtonState represents the current state of a button
type ButtonState struct {
	pressed   bool
	holdStart time.Time
}

// callbackEntry represents a callback with a unique ID
type callbackEntry struct {
	id       int
	callback ButtonCallback
}

// Joystick represents a joystick with directional controls and buttons
type Joystick struct {
	up    gpio.PinIn
	down  gpio.PinIn
	left  gpio.PinIn
	right gpio.PinIn

	button1 gpio.PinIn
	button2 gpio.PinIn
	button3 gpio.PinIn

	// Button states
	states map[string]*ButtonState
	mu     sync.RWMutex

	// Click callbacks
	onClickUp      []callbackEntry
	onClickDown    []callbackEntry
	onClickLeft    []callbackEntry
	onClickRight   []callbackEntry
	onClickButton1 []callbackEntry
	onClickButton2 []callbackEntry
	onClickButton3 []callbackEntry

	// Hold callbacks
	onHoldUp      []callbackEntry
	onHoldDown    []callbackEntry
	onHoldLeft    []callbackEntry
	onHoldRight   []callbackEntry
	onHoldButton1 []callbackEntry
	onHoldButton2 []callbackEntry
	onHoldButton3 []callbackEntry

	// Release callbacks
	onReleaseUp      []callbackEntry
	onReleaseDown    []callbackEntry
	onReleaseLeft    []callbackEntry
	onReleaseRight   []callbackEntry
	onReleaseButton1 []callbackEntry
	onReleaseButton2 []callbackEntry
	onReleaseButton3 []callbackEntry

	// Configuration
	holdDuration time.Duration
	pollInterval time.Duration

	// Control
	stopChan       chan struct{}
	running        bool
	nextCallbackID int
}

// NewJoystick creates a new joystick instance
func NewJoystick(up, down, left, right, button1, button2, button3 gpio.PinIn) *Joystick {
	return &Joystick{
		up:           up,
		down:         down,
		left:         left,
		right:        right,
		button1:      button1,
		button2:      button2,
		button3:      button3,
		states:       make(map[string]*ButtonState),
		holdDuration: 500 * time.Millisecond,
		pollInterval: 50 * time.Millisecond,
		stopChan:     make(chan struct{}),
	}
}

// Click callbacks
func (j *Joystick) OnClickUp(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickUp, callback)
}
func (j *Joystick) OnClickDown(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickDown, callback)
}
func (j *Joystick) OnClickLeft(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickLeft, callback)
}
func (j *Joystick) OnClickRight(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickRight, callback)
}
func (j *Joystick) OnClickButton1(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickButton1, callback)
}
func (j *Joystick) OnClickButton2(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickButton2, callback)
}
func (j *Joystick) OnClickButton3(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onClickButton3, callback)
}

// Hold callbacks
func (j *Joystick) OnHoldUp(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldUp, callback)
}
func (j *Joystick) OnHoldDown(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldDown, callback)
}
func (j *Joystick) OnHoldLeft(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldLeft, callback)
}
func (j *Joystick) OnHoldRight(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldRight, callback)
}
func (j *Joystick) OnHoldButton1(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldButton1, callback)
}
func (j *Joystick) OnHoldButton2(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldButton2, callback)
}
func (j *Joystick) OnHoldButton3(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onHoldButton3, callback)
}

// Release callbacks
func (j *Joystick) OnReleaseUp(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseUp, callback)
}
func (j *Joystick) OnReleaseDown(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseDown, callback)
}
func (j *Joystick) OnReleaseLeft(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseLeft, callback)
}
func (j *Joystick) OnReleaseRight(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseRight, callback)
}
func (j *Joystick) OnReleaseButton1(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseButton1, callback)
}
func (j *Joystick) OnReleaseButton2(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseButton2, callback)
}
func (j *Joystick) OnReleaseButton3(callback ButtonCallback) RemoveCallbackFunc {
	return j.addCallback(&j.onReleaseButton3, callback)
}

// addCallback adds a callback to a list and returns a function to remove it
func (j *Joystick) addCallback(list *[]callbackEntry, callback ButtonCallback) RemoveCallbackFunc {
	j.mu.Lock()
	defer j.mu.Unlock()

	id := j.nextCallbackID
	j.nextCallbackID++

	*list = append(*list, callbackEntry{
		id:       id,
		callback: callback,
	})

	return func() {
		j.removeCallback(list, id)
	}
}

// removeCallback removes a callback from a list by ID
func (j *Joystick) removeCallback(list *[]callbackEntry, id int) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, entry := range *list {
		if entry.id == id {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return
		}
	}
}

// SetHoldDuration sets the duration after which a button press is considered a hold
func (j *Joystick) SetHoldDuration(duration time.Duration) {
	j.holdDuration = duration
}

// SetPollInterval sets the interval at which button states are polled
func (j *Joystick) SetPollInterval(interval time.Duration) {
	j.pollInterval = interval
}

// Start begins polling the joystick buttons
func (j *Joystick) Start() {
	if j.running {
		return
	}

	j.running = true
	go j.pollLoop()
}

// Stop stops polling the joystick buttons
func (j *Joystick) Stop() {
	if !j.running {
		return
	}

	close(j.stopChan)
	j.running = false
}

// pollLoop continuously polls button states and triggers callbacks
func (j *Joystick) pollLoop() {
	ticker := time.NewTicker(j.pollInterval)
	defer ticker.Stop()

	buttons := []struct {
		name             string
		pin              gpio.PinIn
		clickCallbacks   *[]callbackEntry
		holdCallbacks    *[]callbackEntry
		releaseCallbacks *[]callbackEntry
	}{
		{"up", j.up, &j.onClickUp, &j.onHoldUp, &j.onReleaseUp},
		{"down", j.down, &j.onClickDown, &j.onHoldDown, &j.onReleaseDown},
		{"left", j.left, &j.onClickLeft, &j.onHoldLeft, &j.onReleaseLeft},
		{"right", j.right, &j.onClickRight, &j.onHoldRight, &j.onReleaseRight},
		{"button1", j.button1, &j.onClickButton1, &j.onHoldButton1, &j.onReleaseButton1},
		{"button2", j.button2, &j.onClickButton2, &j.onHoldButton2, &j.onReleaseButton2},
		{"button3", j.button3, &j.onClickButton3, &j.onHoldButton3, &j.onReleaseButton3},
	}

	for {
		select {
		case <-j.stopChan:
			return
		case <-ticker.C:
			for _, btn := range buttons {
				j.checkButton(btn.name, btn.pin, btn.clickCallbacks, btn.holdCallbacks, btn.releaseCallbacks)
			}
		}
	}
}

// checkButton checks the state of a button and triggers appropriate callbacks
func (j *Joystick) checkButton(name string, pin gpio.PinIn, clickCallbacks, holdCallbacks, releaseCallbacks *[]callbackEntry) {
	if pin == nil {
		return
	}

	// Read current state (assuming LOW = pressed for pull-up configuration)
	pressed := pin.Read() == gpio.Low

	j.mu.Lock()
	state, exists := j.states[name]
	if !exists {
		state = &ButtonState{}
		j.states[name] = state
	}

	wasPressed := state.pressed
	state.pressed = pressed

	if pressed && !wasPressed {
		// Button just pressed
		state.holdStart = time.Now()
		// Execute all click callbacks
		for _, entry := range *clickCallbacks {
			go entry.callback()
		}
	} else if pressed && wasPressed {
		// Button is being held
		holdTime := time.Since(state.holdStart)
		if holdTime >= j.holdDuration {
			// Trigger hold callback (only once when threshold is reached)
			if holdTime < j.holdDuration+j.pollInterval {
				for _, entry := range *holdCallbacks {
					go entry.callback()
				}
			}
		}
	} else if !pressed && wasPressed {
		// Button just released
		for _, entry := range *releaseCallbacks {
			go entry.callback()
		}
	}

	j.mu.Unlock()
}
