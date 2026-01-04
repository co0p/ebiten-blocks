package maps

import (
	"image"
	"image/color"
	"testing"
)

// helper to build a solid-color RGBA image of given size
func newSolidImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

var (
	grassColor = color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	sandColor  = color.RGBA{0xFF, 0xFF, 0x00, 0xFF}
	roadColor  = color.RGBA{0x00, 0x00, 0x00, 0xFF}
)

func TestGenerateFromImage_AllGrassSandRoad(t *testing.T) {
	tests := []struct {
		name string
		c    color.Color
	}{
		{"grass", grassColor},
		{"sand", sandColor},
		{"road", roadColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := newSolidImage(2, 2, tt.c)

			m, err := GenerateFromImage(img)
			if err != nil {
				t.Fatalf("GenerateFromImage returned error: %v", err)
			}
			if m == nil {
				t.Fatalf("expected non-nil map")
			}

			if m.Width != 2 || m.Height != 2 {
				t.Fatalf("unexpected dimensions: got %dx%d", m.Width, m.Height)
			}

			if len(m.Tiles) != m.Height {
				t.Fatalf("tiles height mismatch: got %d want %d", len(m.Tiles), m.Height)
			}
			for y := 0; y < m.Height; y++ {
				if len(m.Tiles[y]) != m.Width {
					t.Fatalf("tiles width mismatch in row %d: got %d want %d", y, len(m.Tiles[y]), m.Width)
				}
				for x := 0; x < m.Width; x++ {
					if m.Tiles[y][x] == "" {
						t.Fatalf("empty tile ID at (%d,%d)", x, y)
					}
				}
			}
		})
	}
}

func TestGenerateFromImage_UnsupportedColor(t *testing.T) {
	img := newSolidImage(1, 1, color.RGBA{0x12, 0x34, 0x56, 0xFF})

	m, err := GenerateFromImage(img)
	if err == nil {
		t.Fatalf("expected error for unsupported color, got nil and map: %#v", m)
	}
}

func TestGenerateFromImage_VerticalAndHorizontalRoad(t *testing.T) {
	// 3x3 with a vertical road in the middle column, grass elsewhere.
	imgVert := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grassColor
			if x == 1 { // middle column is road
				c = roadColor
			}
			imgVert.Set(x, y, c)
		}
	}

	mVert, err := GenerateFromImage(imgVert)
	if err != nil {
		t.Fatalf("GenerateFromImage (vertical) returned error: %v", err)
	}
	if got := mVert.Tiles[1][1]; got != "tileGrass_roadNorth" {
		t.Fatalf("expected vertical road tile 'tileGrass_roadNorth' at center, got %q", got)
	}

	// 3x3 with a horizontal road in the middle row, grass elsewhere.
	imgHoriz := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grassColor
			if y == 1 { // middle row is road
				c = roadColor
			}
			imgHoriz.Set(x, y, c)
		}
	}

	mHoriz, err := GenerateFromImage(imgHoriz)
	if err != nil {
		t.Fatalf("GenerateFromImage (horizontal) returned error: %v", err)
	}
	if got := mHoriz.Tiles[1][1]; got != "tileGrass_roadEast" {
		t.Fatalf("expected horizontal road tile 'tileGrass_roadEast' at center, got %q", got)
	}
}

func TestGenerateFromImage_RoadCornersTJunctionsCrossing(t *testing.T) {
	// Corner: N+E neighbors -> UR corner.
	imgCorner := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grassColor
			if (x == 1 && y <= 1) || (y == 1 && x >= 1) { // vertical up and horizontal right from center
				c = roadColor
			}
			imgCorner.Set(x, y, c)
		}
	}
	mCorner, err := GenerateFromImage(imgCorner)
	if err != nil {
		t.Fatalf("GenerateFromImage (corner) returned error: %v", err)
	}
	if got := mCorner.Tiles[1][1]; got != "tileGrass_roadCornerUR" {
		t.Fatalf("expected corner tile 'tileGrass_roadCornerUR' at center, got %q", got)
	}

	// T-junction: N, E, W neighbors -> SplitN.
	imgT := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grassColor
			if (x == 1 && y == 0) || // N
				(x == 0 && y == 1) || // W
				(x == 1 && y == 1) || // center
				(x == 2 && y == 1) { // E
				c = roadColor
			}
			imgT.Set(x, y, c)
		}
	}
	mT, err := GenerateFromImage(imgT)
	if err != nil {
		t.Fatalf("GenerateFromImage (T) returned error: %v", err)
	}
	if got := mT.Tiles[1][1]; got != "tileGrass_roadSplitN" {
		t.Fatalf("expected T-junction tile 'tileGrass_roadSplitN' at center, got %q", got)
	}

	// Crossing: N, E, S, W neighbors -> Crossing.
	imgCross := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := grassColor
			if (x == 1 && y != 1) || (y == 1 && x != 1) || (x == 1 && y == 1) {
				c = roadColor
			}
			imgCross.Set(x, y, c)
		}
	}
	mCross, err := GenerateFromImage(imgCross)
	if err != nil {
		t.Fatalf("GenerateFromImage (crossing) returned error: %v", err)
	}
	if got := mCross.Tiles[1][1]; got != "tileGrass_roadCrossing" {
		t.Fatalf("expected crossing tile 'tileGrass_roadCrossing' at center, got %q", got)
	}
}

func TestGenerateFromImage_GrassSandTransitions(t *testing.T) {
	// Horizontal boundary: sand on top row, grass on bottom row -> grass transition north.
	imgH := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			c := grassColor
			if y == 0 {
				c = sandColor
			}
			imgH.Set(x, y, c)
		}
	}
	mH, err := GenerateFromImage(imgH)
	if err != nil {
		t.Fatalf("GenerateFromImage (horizontal transition) returned error: %v", err)
	}
	if got := mH.Tiles[1][1]; got != "tileGrass_transitionN" {
		t.Fatalf("expected grass transition 'tileGrass_transitionN' at (1,1), got %q", got)
	}

	// Vertical boundary: sand on left column, grass on right column -> grass transition west.
	imgV := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			c := grassColor
			if x == 0 {
				c = sandColor
			}
			imgV.Set(x, y, c)
		}
	}
	mV, err := GenerateFromImage(imgV)
	if err != nil {
		t.Fatalf("GenerateFromImage (vertical transition) returned error: %v", err)
	}
	if got := mV.Tiles[1][1]; got != "tileGrass_transitionW" {
		t.Fatalf("expected grass transition 'tileGrass_transitionW' at (1,1), got %q", got)
	}
}
