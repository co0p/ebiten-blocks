package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	mappkg "github.com/co0p/tankismus/pkg/map"
)

// run is the core entrypoint for the map generator CLI. It is kept
// separate from main to make it easy to test.
func run(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: genesis <seed> <width> <height> <output-path>")
	}

	seedStr := args[0]
	widthStr := args[1]
	heightStr := args[2]
	outputPath := args[3]

	seed, err := strconv.ParseInt(seedStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid seed %q: %w", seedStr, err)
	}
	if seed < 0 || seed > mappkg.MaxSeed {
		return fmt.Errorf("seed must be between 0 and %d", mappkg.MaxSeed)
	}

	width, err := strconv.Atoi(widthStr)
	if err != nil || width <= 0 {
		return fmt.Errorf("invalid width %q", widthStr)
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil || height <= 0 {
		return fmt.Errorf("invalid height %q", heightStr)
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

	world, err := mappkg.NewGrassMap(seed, width, height)
	if err != nil {
		return err
	}

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

	fmt.Printf("generated map %dx%d with seed %d at %s\n", width, height, seed, outputPath)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
