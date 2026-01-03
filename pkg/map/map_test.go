package mappkg

import (
	"encoding/json"
	"math"
	"testing"
)

func TestMapJSONRoundTrip(t *testing.T) {
	m := &Map{
		Width:  2,
		Height: 2,
		Seed:   1234567890,
		Tiles: [][]string{
			{"tileGrass1", "tileGrass2"},
			{"tileGrass2", "tileGrass1"},
		},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Map
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Width != m.Width || decoded.Height != m.Height || decoded.Seed != m.Seed {
		t.Fatalf("basic fields did not round-trip: %#v", decoded)
	}

	if len(decoded.Tiles) != len(m.Tiles) {
		t.Fatalf("tiles height mismatch: got %d want %d", len(decoded.Tiles), len(m.Tiles))
	}
	for y := range m.Tiles {
		if len(decoded.Tiles[y]) != len(m.Tiles[y]) {
			t.Fatalf("tiles width mismatch in row %d", y)
		}
		for x := range m.Tiles[y] {
			if decoded.Tiles[y][x] != m.Tiles[y][x] {
				t.Fatalf("tile mismatch at (%d,%d): got %q want %q", x, y, decoded.Tiles[y][x], m.Tiles[y][x])
			}
		}
	}
}

func TestMapValidateRejectsMismatchedDimensions(t *testing.T) {
	m := &Map{
		Width:  2,
		Height: 2,
		Seed:   1,
		Tiles:  [][]string{{"tileGrass1"}},
	}

	if err := m.validate(); err == nil {
		t.Fatalf("expected validation error for mismatched tiles dimensions, got nil")
	}
}

func TestMapValidateRejectsOutOfRangeSeed(t *testing.T) {
	m := &Map{
		Width:  2,
		Height: 2,
		Seed:   MaxSeed + 1,
		Tiles: [][]string{
			{"tileGrass1", "tileGrass2"},
			{"tileGrass2", "tileGrass1"},
		},
	}

	if err := m.validate(); err == nil {
		t.Fatalf("expected validation error for out-of-range seed, got nil")
	}
}

func TestNewGrassMapDeterministicAndValid(t *testing.T) {
	seed := int64(42)
	width, height := 4, 3

	first, err := NewGrassMap(seed, width, height)
	if err != nil {
		t.Fatalf("NewGrassMap failed: %v", err)
	}

	second, err := NewGrassMap(seed, width, height)
	if err != nil {
		t.Fatalf("NewGrassMap second call failed: %v", err)
	}

	if first.Width != width || first.Height != height {
		t.Fatalf("unexpected dimensions: %#v", first)
	}
	if len(first.Tiles) != height {
		t.Fatalf("tiles height mismatch: got %d want %d", len(first.Tiles), height)
	}
	for y := 0; y < height; y++ {
		if len(first.Tiles[y]) != width {
			t.Fatalf("tiles width mismatch in row %d: got %d want %d", y, len(first.Tiles[y]), width)
		}
		for x := 0; x < width; x++ {
			tile := first.Tiles[y][x]
			if tile != "tileGrass1" && tile != "tileGrass2" {
				t.Fatalf("unexpected tile at (%d,%d): %q", x, y, tile)
			}
		}
	}

	// Determinism: two maps with same seed and dimensions must match.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if first.Tiles[y][x] != second.Tiles[y][x] {
				t.Fatalf("non-deterministic tile at (%d,%d): %q vs %q", x, y, first.Tiles[y][x], second.Tiles[y][x])
			}
		}
	}
}

func TestTileAtReturnsExpectedTilesAndOutOfBoundsFalse(t *testing.T) {
	m, err := NewGrassMap(1, 3, 2)
	if err != nil {
		t.Fatalf("NewGrassMap failed: %v", err)
	}

	// Sanity: dimensions match what we requested.
	if m.Width != 3 || m.Height != 2 {
		t.Fatalf("unexpected dimensions: %#v", m)
	}

	// In-bounds queries should return a tile ID and ok == true.
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tile, ok := m.TileAt(x, y)
			if !ok {
				t.Fatalf("expected ok for in-bounds TileAt(%d,%d)", x, y)
			}
			if tile != "tileGrass1" && tile != "tileGrass2" {
				t.Fatalf("unexpected tile at (%d,%d): %q", x, y, tile)
			}
		}
	}

	// Out-of-bounds coordinates should return ok == false.
	if _, ok := m.TileAt(-1, 0); ok {
		t.Fatalf("expected ok == false for TileAt(-1,0)")
	}
	if _, ok := m.TileAt(0, -1); ok {
		t.Fatalf("expected ok == false for TileAt(0,-1)")
	}
	if _, ok := m.TileAt(m.Width, 0); ok {
		t.Fatalf("expected ok == false for TileAt(width,0)")
	}
	if _, ok := m.TileAt(0, m.Height); ok {
		t.Fatalf("expected ok == false for TileAt(0,height)")
	}
}

func TestTileAtWorldMapsWorldToCorrectTileAndOutOfBounds(t *testing.T) {
	m, err := NewGrassMap(1, 3, 2)
	if err != nil {
		t.Fatalf("NewGrassMap failed: %v", err)
	}

	tileSize := 16.0

	// World coordinates at tile centers should map to the corresponding tiles.
	for ty := 0; ty < m.Height; ty++ {
		for tx := 0; tx < m.Width; tx++ {
			worldX := (float64(tx) + 0.5) * tileSize
			worldY := (float64(ty) + 0.5) * tileSize

			tileByWorld, ok := m.TileAtWorld(worldX, worldY, tileSize)
			if !ok {
				t.Fatalf("expected ok for in-bounds TileAtWorld at (%f,%f)", worldX, worldY)
			}

			tileByIndex, ok := m.TileAt(tx, ty)
			if !ok {
				t.Fatalf("expected ok for TileAt(%d,%d)", tx, ty)
			}

			if tileByWorld != tileByIndex {
				t.Fatalf("TileAtWorld mismatch at tile (%d,%d): got %q want %q", tx, ty, tileByWorld, tileByIndex)
			}
		}
	}

	// Out-of-bounds world coordinates should return ok == false.
	// Negative coordinates.
	if _, ok := m.TileAtWorld(-1, 0, tileSize); ok {
		t.Fatalf("expected ok == false for TileAtWorld(-1,0)")
	}
	if _, ok := m.TileAtWorld(0, -1, tileSize); ok {
		t.Fatalf("expected ok == false for TileAtWorld(0,-1)")
	}

	// Coordinates beyond the map extents.
	maxX := float64(m.Width) * tileSize
	maxY := float64(m.Height) * tileSize
	if _, ok := m.TileAtWorld(maxX+1, 0, tileSize); ok {
		t.Fatalf("expected ok == false for TileAtWorld(x beyond map,0)")
	}
	if _, ok := m.TileAtWorld(0, maxY+1, tileSize); ok {
		t.Fatalf("expected ok == false for TileAtWorld(0,y beyond map)")
	}

	// Very large coordinates should also be treated as out-of-bounds, but
	// must not cause panics or overflow.
	if _, ok := m.TileAtWorld(math.MaxFloat64, math.MaxFloat64, tileSize); ok {
		t.Fatalf("expected ok == false for TileAtWorld at huge coordinates")
	}
}
