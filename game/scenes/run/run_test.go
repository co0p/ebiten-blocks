package run

import (
	"testing"

	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
	"github.com/co0p/tankismus/pkg/input"
)

func TestNewRunScene_HasRequiredPlayerComponents(t *testing.T) {
	s := New(nil)
	world := s.World()
	player := s.Player()

	required := []ecs.ComponentType{
		components.TypeTransform,
		components.TypeVelocity,
		components.TypeControlIntent,
		components.TypeMovementParams,
		components.TypeSprite,
	}

	for _, ct := range required {
		if !world.HasComponent(player, ct) {
			t.Fatalf("player missing component type %v", ct)
		}
	}
}

func TestRunScene_UpdateAppliesInputAndMovement(t *testing.T) {
	s := New(nil)
	world := s.World()
	player := s.Player()
	testMgr := input.NewTestManager()
	input.SetManager(testMgr)

	cT, _ := world.GetComponent(player, components.TypeTransform)
	p := cT.(*components.Transform)
	p.X, p.Y, p.Rotation = 0, 0, 0

	// Hold forward key down.
	testMgr.State[input.ActionMoveForward] = true

	// Call Update repeatedly; Scene should poll input, update intent, and apply movement.
	for i := 0; i < 10; i++ {
		s.Update(0.1)
	}

	if p.X <= 0 {
		t.Fatalf("expected player to move forward in +X, got X=%v", p.X)
	}
}
