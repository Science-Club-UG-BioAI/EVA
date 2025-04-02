package entities

import (
	"projectEVA/animations"
	"projectEVA/components"
)

type PlayerState uint8

const (
	Down PlayerState = iota
	Up
	Left
	Right
	Idle
)

type Player struct {
	*Sprite
	Calories             float64
	Speed                float64
	Efficiency           float64
	SpeedMultiplier      float64
	EfficiencyMultiplier float64
	TempHP               float64
	Animations           map[PlayerState]*animations.Animation
	CombatComp           *components.BasicCombat
}

func (p *Player) ActiveAnimation(dx, dy int) *animations.Animation {
	if dx > 0 {
		return p.Animations[Right]
	}
	if dx < 0 {
		return p.Animations[Left]
	}
	if dy > 0 {
		return p.Animations[Down]
	}
	if dy < 0 {
		return p.Animations[Up]
	} else {
		return p.Animations[Idle]
	}
}
