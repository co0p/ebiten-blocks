package mappkg

import (
	"encoding/json"
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
