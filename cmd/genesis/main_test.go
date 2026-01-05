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

func TestRunGeneratesValidJSON(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	out := filepath.Join(dir, "map.json")

	// Create a simple valid 3x2 PNG using grass pixels.
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	grass := color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, grass)
		}
	}
	f, err := os.Create(input)
	if err != nil {
		t.Fatalf("creating input PNG failed: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encoding input PNG failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing input PNG failed: %v", err)
	}

	if err := run([]string{input, out}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if payload["width"].(float64) != 3 || payload["height"].(float64) != 2 {
		t.Fatalf("unexpected dimensions in JSON: %v", payload)
	}

	// Expect a visualization PNG to be generated next to the JSON output.
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
}

func TestVisualizationPathForDerivesPngNextToJson(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "simple filename with json extension",
			in:   "/tmp/map.json",
			out:  "/tmp/map.visual.png",
		},
		{
			name: "relative path",
			in:   "./out/level-01.json",
			out:  "./out/level-01.visual.png",
		},
		{
			name: "no extension",
			in:   "./out/level-01",
			out:  "./out/level-01.visual.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visualizationPathFor(tt.in)
			if got != tt.out {
				t.Fatalf("visualizationPathFor(%q) = %q, want %q", tt.in, got, tt.out)
			}
		})
	}
}

func TestRunIntegrationProducesValidMapAndPng(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	out := filepath.Join(dir, "map.json")

	// Create a small 2x2 PNG using grass pixels so the generator
	// produces a simple map we can validate structurally.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	grass := color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, grass)
		}
	}
	f, err := os.Create(input)
	if err != nil {
		t.Fatalf("creating input PNG failed: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encoding input PNG failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing input PNG failed: %v", err)
	}

	if err := run([]string{input, out}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Decode JSON into a concrete map and check dimensions.
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output failed: %v", err)
	}

	var m mappkg.Map
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("output is not valid map JSON: %v", err)
	}
	if m.Width != 2 || m.Height != 2 {
		t.Fatalf("unexpected map dimensions: got %dx%d, want 2x2", m.Width, m.Height)
	}

	// Decode the visualization PNG and ensure its size matches
	// the expected map dimensions times the tile size used by
	// writeValidationImage (currently 4).
	pngPath := visualizationPathFor(out)
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
	tileW, tileH := w/m.Width, h/m.Height
	if tileW != tileH {
		t.Fatalf("expected square tiles, got tile size %dx%d", tileW, tileH)
	}
}
