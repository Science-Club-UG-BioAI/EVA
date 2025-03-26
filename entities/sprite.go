package entities

import "github.com/hajimehoshi/ebiten/v2"

type Sprite struct {
	Img  *ebiten.Image
	X, Y float64
}
