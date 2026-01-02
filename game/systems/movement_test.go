package systems

import (
	"math"
	"testing"

	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
)

func newTestTank(world *ecs.World) ecs.EntityID {
	id := world.NewEntity()
	world.AddComponent(id, &components.Transform{X: 0, Y: 0, Rotation: 0, Scale: 1})
	world.AddComponent(id, &components.Velocity{})
	world.AddComponent(id, &components.ControlIntent{})
	world.AddComponent(id, &components.MovementParams{
		MaxForwardSpeed:     133.3333,
		MaxBackwardSpeed:    80,
		LinearAcceleration:  250,
		LinearDeceleration:  300,
		MaxTurnRate:         3,
		AngularAcceleration: 6,
		AngularDeceleration: 9,
	})
	return id
}

func linearSpeed(world *ecs.World, id ecs.EntityID) float64 {
	cT, _ := world.GetComponent(id, components.TypeTransform)
	cV, _ := world.GetComponent(id, components.TypeVelocity)
	p := cT.(*components.Transform)
	v := cV.(*components.Velocity)

	fx := math.Cos(p.Rotation)
	fy := math.Sin(p.Rotation)
	return v.VX*fx + v.VY*fy
}

func TestMovementSystem_ForwardThrottleAcceleratesTowardMax(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	// Constant full forward throttle.
	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Throttle = 1

	dt := 0.1
	prevSpeed := 0.0
	for i := 0; i < 40; i++ {
		MovementSystem(world, dt)
		spd := linearSpeed(world, id)
		if spd < prevSpeed-1e-6 {
			t.Fatalf("speed decreased at step %d: prev=%v, got=%v", i, prevSpeed, spd)
		}
		if spd > 133.3333+1e-3 {
			t.Fatalf("speed exceeded max forward: %v", spd)
		}
		prevSpeed = spd
	}

	if final := linearSpeed(world, id); final < 120 {
		t.Fatalf("final forward speed too low, got %v, want near 133", final)
	}
}

func TestMovementSystem_BackwardThrottleAcceleratesTowardNegativeMax(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Throttle = -1

	dt := 0.1
	prevSpeed := 0.0
	for i := 0; i < 40; i++ {
		MovementSystem(world, dt)
		spd := linearSpeed(world, id)
		if spd > prevSpeed+1e-6 {
			t.Fatalf("backward speed increased toward zero at step %d: prev=%v, got=%v", i, prevSpeed, spd)
		}
		if spd < -80-1e-3 {
			t.Fatalf("speed exceeded max backward: %v", spd)
		}
		prevSpeed = spd
	}

	if final := linearSpeed(world, id); final > -65 {
		t.Fatalf("final backward speed too high, got %v, want near -80", final)
	}
}

func TestMovementSystem_DeceleratesToZeroWhenThrottleReleased(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Throttle = 1

	dt := 0.1
	// Ramp up a bit first.
	for i := 0; i < 10; i++ {
		MovementSystem(world, dt)
	}

	// Release throttle.
	intent.Throttle = 0
	prevAbs := math.Abs(linearSpeed(world, id))
	for i := 0; i < 40; i++ {
		MovementSystem(world, dt)
		spd := linearSpeed(world, id)
		abs := math.Abs(spd)
		if abs > prevAbs+1e-6 {
			t.Fatalf("speed magnitude increased during deceleration at step %d: prev=%v, got=%v", i, prevAbs, abs)
		}
		prevAbs = abs
	}

	if final := math.Abs(linearSpeed(world, id)); final > 1 {
		t.Fatalf("final speed magnitude too high after deceleration, got %v, want close to 0", final)
	}
}

func TestMovementSystem_TurnIntentCapsAngularVelocity(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Turn = 1

	dt := 0.1
	prevOmega := 0.0
	for i := 0; i < 40; i++ {
		MovementSystem(world, dt)
		cV, _ := world.GetComponent(id, components.TypeVelocity)
		v := cV.(*components.Velocity)
		if v.Angular < prevOmega-1e-6 {
			t.Fatalf("angular velocity decreased at step %d: prev=%v, got=%v", i, prevOmega, v.Angular)
		}
		if v.Angular > 3+1e-3 {
			t.Fatalf("angular velocity exceeded max turn rate: %v", v.Angular)
		}
		prevOmega = v.Angular
	}

	cV, _ := world.GetComponent(id, components.TypeVelocity)
	v := cV.(*components.Velocity)
	if v.Angular < 2.5 {
		t.Fatalf("final angular velocity too low, got %v, want near 3", v.Angular)
	}
}

func TestMovementSystem_StraightLineMotionMatchesRotation(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	// Set an arbitrary facing direction and keep angular velocity at zero.
	cT, _ := world.GetComponent(id, components.TypeTransform)
	p := cT.(*components.Transform)
	p.Rotation = math.Pi / 4 // 45 degrees

	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Throttle = 1 // move forward
	intent.Turn = 0     // no turning

	dt := 0.1
	prevX, prevY := p.X, p.Y
	fx := math.Cos(p.Rotation)
	fy := math.Sin(p.Rotation)

	for i := 0; i < 20; i++ {
		MovementSystem(world, dt)
		cT, _ = world.GetComponent(id, components.TypeTransform)
		p = cT.(*components.Transform)

		dx := p.X - prevX
		dy := p.Y - prevY
		prevX, prevY = p.X, p.Y

		// Ignore the very first step if movement is still effectively zero.
		len := math.Hypot(dx, dy)
		if len < 1e-6 {
			continue
		}

		// Displacement should be aligned with facing direction (cross product ~= 0).
		cross := dx*fy - dy*fx
		if math.Abs(cross) > 1e-5 {
			t.Fatalf("step %d: movement not aligned with rotation; cross=%v", i, cross)
		}
	}
}

func TestMovementSystem_ForwardAndTurnFollowArc(t *testing.T) {
	world := ecs.NewWorld()
	id := newTestTank(world)

	cI, _ := world.GetComponent(id, components.TypeControlIntent)
	intent := cI.(*components.ControlIntent)
	intent.Throttle = 1
	intent.Turn = 1

	dt := 0.1
	cT, _ := world.GetComponent(id, components.TypeTransform)
	p := cT.(*components.Transform)
	startX, startY := p.X, p.Y
	prevRot := p.Rotation

	for i := 0; i < 40; i++ {
		MovementSystem(world, dt)
		cT, _ = world.GetComponent(id, components.TypeTransform)
		p = cT.(*components.Transform)

		if p.Rotation <= prevRot-1e-6 {
			t.Fatalf("rotation did not increase monotonically at step %d: prev=%v, got=%v", i, prevRot, p.Rotation)
		}
		prevRot = p.Rotation
	}

	// After moving and turning, position should have changed in both axes
	// and rotation should be significantly non-zero, indicating an arc.
	cT, _ = world.GetComponent(id, components.TypeTransform)
	p = cT.(*components.Transform)
	if math.Abs(p.X-startX) < 1e-3 || math.Abs(p.Y-startY) < 1e-3 {
		t.Fatalf("expected movement in both axes, got dx=%v, dy=%v", p.X-startX, p.Y-startY)
	}
	if p.Rotation < 0.5 {
		t.Fatalf("rotation too small after turning, got %v", p.Rotation)
	}
}
