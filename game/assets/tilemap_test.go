package assets

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	mappkg "github.com/co0p/tankismus/pkg/map"
)

// helper for creating a solid-color sprite of a given size.
func newTestSprite(w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(color.White)
	return img
}

func TestComposeTilemapRegistersComposedImage(t *testing.T) {
	// Reset registry so this test is isolated.
	Registry = map[string]*ebiten.Image{}

	tileSize := 8
	// Register two grass variant sprites that the map will reference.
	RegisterSpriteForTest("tileGrass1", newTestSprite(tileSize, tileSize))
	RegisterSpriteForTest("tileGrass2", newTestSprite(tileSize, tileSize))

	// Build a small deterministic grass map.
	m, err := mappkg.NewGrassMap(1, 3, 2)
	if err != nil {
		t.Fatalf("NewGrassMap failed: %v", err)
	}

	const composedID = "test_tilemap_ground"
	img, err := ComposeTilemap(composedID, m, tileSize)
	if err != nil {
		t.Fatalf("ComposeTilemap failed: %v", err)
	}

	if img == nil {
		t.Fatalf("ComposeTilemap returned nil image")
	}

	// Expect composed dimensions to match map size * tileSize.
	w, h := img.Size()
	if wantW, wantH := m.Width*tileSize, m.Height*tileSize; w != wantW || h != wantH {
		t.Fatalf("unexpected composed image size: got (%d,%d) want (%d,%d)", w, h, wantW, wantH)
	}

	// The composed image should also be registered in the assets registry
	// under the provided sprite ID so that scenes can reference it.
	fromRegistry := GetSprite(composedID)
	if fromRegistry == nil {
		t.Fatalf("expected composed tilemap to be registered under %q", composedID)
	}
	if fromRegistry != img {
		t.Fatalf("registry image pointer mismatch; registry=%p, returned=%p", fromRegistry, img)
	}
}
