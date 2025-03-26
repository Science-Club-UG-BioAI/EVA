package entities

// import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	*Sprite
	Speed, Efficiency, HP float64
}
