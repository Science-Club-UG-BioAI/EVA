package scenes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PauseScene struct {
	loaded bool
}

func NewPauseScene() *PauseScene {
	return &PauseScene{
		loaded: false,
	}
}

// Draw implements Scene.
func (p *PauseScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{200, 200, 50, 1})
	ebitenutil.DebugPrint(screen, "Press escape unpause")
}

// FirstLoad implements Scene.
func (p *PauseScene) FirstLoad() {
	p.loaded = true
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
