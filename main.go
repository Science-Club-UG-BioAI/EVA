package main

import (
	"log"
	"projectEVA/constants"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(constants.WindowWidth, constants.WindowHeight)
	ebiten.SetWindowTitle("ProjectEVA")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
