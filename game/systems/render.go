package systems

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/co0p/tankismus/game/assets"
	"github.com/co0p/tankismus/game/components"
	"github.com/co0p/tankismus/pkg/ecs"
)

type drawable struct {
	entity    ecs.EntityID
	transform *components.Transform
	sprite    *components.Sprite
	z         int
}

// collectDrawables finds all entities with Transform and Sprite components,
// attaches an optional RenderOrder (defaulting to zero when absent), and
// returns them sorted by increasing z (and entity ID as a stable tiebreaker).
func collectDrawables(world *ecs.World) []drawable {
	required := ecs.MaskFor(components.TypeTransform, components.TypeSprite)
	entities := world.Find(required)
	drawables := make([]drawable, 0, len(entities))

	for _, id := range entities {
		cT, okT := world.GetComponent(id, components.TypeTransform)
		cS, okS := world.GetComponent(id, components.TypeSprite)
		if !okT || !okS {
			continue
		}

		p, okP := cT.(*components.Transform)
		s, okSprite := cS.(*components.Sprite)
		if !okP || !okSprite {
			continue
		}

		z := 0
		if cZ, okZ := world.GetComponent(id, components.TypeRenderOrder); okZ {
			if ro, okRO := cZ.(*components.RenderOrder); okRO {
				z = ro.Z
			}
		}

		drawables = append(drawables, drawable{
			entity:    id,
			transform: p,
			sprite:    s,
			z:         z,
		})
	}

	sort.Slice(drawables, func(i, j int) bool {
		if drawables[i].z == drawables[j].z {
			return drawables[i].entity < drawables[j].entity
		}
		return drawables[i].z < drawables[j].z
	})

	return drawables
}

// RenderSystem draws all entities that have a Transform and a Sprite component.
func RenderSystem(world *ecs.World, screen *ebiten.Image) {
	drawables := collectDrawables(world)
	for _, d := range drawables {
		img := assets.GetSprite(d.sprite.SpriteID)
		if img == nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		w, h := img.Size()
		cx := float64(w) / 2
		cy := float64(h) / 2

		// Move origin to sprite center, rotate by logical rotation, then move
		// to world position. Sprites are authored facing +X (to the right).
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Rotate(d.transform.Rotation)
		op.GeoM.Translate(d.transform.X, d.transform.Y)
		screen.DrawImage(img, op)
	}
}
