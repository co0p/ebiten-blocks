package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	mappkg "github.com/co0p/tankismus/pkg/map"
)

// TestEndToEndSamplePixelmap exercises the full genesis CLI pipeline on a
// slightly more interesting input image that includes grass, sand and road
// pixels. It asserts on the resulting map JSON and performs structural
// checks on the visualization PNG so that regressions in the wiring are
// caught early.
func TestEndToEndSamplePixelmap(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "sample.png")
	out := filepath.Join(dir, "sample.json")

	// Construct a 3x3 image with:
	// - top row: grass, grass, grass
	// - middle row: road, road, road
	// - bottom row: sand, sand, sand
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	grass := color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	sand := color.RGBA{0xFF, 0xFF, 0x00, 0xFF}
	road := color.RGBA{0x00, 0x00, 0x00, 0xFF}

	for x := 0; x < 3; x++ {
		img.Set(x, 0, grass)
		img.Set(x, 1, road)
		img.Set(x, 2, sand)
	}

	f, err := os.Create(input)
	if err != nil {
		t.Fatalf("creating sample input PNG failed: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encoding sample input PNG failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing sample input PNG failed: %v", err)
	}

	if err := run([]string{input, out}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output JSON failed: %v", err)
	}

	var m mappkg.Map
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("output is not valid map JSON: %v", err)
	}
	if m.Width != 3 || m.Height != 3 {
		t.Fatalf("unexpected map dimensions: got %dx%d, want 3x3", m.Width, m.Height)
	}

	// The middle row should consist of road tiles chosen by the generator.
	for x := 0; x < 3; x++ {
		id := m.Tiles[1][x]
		if id == "" {
			t.Fatalf("expected road tile at (x=%d,y=1), got empty ID", x)
		}
		if id != "tileGrass_roadEast" && id != "tileGrass_roadNorth" && id != "tileGrass_roadCrossing" && id != "tileGrass_roadCrossingRound" && id != "tileGrass_roadSplitE" && id != "tileGrass_roadSplitN" && id != "tileGrass_roadSplitS" && id != "tileGrass_roadSplitW" {
			t.Fatalf("expected a road-related tile ID at (x=%d,y=1), got %q", x, id)
		}
	}

	// Top row tiles should be grass or grass-transition variants.
	for x := 0; x < 3; x++ {
		id := m.Tiles[0][x]
		if id == "" {
			t.Fatalf("expected grass tile at (x=%d,y=0), got empty ID", x)
		}
		if id[:4] != "tile" {
			t.Fatalf("expected tile-prefixed ID at (x=%d,y=0), got %q", x, id)
		}
	}

	// Bottom row tiles should be sand variants.
	for x := 0; x < 3; x++ {
		id := m.Tiles[2][x]
		if id == "" {
			t.Fatalf("expected sand tile at (x=%d,y=2), got empty ID", x)
		}
		if id[:8] != "tileSand" {
			t.Fatalf("expected sand-related tile ID at (x=%d,y=2), got %q", x, id)
		}
	}

	// Finally, ensure the visualization PNG exists and is structurally sound.
	pngPath := visualizationPathFor(out)
	info, err := os.Stat(pngPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("expected visualization PNG at %q to exist", pngPath)
		}
		t.Fatalf("stat visualization PNG failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("visualization PNG at %q is empty", pngPath)
	}

	pngFile, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("opening visualization PNG failed: %v", err)
	}
	defer pngFile.Close()

	outImg, format, err := image.Decode(pngFile)
	if err != nil {
		t.Fatalf("decoding visualization PNG failed: %v", err)
	}
	if format != "png" {
		t.Fatalf("unexpected visualization format: %q", format)
	}

	w, h := outImg.Bounds().Dx(), outImg.Bounds().Dy()
	if w <= 0 || h <= 0 {
		t.Fatalf("visualization image has non-positive dimensions: (%d,%d)", w, h)
	}
	if w%m.Width != 0 || h%m.Height != 0 {
		t.Fatalf("visualization size (%d,%d) is not an integer multiple of map dimensions (%d,%d)", w, h, m.Width, m.Height)
	}
}
