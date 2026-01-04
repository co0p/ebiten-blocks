package maps

import (
	"fmt"
	"image"
	"image/color"

	mappkg "github.com/co0p/tankismus/pkg/map"
)

// TerrainKind represents the logical terrain derived from pixel colors.
type TerrainKind int

const (
	terrainUnknown TerrainKind = iota
	terrainGrass
	terrainSand
	terrainRoad
)

// Fixed RGB values for terrain kinds (see design.md Color Mapping Contract).
var (
	grassRGB = color.RGBA{0x00, 0xFF, 0x00, 0xFF} // #00FF00
	sandRGB  = color.RGBA{0xFF, 0xFF, 0x00, 0xFF} // #FFFF00
	roadRGB  = color.RGBA{0x00, 0x00, 0x00, 0xFF} // #000000
)

// colorToTerrain maps an exact RGBA color to a TerrainKind.
// Any non-matching color results in an error.
func colorToTerrain(c color.Color) (TerrainKind, error) {
	r, g, b, a := c.RGBA()
	// Normalize down to 8-bit per channel.
	cr := uint8(r >> 8)
	cg := uint8(g >> 8)
	cb := uint8(b >> 8)
	ca := uint8(a >> 8)

	candidate := color.RGBA{cr, cg, cb, ca}

	switch candidate {
	case grassRGB:
		return terrainGrass, nil
	case sandRGB:
		return terrainSand, nil
	case roadRGB:
		return terrainRoad, nil
	default:
		return terrainUnknown, fmt.Errorf("unsupported color: %#v", candidate)
	}
}

// neighborMaskForKind computes a bitmask of neighbors of a given terrain kind
// for the cell at (x, y). Bits: 1 = north, 2 = east, 4 = south, 8 = west.
func neighborMaskForKind(grid [][]TerrainKind, x, y int, kind TerrainKind) uint8 {
	h := len(grid)
	if h == 0 {
		return 0
	}
	w := len(grid[0])
	mask := uint8(0)

	if y-1 >= 0 && grid[y-1][x] == kind {
		mask |= 1 // N
	}
	if x+1 < w && grid[y][x+1] == kind {
		mask |= 2 // E
	}
	if y+1 < h && grid[y+1][x] == kind {
		mask |= 4 // S
	}
	if x-1 >= 0 && grid[y][x-1] == kind {
		mask |= 8 // W
	}
	return mask
}

// selectRoadTileForStraight chooses a road tile ID for straight segments
// based on the neighbor mask. For now we only distinguish vertical vs
// horizontal; other patterns fall back to a generic vertical road tile.
func selectRoadTileForStraight(mask uint8) string {
	// Vertical: has N or S, but no E/W.
	vertical := (mask&(1|4) != 0) && (mask&(2|8) == 0)
	// Horizontal: has E or W, but no N/S.
	horizontal := (mask&(2|8) != 0) && (mask&(1|4) == 0)

	if horizontal {
		return "tileGrass_roadEast"
	}
	// Default and vertical both use the North-oriented straight tile for now.
	_ = vertical
	return "tileGrass_roadNorth"
}

// selectRoadTile chooses a road tile based on the full neighbor mask,
// handling straight segments, corners, T-junctions, and crossings.
func selectRoadTile(mask uint8) string {
	// Four-way crossing.
	if mask == 1|2|4|8 {
		return "tileGrass_roadCrossing"
	}

	// T-junctions (three neighbors).
	switch mask {
	case 1 | 2 | 8: // N, E, W
		return "tileGrass_roadSplitN"
	case 2 | 4 | 8: // E, S, W
		return "tileGrass_roadSplitS"
	case 1 | 2 | 4: // N, E, S
		return "tileGrass_roadSplitE"
	case 1 | 4 | 8: // N, S, W
		return "tileGrass_roadSplitW"
	}

	// Corners (two orthogonal neighbors).
	switch mask {
	case 1 | 2: // N+E
		return "tileGrass_roadCornerUR"
	case 2 | 4: // E+S
		return "tileGrass_roadCornerLR"
	case 4 | 8: // S+W
		return "tileGrass_roadCornerLL"
	case 1 | 8: // N+W
		return "tileGrass_roadCornerUL"
	}

	// Straight segments and misc fall back to the straight helper.
	return selectRoadTileForStraight(mask)
}

// GenerateFromImage converts an image into a *mappkg.Map using fixed RGB
// color mappings and simple tile IDs. Straight road segments are classified
// as vertical vs horizontal based on neighbor road presence. More complex
// neighbor-aware logic (corners, junctions, transitions) is handled in
// later steps of this increment.
func GenerateFromImage(img image.Image) (*mappkg.Map, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image must have positive dimensions, got %dx%d", width, height)
	}

	// First, build a terrain grid and fail fast on any unsupported color.
	terrain := make([][]TerrainKind, height)
	for y := 0; y < height; y++ {
		row := make([]TerrainKind, width)
		for x := 0; x < width; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			kind, err := colorToTerrain(c)
			if err != nil {
				return nil, err
			}
			row[x] = kind
		}
		terrain[y] = row
	}

	// Construct the map and assign tile IDs per terrain kind.
	m := &mappkg.Map{
		Width:  width,
		Height: height,
		Seed:   0,
		Tiles:  make([][]string, height),
	}

	for y := 0; y < height; y++ {
		row := make([]string, width)
		for x := 0; x < width; x++ {
			switch terrain[y][x] {
			case terrainGrass:
				// Grass–sand boundary transitions.
				mask := neighborMaskForKind(terrain, x, y, terrainSand)
				if mask != 0 {
					// Prefer N/S over E/W when multiple neighbors exist.
					if mask&1 != 0 {
						row[x] = "tileGrass_transitionN"
					} else if mask&4 != 0 {
						row[x] = "tileGrass_transitionS"
					} else if mask&2 != 0 {
						row[x] = "tileGrass_transitionE"
					} else {
						row[x] = "tileGrass_transitionW"
					}
				} else {
					row[x] = "tileGrass1"
				}
			case terrainSand:
				row[x] = "tileSand1"
			case terrainRoad:
				mask := neighborMaskForKind(terrain, x, y, terrainRoad)
				row[x] = selectRoadTile(mask)
			default:
				// Should not occur because colorToTerrain rejects unknowns.
				return nil, fmt.Errorf("internal error: unknown terrain at (%d,%d)", x, y)
			}
		}
		m.Tiles[y] = row
	}

	if err := m.ValidateForGenerator(); err != nil {
		return nil, err
	}

	return m, nil
}
