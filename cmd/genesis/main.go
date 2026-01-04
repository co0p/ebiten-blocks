package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/co0p/tankismus/game/maps"
)

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
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
