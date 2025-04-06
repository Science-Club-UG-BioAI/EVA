package scenes

import "github.com/hajimehoshi/ebiten/v2"

type SceneId uint

// define scenes here
const (
	GameSceneId SceneId = iota
	StartSceneId
	PauseSceneId
	DietSelectionSceneId
	CharacterSelectionSceneId
	ExitSceneId
)

type Scene interface {
	Update() SceneId
	Draw(screen *ebiten.Image)
	FirstLoad()
	OnEnter()
	OnExit()
	IsLoaded() bool
}
