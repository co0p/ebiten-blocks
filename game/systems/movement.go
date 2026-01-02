package systems

import (
	"math"

	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
)

// MovementSystem updates velocity based on control intent and movement
// parameters and then applies velocity to transform for all
// entities that participate in the tank movement model.
func MovementSystem(world *ecs.World, dt float64) {
	required := ecs.MaskFor(
		components.TypeTransform,
		components.TypeVelocity,
		components.TypeControlIntent,
		components.TypeMovementParams,
	)
	entities := world.Find(required)
	for _, id := range entities {
		cT, okT := world.GetComponent(id, components.TypeTransform)
		cV, okV := world.GetComponent(id, components.TypeVelocity)
		cIntent, okCI := world.GetComponent(id, components.TypeControlIntent)
		cParams, okMP := world.GetComponent(id, components.TypeMovementParams)
		if !okT || !okV || !okCI || !okMP {
			panic("MovementSystem: entity missing required movement components")
		}

		p, okP := cT.(*components.Transform)
		v, okVel := cV.(*components.Velocity)
		intent, okI := cIntent.(*components.ControlIntent)
		params, okM := cParams.(*components.MovementParams)
		if !okP || !okVel || !okI || !okM {
			panic("MovementSystem: component type assertion failed")
		}

		applyMovementModel(p, v, intent, params, dt)
	}
}

func applyMovementModel(p *components.Transform, v *components.Velocity, intent *components.ControlIntent, params *components.MovementParams, dt float64) {
	if dt <= 0 {
		return
	}

	// Forward direction from current rotation.
	fx := math.Cos(p.Rotation)
	fy := math.Sin(p.Rotation)

	// Project current velocity onto forward direction to get scalar speed.
	currentSpeed := v.VX*fx + v.VY*fy

	// Target linear speed from throttle and movement params.
	clampedThrottle := clamp(intent.Throttle, -1, 1)
	var targetSpeed float64
	if clampedThrottle > 0 {
		targetSpeed = clampedThrottle * params.MaxForwardSpeed
	} else if clampedThrottle < 0 {
		targetSpeed = clampedThrottle * params.MaxBackwardSpeed
	} else {
		targetSpeed = 0
	}

	// Choose acceleration vs deceleration.
	accel := params.LinearDeceleration
	if clampedThrottle != 0 {
		accel = params.LinearAcceleration
	}
	if accel < 0 {
		accel = 0
	}

	maxDelta := accel * dt
	speedDelta := targetSpeed - currentSpeed
	if maxDelta > 0 {
		if speedDelta > maxDelta {
			speedDelta = maxDelta
		} else if speedDelta < -maxDelta {
			speedDelta = -maxDelta
		}
		currentSpeed += speedDelta
	}

	// Clamp final speed to bounds.
	if currentSpeed > params.MaxForwardSpeed {
		currentSpeed = params.MaxForwardSpeed
	}
	if currentSpeed < -params.MaxBackwardSpeed {
		currentSpeed = -params.MaxBackwardSpeed
	}

	// Reconstruct world-space linear velocity from scalar speed.
	v.VX = fx * currentSpeed
	v.VY = fy * currentSpeed

	// Angular velocity.
	clampedTurn := clamp(intent.Turn, -1, 1)
	targetOmega := clampedTurn * params.MaxTurnRate
	currentOmega := v.Angular

	angAccel := params.AngularDeceleration
	if clampedTurn != 0 {
		angAccel = params.AngularAcceleration
	}
	if angAccel < 0 {
		angAccel = 0
	}

	maxOmegaDelta := angAccel * dt
	omegaDelta := targetOmega - currentOmega
	if maxOmegaDelta > 0 {
		if omegaDelta > maxOmegaDelta {
			omegaDelta = maxOmegaDelta
		} else if omegaDelta < -maxOmegaDelta {
			omegaDelta = -maxOmegaDelta
		}
		currentOmega += omegaDelta
	}

	// Clamp final angular velocity.
	if currentOmega > params.MaxTurnRate {
		currentOmega = params.MaxTurnRate
	}
	if currentOmega < -params.MaxTurnRate {
		currentOmega = -params.MaxTurnRate
	}
	v.Angular = currentOmega

	// Integrate updated velocity into transform.
	p.X += v.VX * dt
	p.Y += v.VY * dt
	p.Rotation += v.Angular * dt
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
