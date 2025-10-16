package joystick

import (
	"testing"
	"time"
)

func TestNewJoystick(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	if joy == nil {
		t.Fatal("NewJoystick should not return nil")
	}

	if joy.holdDuration != 500*time.Millisecond {
		t.Errorf("Expected default hold duration 500ms, got %v", joy.holdDuration)
	}

	if joy.pollInterval != 50*time.Millisecond {
		t.Errorf("Expected default poll interval 50ms, got %v", joy.pollInterval)
	}

	if joy.states == nil {
		t.Error("States map should be initialized")
	}

	if joy.running {
		t.Error("Joystick should not be running initially")
	}
}

func TestSetHoldDuration(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)
	customDuration := 1 * time.Second

	joy.SetHoldDuration(customDuration)

	if joy.holdDuration != customDuration {
		t.Errorf("Expected hold duration %v, got %v", customDuration, joy.holdDuration)
	}
}

func TestSetPollInterval(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)
	customInterval := 100 * time.Millisecond

	joy.SetPollInterval(customInterval)

	if joy.pollInterval != customInterval {
		t.Errorf("Expected poll interval %v, got %v", customInterval, joy.pollInterval)
	}
}

func TestOnClickCallbacks(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	called := 0
	callback := func() { called++ }

	removeUp := joy.OnClickUp(callback)
	if removeUp == nil {
		t.Error("OnClickUp should return a remove function")
	}

	if len(joy.onClickUp) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(joy.onClickUp))
	}

	joy.OnClickDown(callback)
	joy.OnClickLeft(callback)
	joy.OnClickRight(callback)
	joy.OnClickButton1(callback)
	joy.OnClickButton2(callback)
	joy.OnClickButton3(callback)

	if len(joy.onClickDown) != 1 {
		t.Error("OnClickDown callback should be set")
	}

	// Test multiple callbacks
	joy.OnClickUp(callback)
	if len(joy.onClickUp) != 2 {
		t.Errorf("Expected 2 callbacks, got %d", len(joy.onClickUp))
	}

	// Test callback execution
	for _, entry := range joy.onClickUp {
		entry.callback()
	}
	if called != 2 {
		t.Errorf("Expected 2 callback executions, got %d", called)
	}

	// Test callback removal
	removeUp()
	if len(joy.onClickUp) != 1 {
		t.Errorf("Expected 1 callback after removal, got %d", len(joy.onClickUp))
	}
}

func TestOnHoldCallbacks(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	callback := func() {}

	joy.OnHoldUp(callback)
	if len(joy.onHoldUp) != 1 {
		t.Error("OnHoldUp callback should be set")
	}

	joy.OnHoldDown(callback)
	if len(joy.onHoldDown) != 1 {
		t.Error("OnHoldDown callback should be set")
	}

	joy.OnHoldLeft(callback)
	if len(joy.onHoldLeft) != 1 {
		t.Error("OnHoldLeft callback should be set")
	}

	joy.OnHoldRight(callback)
	if len(joy.onHoldRight) != 1 {
		t.Error("OnHoldRight callback should be set")
	}

	joy.OnHoldButton1(callback)
	if len(joy.onHoldButton1) != 1 {
		t.Error("OnHoldButton1 callback should be set")
	}

	joy.OnHoldButton2(callback)
	if len(joy.onHoldButton2) != 1 {
		t.Error("OnHoldButton2 callback should be set")
	}

	joy.OnHoldButton3(callback)
	if len(joy.onHoldButton3) != 1 {
		t.Error("OnHoldButton3 callback should be set")
	}
}

func TestOnReleaseCallbacks(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	callback := func() {}

	joy.OnReleaseUp(callback)
	if len(joy.onReleaseUp) != 1 {
		t.Error("OnReleaseUp callback should be set")
	}

	joy.OnReleaseDown(callback)
	if len(joy.onReleaseDown) != 1 {
		t.Error("OnReleaseDown callback should be set")
	}

	joy.OnReleaseLeft(callback)
	if len(joy.onReleaseLeft) != 1 {
		t.Error("OnReleaseLeft callback should be set")
	}

	joy.OnReleaseRight(callback)
	if len(joy.onReleaseRight) != 1 {
		t.Error("OnReleaseRight callback should be set")
	}

	joy.OnReleaseButton1(callback)
	if len(joy.onReleaseButton1) != 1 {
		t.Error("OnReleaseButton1 callback should be set")
	}

	joy.OnReleaseButton2(callback)
	if len(joy.onReleaseButton2) != 1 {
		t.Error("OnReleaseButton2 callback should be set")
	}

	joy.OnReleaseButton3(callback)
	if len(joy.onReleaseButton3) != 1 {
		t.Error("OnReleaseButton3 callback should be set")
	}
}

func TestCheckButtonWithNilPin(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	// Should not panic with nil pin
	emptyCallbacks := &[]callbackEntry{}
	joy.checkButton("test", nil, emptyCallbacks, emptyCallbacks, emptyCallbacks)
}

func TestStartStop(t *testing.T) {
	joy := NewJoystick(nil, nil, nil, nil, nil, nil, nil)

	if joy.running {
		t.Error("Joystick should not be running initially")
	}

	joy.Start()
	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	if !joy.running {
		t.Error("Joystick should be running after Start()")
	}

	// Starting again should be safe
	joy.Start()
	if !joy.running {
		t.Error("Joystick should still be running")
	}

	joy.Stop()
	// Give it a moment to stop
	time.Sleep(10 * time.Millisecond)

	if joy.running {
		t.Error("Joystick should not be running after Stop()")
	}

	// Stopping again should be safe
	joy.Stop()
}
