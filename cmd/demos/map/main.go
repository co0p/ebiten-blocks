package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/co0p/tankismus/game/scenes/run"
	"github.com/co0p/tankismus/pkg/scene"
)

// sceneGame is a thin Ebiten adapter around the scene manager.
type sceneGame struct {
	manager *scene.Manager
}

func (g *sceneGame) Update() error {
	// Use a fixed timestep for the demo; the underlying scene logic
	// already handles dt in seconds.
	const dt = 1.0 / 60.0
	g.manager.Update(dt)
	return nil
}

func (g *sceneGame) Draw(screen *ebiten.Image) {
	g.manager.Draw(screen)
}

func (g *sceneGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// main starts a minimal Ebiten game that boots directly into the
// run scene, which includes a generated tilemap ground layer with the
// tank rendered above it.
func main() {
	manager := scene.NewManager(nil)
	startScene := run.New(manager)
	manager.SetScene(startScene)

	game := &sceneGame{manager: manager}

	ebiten.SetWindowTitle("tankismus – map demo")
	ebiten.SetWindowSize(800, 600)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
