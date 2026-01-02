package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesValidJSON(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "map.json")

	if err := run([]string{"123", "3", "2", out}); err != nil {
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

func TestRunRejectsInvalidSeed(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "map.json")

	if err := run([]string{"-1", "3", "2", out}); err == nil {
		t.Fatalf("expected error for invalid seed, got nil")
	}
}

func TestRunRejectsInvalidDimensions(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "map.json")

	if err := run([]string{"1", "0", "2", out}); err == nil {
		t.Fatalf("expected error for invalid width, got nil")
	}

	if err := run([]string{"1", "3", "0", out}); err == nil {
		t.Fatalf("expected error for invalid height, got nil")
	}
}
