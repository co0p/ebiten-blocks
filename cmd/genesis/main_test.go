package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
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
}
