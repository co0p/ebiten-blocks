package mappkg

import (
	"errors"
	"math/rand"
)

// Map represents a tile map as described in the seeded grass map
// generator design. It is intentionally independent of Ebiten and
// game-specific concepts.

type Map struct {
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Seed   int64      `json:"seed"`
	Tiles  [][]string `json:"tiles"`
}

var (
	ErrInvalidDimensions  = errors.New("map: width and height must be positive")
	ErrInvalidTilesHeight = errors.New("map: tiles height does not match map height")
	ErrInvalidTilesWidth  = errors.New("map: tiles width does not match map width")
	ErrInvalidSeed        = errors.New("map: seed must be between 0 and 9999999999")
)

// MaxSeed defines the largest allowed seed value (10 decimal digits).
const MaxSeed int64 = 9999999999

// validate checks basic structural invariants on the Map.
func (m *Map) validate() error {
	if m.Width <= 0 || m.Height <= 0 {
		return ErrInvalidDimensions
	}
	if m.Seed < 0 || m.Seed > MaxSeed {
		return ErrInvalidSeed
	}
	if len(m.Tiles) != m.Height {
		return ErrInvalidTilesHeight
	}
	for _, row := range m.Tiles {
		if len(row) != m.Width {
			return ErrInvalidTilesWidth
		}
	}
	return nil
}

// TileAt returns the tile ID at the given tile coordinates (x, y).
//
// The coordinate system is zero-based with (0,0) at the top-left
// corner of the map, x increasing to the right and y increasing
// downward. If the coordinates are out of bounds, TileAt returns
// ("", false).
func (m *Map) TileAt(x, y int) (string, bool) {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return "", false
	}
	if len(m.Tiles) != m.Height {
		return "", false
	}
	row := m.Tiles[y]
	if len(row) != m.Width {
		return "", false
	}
	return row[x], true
}

// TileAtWorld maps world-space coordinates (worldX, worldY) to the
// corresponding tile and returns its tile ID.
//
// It assumes a uniform square tile size given by tileSize, with the
// top-left corner of the map located at world coordinate (0,0).
// World coordinates are interpreted in the same unit system as the
// tile size (for example, pixels). Coordinates that fall outside the
// map's rectangular extent result in ("", false).
func (m *Map) TileAtWorld(worldX, worldY, tileSize float64) (string, bool) {
	if tileSize <= 0 {
		return "", false
	}
	if worldX < 0 || worldY < 0 {
		return "", false
	}
	tileX := int(worldX / tileSize)
	tileY := int(worldY / tileSize)
	return m.TileAt(tileX, tileY)
}

// NewGrassMap constructs a grass-only map using tileGrass1 and
// tileGrass2, with deterministic layout based on the given seed
// and dimensions.
func NewGrassMap(seed int64, width, height int) (*Map, error) {
	m := &Map{
		Width:  width,
		Height: height,
		Seed:   seed,
	}
	rng := rand.New(rand.NewSource(seed))
	tiles := make([][]string, height)
	for y := 0; y < height; y++ {
		row := make([]string, width)
		for x := 0; x < width; x++ {
			if rng.Intn(2) == 0 {
				row[x] = "tileGrass1"
			} else {
				row[x] = "tileGrass2"
			}
		}
		tiles[y] = row
	}

	m.Tiles = tiles
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}
