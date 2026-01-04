package maps

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/co0p/tankismus/game/assets"
)

func newTestSprite(w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(color.White)
	return img
}

func TestGeneratedMapComposesTilemap(t *testing.T) {
	// Reset registry so this test is isolated.
	assets.Registry = map[string]*ebiten.Image{}

	// Build a small image that yields a simple road crossing pattern on grass.
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	grass := color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	road := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grass
			if (x == 1 && y != 1) || (y == 1 && x != 1) || (x == 1 && y == 1) {
				c = road
			}
			img.Set(x, y, c)
		}
	}

	m, err := GenerateFromImage(img)
	if err != nil {
		t.Fatalf("GenerateFromImage failed: %v", err)
	}

	tileSize := 8

	// Register dummy sprites for all tile IDs used in the generated map.
	seen := map[string]struct{}{}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			id := m.Tiles[y][x]
			if id == "" {
				t.Fatalf("empty tile ID at (%d,%d)", x, y)
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			assets.RegisterSpriteForTest(id, newTestSprite(tileSize, tileSize))
		}
	}

	const composedID = "test_generated_tilemap"
	imgOut, err := assets.ComposeTilemap(composedID, m, tileSize)
	if err != nil {
		t.Fatalf("ComposeTilemap failed: %v", err)
	}
	if imgOut == nil {
		t.Fatalf("ComposeTilemap returned nil image")
	}

	w, h := imgOut.Size()
	if wantW, wantH := m.Width*tileSize, m.Height*tileSize; w != wantW || h != wantH {
		t.Fatalf("unexpected composed image size: got (%d,%d) want (%d,%d)", w, h, wantW, wantH)
	}

	fromRegistry := assets.GetSprite(composedID)
	if fromRegistry == nil {
		t.Fatalf("expected composed tilemap to be registered under %q", composedID)
	}
	if fromRegistry != imgOut {
		t.Fatalf("registry image pointer mismatch; registry=%p, returned=%p", fromRegistry, imgOut)
	}
}
