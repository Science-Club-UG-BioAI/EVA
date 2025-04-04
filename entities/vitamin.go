package entities

import (
	"projectEVA/animations"
	"projectEVA/components"
)

// import "github.com/hajimehoshi/ebiten/v2"

type VitaminState uint8

const (
	Blue VitaminState = iota
	Red
	Green
	Bronze
)

type Vitamin struct {
	*Sprite
	Speed, Efficiency, TempHP, Duration float64
	StopCalory                          bool
	Type                                int
	CombatComp                          *components.EnemyCombat
	Animations                          map[VitaminState]*animations.Animation
}

func (v *Vitamin) ActiveAnimation(vtype int) *animations.Animation {
	if vtype == 0 {
		return v.Animations[Blue]
	}
	if vtype == 1 {
		return v.Animations[Red]
	}
	if vtype == 2 {
		return v.Animations[Green]
	}
	if vtype == 3 {
		return v.Animations[Bronze]
	} else {
		return nil
	}
}
