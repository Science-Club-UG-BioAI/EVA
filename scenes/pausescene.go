package scenes

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PauseScene struct {
	loaded          bool
	backgroundImage *ebiten.Image
}

func NewPauseScene() *PauseScene {
	return &PauseScene{
		loaded: false,
	}
}

// Draw implements Scene
func (p *PauseScene) Draw(screen *ebiten.Image) {
	if p.backgroundImage != nil {
			op := &ebiten.DrawImageOptions{}
			screen.DrawImage(p.backgroundImage, op)
	}
	ebitenutil.DebugPrint(screen, "Press escape to unpause")
}

// FirstLoad implements Scene.
func (p *PauseScene) FirstLoad() {
	p.loaded = true
	img, _, err := ebitenutil.NewImageFromFile("assets/images/start_background.png")
	if err != nil {
			log.Fatalf("failed to load background image: %v", err)
	}
	p.backgroundImage = img
}

// IsLoaded implements Scene.
func (p *PauseScene) IsLoaded() bool {
	return p.loaded
}

// OnEnter implements Scene.
func (p *PauseScene) OnEnter() {
}

// OnExit implements Scene.
func (p *PauseScene) OnExit() {
}

// Update implements Scene.
func (p *PauseScene) Update() SceneId {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return GameSceneId
	}

	return PauseSceneId
}

var _ Scene = (*PauseScene)(nil)
