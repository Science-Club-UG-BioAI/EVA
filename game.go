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
		scenes.GameSceneId:  scenes.NewGameScene(),
		scenes.StartSceneId: scenes.NewStartScene(),
		scenes.PauseSceneId: scenes.NewPauseScene(),
	}
	activeSceneId := scenes.StartSceneId // Set sterting scene
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
		nextSene := g.sceneMap[nextSceneId]
		// if not loaded then load in
		if !nextSene.IsLoaded() {
			nextSene.FirstLoad()
		}
		nextSene.OnEnter()
		g.sceneMap[g.activeSceneId].OnExit()

	}
	g.activeSceneId = nextSceneId
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneMap[g.activeSceneId].Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}
