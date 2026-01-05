package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/co0p/tankismus/game/assets"
	"github.com/co0p/tankismus/game/maps"
	mappkg "github.com/co0p/tankismus/pkg/map"
)

// visualizationPathFor derives the PNG validation image path from the JSON
// output path by keeping the directory and base name and appending a
// ".visual.png" suffix. For example:
//
//	/tmp/map.json        -> /tmp/map.visual.png
//	./out/level-01.json  -> ./out/level-01.visual.png
//	./out/level-01       -> ./out/level-01.visual.png
func visualizationPathFor(jsonPath string) string {
	dir, file := filepath.Split(jsonPath)
	base := strings.TrimSuffix(file, filepath.Ext(file))
	if base == "" {
		base = "map"
	}
	return dir + base + ".visual.png"
}

// writeValidationImage renders a visualization of the generated map into a
// PNG file next to the JSON output. It uses the real tile sprites from the
// assets package so that developers can visually confirm the layout in a way
// that closely matches in-game rendering, without requiring the Ebiten
// rendering pipeline.
func writeValidationImage(m *mappkg.Map, jsonPath string) error {
	if m == nil {
		return nil
	}
	img, err := assets.BuildValidationImage(m)
	if err != nil {
		return fmt.Errorf("failed to build validation image: %w", err)
	}
	if img == nil {
		return nil
	}

	pngPath := visualizationPathFor(jsonPath)
	if err := os.MkdirAll(filepath.Dir(pngPath), 0o755); err != nil {
		return fmt.Errorf("cannot create visualization directory: %w", err)
	}

	tmp := pngPath + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("cannot create visualization PNG file: %w", err)
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		return fmt.Errorf("failed to encode visualization PNG: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing visualization PNG file failed: %w", err)
	}
	if err := os.Rename(tmp, pngPath); err != nil {
		return fmt.Errorf("failed to move visualization PNG into place: %w", err)
	}

	return nil
}

// run is the core entrypoint for the map generator CLI. It is kept
// separate from main to make it easy to test.
func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: genesis <input-png> <output-path>")
	}

	inputPath := args[0]
	outputPath := args[1]

	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("cannot open input image %q: %w", inputPath, err)
	}
	defer inFile.Close()

	img, format, err := image.Decode(inFile)
	if err != nil {
		return fmt.Errorf("failed to decode input image %q: %w", inputPath, err)
	}
	if format != "png" {
		return fmt.Errorf("unsupported input format %q: expected png", format)
	}

	world, err := maps.GenerateFromImage(img)
	if err != nil {
		return fmt.Errorf("failed to generate map from image: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	tmpPath := outputPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(world); err != nil {
		return fmt.Errorf("failed to encode map JSON: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("closing temp file failed: %w", err)
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return fmt.Errorf("failed to move temp file into place: %w", err)
	}

	// After successfully writing the JSON map, generate a simple visualization
	// PNG next to the JSON output so developers can visually inspect the map.
	if err := writeValidationImage(world, outputPath); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
