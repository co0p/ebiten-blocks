package run

import (
	"image/color"
	"testing"

	"github.com/co0p/tankismus/game/assets"
	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
	"github.com/co0p/tankismus/pkg/input"
	"github.com/hajimehoshi/ebiten/v2"
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

func TestNewRunScene_HasMapAndTilemapEntity(t *testing.T) {
	// Provide test sprites for tiles so that ComposeTilemap in the scene
	// constructor can succeed without requiring on-disk assets.
	tileSize := 16
	img := ebiten.NewImage(tileSize, tileSize)
	img.Fill(color.White)
	assets.RegisterSpriteForTest("tileGrass1", img)
	assets.RegisterSpriteForTest("tileGrass2", img)

	s := New(nil)
	world := s.World()
	player := s.Player()

	if s.levelMap == nil {
		t.Fatalf("expected levelMap to be initialized")
	}
	if s.levelMap.Width == 0 || s.levelMap.Height == 0 {
		t.Fatalf("expected levelMap to have non-zero dimensions, got %dx%d", s.levelMap.Width, s.levelMap.Height)
	}
	if len(s.levelMap.Tiles) == 0 {
		t.Fatalf("expected levelMap to have tiles")
	}

	// Tilemap entity should be distinct from the player and have
	// Transform, Sprite, and RenderOrder components.
	if s.tilemap == 0 {
		t.Fatalf("expected a non-zero tilemap entity ID")
	}
	if s.tilemap == player {
		t.Fatalf("tilemap entity must not be the player entity")
	}

	required := []ecs.ComponentType{
		components.TypeTransform,
		components.TypeSprite,
		components.TypeRenderOrder,
	}
	for _, ct := range required {
		if !world.HasComponent(s.tilemap, ct) {
			t.Fatalf("tilemap entity missing component type %v", ct)
		}
	}

	cS, _ := world.GetComponent(s.tilemap, components.TypeSprite)
	sprite := cS.(*components.Sprite)
	if sprite.SpriteID != "tilemap_ground" {
		t.Fatalf("tilemap sprite ID = %q, want %q", sprite.SpriteID, "tilemap_ground")
	}

	cZ, _ := world.GetComponent(s.tilemap, components.TypeRenderOrder)
	ro := cZ.(*components.RenderOrder)
	if ro.Z >= 10 {
		t.Fatalf("expected tilemap render order to be below player layer, got %d", ro.Z)
	}

	// Player should still have all required components, including RenderOrder.
	requiredPlayer := []ecs.ComponentType{
		components.TypeTransform,
		components.TypeVelocity,
		components.TypeControlIntent,
		components.TypeMovementParams,
		components.TypeSprite,
		components.TypeRenderOrder,
	}
	for _, ct := range requiredPlayer {
		if !world.HasComponent(player, ct) {
			t.Fatalf("player missing component type %v after tilemap integration", ct)
		}
	}
}
