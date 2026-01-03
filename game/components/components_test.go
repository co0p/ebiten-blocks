package components

import "testing"

func TestControlIntentComponentType(t *testing.T) {
	c := ControlIntent{}
	if c.Type() != TypeControlIntent {
		t.Fatalf("ControlIntent Type() = %v, want %v", c.Type(), TypeControlIntent)
	}

	if c.Throttle != 0 {
		t.Errorf("default ControlIntent.Throttle = %v, want 0", c.Throttle)
	}
	if c.Turn != 0 {
		t.Errorf("default ControlIntent.Turn = %v, want 0", c.Turn)
	}
}

func TestMovementParamsComponentTypeAndDefaults(t *testing.T) {
	m := MovementParams{}
	if m.Type() != TypeMovementParams {
		t.Fatalf("MovementParams Type() = %v, want %v", m.Type(), TypeMovementParams)
	}

	// Ensure defaults are non-negative and reasonably sized for initial tuning.
	if m.MaxForwardSpeed < 0 {
		t.Errorf("MaxForwardSpeed default = %v, want non-negative", m.MaxForwardSpeed)
	}
	if m.MaxBackwardSpeed < 0 {
		t.Errorf("MaxBackwardSpeed default = %v, want non-negative", m.MaxBackwardSpeed)
	}
	if m.LinearAcceleration < 0 {
		t.Errorf("LinearAcceleration default = %v, want non-negative", m.LinearAcceleration)
	}
	if m.LinearDeceleration < 0 {
		t.Errorf("LinearDeceleration default = %v, want non-negative", m.LinearDeceleration)
	}
	if m.MaxTurnRate < 0 {
		t.Errorf("MaxTurnRate default = %v, want non-negative", m.MaxTurnRate)
	}
	if m.AngularAcceleration < 0 {
		t.Errorf("AngularAcceleration default = %v, want non-negative", m.AngularAcceleration)
	}
	if m.AngularDeceleration < 0 {
		t.Errorf("AngularDeceleration default = %v, want non-negative", m.AngularDeceleration)
	}
}

func TestRenderOrderComponentTypeAndDefaults(t *testing.T) {
	r := RenderOrder{}
	if r.Type() != TypeRenderOrder {
		t.Fatalf("RenderOrder Type() = %v, want %v", r.Type(), TypeRenderOrder)
	}

	// Default Z should be zero, which can be treated as the baseline layer.
	if r.Z != 0 {
		t.Errorf("default RenderOrder.Z = %v, want 0", r.Z)
	}
}
