package entities

import "projectEVA/animations"

// import "github.com/hajimehoshi/ebiten/v2"

type VitaminState uint8

const (
	VitaminIdle VitaminState = iota
)

type Vitamin struct {
	*Sprite
	Speed, Efficiency, TempHP, Duration float64
	StopCalory                          bool
	Type                                int
	Animations                          map[PlayerState]*animations.Animation
}

func (v *Vitamin) ActiveAnimation(dx, dy int) *animations.Animation {
	return v.Animations[PlayerState(VitaminIdle)]
}
