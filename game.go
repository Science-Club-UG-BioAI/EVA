package main

import (
	"projectEVA/scenes"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	sceneMap      map[scenes.SceneId]scenes.Scene
	activeSceneId scenes.SceneId
}

func NewGame() *Game {
	// List of scenes
	sceneMap := map[scenes.SceneId]scenes.Scene{
		scenes.StartSceneId:        scenes.NewStartScene(),
		scenes.DietSelectionSceneId: scenes.NewDietSelectionScene(),
		scenes.GameSceneId:         scenes.NewGameScene(), // Add GameScene
	}
	activeSceneId := scenes.StartSceneId // Set starting scene
	sceneMap[activeSceneId].FirstLoad()
	return &Game{
		sceneMap,
		activeSceneId,
	}
}

func (g *Game) Update() error {
	nextSceneId := g.sceneMap[g.activeSceneId].Update()
	// switched scenes
	if nextSceneId == scenes.ExitSceneId {
		g.sceneMap[g.activeSceneId].OnExit()
		return ebiten.Termination
	}
	if nextSceneId != g.activeSceneId {
		g.sceneMap[g.activeSceneId].OnExit()
		g.activeSceneId = nextSceneId
		if !g.sceneMap[g.activeSceneId].IsLoaded() {
			g.sceneMap[g.activeSceneId].FirstLoad()
		}
		g.sceneMap[g.activeSceneId].OnEnter()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneMap[g.activeSceneId].Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}
