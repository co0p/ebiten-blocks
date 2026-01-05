package assets

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"

	mappkg "github.com/co0p/tankismus/pkg/map"
)

// BuildValidationImage composes a visual representation of the given map
// using the real tile sprites from the embedded assets/images directory.
// The resulting image places each tile sprite into a grid matching the
// logical map layout so that the final PNG closely resembles in-game
// rendering.
func BuildValidationImage(m *mappkg.Map) (image.Image, error) {
	if m == nil {
		return nil, nil
	}

	// Simple cache to avoid decoding the same sprite multiple times
	// while composing a single validation image.
	spriteCache := make(map[string]image.Image)

	loadSprite := func(id string) (image.Image, error) {
		if img, ok := spriteCache[id]; ok {
			return img, nil
		}

		f, err := imagesFS.Open("images/" + id + ".png")
		if err != nil {
			return nil, fmt.Errorf("open sprite %q: %w", id, err)
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("decode sprite %q: %w", id, err)
		}

		spriteCache[id] = img
		return img, nil
	}

	var (
		canvas *image.RGBA
		tileW  int
		tileH  int
	)

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			id, ok := m.TileAt(x, y)
			if !ok || id == "" {
				continue
			}

			src, err := loadSprite(id)
			if err != nil {
				return nil, err
			}

			if canvas == nil {
				b := src.Bounds()
				tileW, tileH = b.Dx(), b.Dy()
				canvas = image.NewRGBA(image.Rect(0, 0, m.Width*tileW, m.Height*tileH))
			}

			srcBounds := src.Bounds()
			dx := x * tileW
			dy := y * tileH
			dstRect := image.Rect(dx, dy, dx+tileW, dy+tileH)
			draw.Draw(canvas, dstRect, src, srcBounds.Min, draw.Src)
		}
	}

	return canvas, nil
}
