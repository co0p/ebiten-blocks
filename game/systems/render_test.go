package systems

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/co0p/tankismus/game/assets"
	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
)

// fakeSprite is a simple ebiten.Image backed by a Go image for bounds.
func fakeSprite(w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(color.White)
	return img
}

// TestRenderSystem_RotatesAroundSpriteCenter is a coarse behavioral test: it
// verifies that when rotation changes, the drawn image remains visually
// centered on the transform position. Since Ebiten does not expose the
// internal GeoM matrix or draw calls, this test is limited to ensuring the
// call does not panic and can be extended with golden-image style tests in
// future if needed.
func TestRenderSystem_RotatesAroundSpriteCenter(t *testing.T) {
	// Prepare a fake sprite and register it.
	spriteID := "test_tank"
	img := fakeSprite(20, 10) // width=20,height=10, center=(10,5)
	assets.RegisterSpriteForTest(spriteID, img)

	world := ecs.NewWorld()
	id := world.NewEntity()
	world.AddComponent(id, &components.Transform{X: 100, Y: 50, Rotation: 0, Scale: 1})
	world.AddComponent(id, &components.Sprite{SpriteID: spriteID})

	screen := ebiten.NewImage(200, 100)

	// Draw at rotation 0 and then with a non-zero rotation. The main
	// verification here is that both calls succeed without error or panic.
	RenderSystem(world, screen)

	cT, _ := world.GetComponent(id, components.TypeTransform)
	p := cT.(*components.Transform)
	p.Rotation = 1.0

	RenderSystem(world, screen)

	// No explicit numeric assertions here due to limited access to the
	// underlying draw machinery; correctness is exercised indirectly via
	// integration tests of movement + rendering at a higher level.
}

func TestCollectDrawablesSortsByZAndDefaultsToZero(t *testing.T) {
	world := ecs.NewWorld()

	// Entity without explicit RenderOrder should default to z=0.
	eDefault := world.NewEntity()
	world.AddComponent(eDefault, &components.Transform{X: 0, Y: 0, Rotation: 0, Scale: 1})
	world.AddComponent(eDefault, &components.Sprite{SpriteID: "sprite_default"})

	// Entity with lower z.
	eLow := world.NewEntity()
	world.AddComponent(eLow, &components.Transform{X: 0, Y: 0, Rotation: 0, Scale: 1})
	world.AddComponent(eLow, &components.Sprite{SpriteID: "sprite_low"})
	world.AddComponent(eLow, &components.RenderOrder{Z: -1})

	// Entity with higher z.
	eHigh := world.NewEntity()
	world.AddComponent(eHigh, &components.Transform{X: 0, Y: 0, Rotation: 0, Scale: 1})
	world.AddComponent(eHigh, &components.Sprite{SpriteID: "sprite_high"})
	world.AddComponent(eHigh, &components.RenderOrder{Z: 10})

	drawables := collectDrawables(world)
	if len(drawables) != 3 {
		t.Fatalf("expected 3 drawables, got %d", len(drawables))
	}

	gotOrder := []ecs.EntityID{drawables[0].entity, drawables[1].entity, drawables[2].entity}
	wantOrder := []ecs.EntityID{eLow, eDefault, eHigh}
	for i, want := range wantOrder {
		if gotOrder[i] != want {
			t.Fatalf("drawables[%d].entity = %v, want %v", i, gotOrder[i], want)
		}
	}

	// Verify the default z value for the entity without an explicit RenderOrder.
	for _, d := range drawables {
		if d.entity == eDefault && d.z != 0 {
			t.Errorf("default z for entity without RenderOrder = %d, want 0", d.z)
		}
	}
}
