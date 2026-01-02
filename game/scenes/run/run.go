package run

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/co0p/tankismus/game/assets"
	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/game/systems"
	"github.com/co0p/tankismus/pkg/ecs"
	"github.com/co0p/tankismus/pkg/input"
)

// Scene represents the main gameplay scene.
type Scene struct {
	world      *ecs.World
	player     ecs.EntityID
	lastUpdate time.Time
}

// New constructs a new run scene with a single player tank.
func New(_ interface{}) *Scene {
	w := ecs.NewWorld()
	player := w.NewEntity()
	w.AddComponent(player, &components.Transform{X: 100, Y: 100, Rotation: 0, Scale: 1})
	w.AddComponent(player, &components.Velocity{})
	w.AddComponent(player, &components.ControlIntent{})
	w.AddComponent(player, &components.MovementParams{
		MaxForwardSpeed:     133.3333,
		MaxBackwardSpeed:    80,
		LinearAcceleration:  200,
		LinearDeceleration:  300,
		MaxTurnRate:         3,
		AngularAcceleration: 6,
		AngularDeceleration: 9,
	})
	w.AddComponent(player, &components.Sprite{SpriteID: "player_tank"})

	return &Scene{
		world:      w,
		player:     player,
		lastUpdate: time.Now(),
	}
}

func (s *Scene) OnEnter() {}

func (s *Scene) OnExit() {}

func (s *Scene) Update(dt float64) {
	input.Poll()
	systems.InputMovementSystem(s.world, s.player)
	systems.MovementSystem(s.world, dt)
}

func (s *Scene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 10, G: 40, B: 10, A: 255})

	// ensure assets are loaded; Load is idempotent.
	_ = assets.Load()

	systems.RenderSystem(s.world, screen)
}

// World exposes the underlying ECS world for testing purposes.
func (s *Scene) World() *ecs.World {
	return s.world
}

// Player returns the player entity ID for testing purposes.
func (s *Scene) Player() ecs.EntityID {
	return s.player
}
