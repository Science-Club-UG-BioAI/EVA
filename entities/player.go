package entities

import (
	"projectEVA/animations"
	"projectEVA/components"
)

type PlayerState uint8

const (
	W PlayerState = iota
	WD
	D
	DS
	S
	SA
	A
	AW
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
	Diet                 int
	Animations           map[PlayerState]*animations.Animation
	CombatComp           *components.PlayerCombat
}

func (p *Player) ActiveAnimation(dx, dy int) *animations.Animation {
	if dy < 0 && dx == 0 {
		return p.Animations[W]
	}
	if dx > 0 && dy < 0 {
		return p.Animations[WD]
	}
	if dx > 0 && dy == 0 {
		return p.Animations[D]
	}
	if dx > 0 && dy > 0 {
		return p.Animations[DS]
	}
	if dy > 0 && dx == 0 {
		return p.Animations[S]
	}
	if dy > 0 && dx < 0 {
		return p.Animations[SA]
	}
	if dx < 0 && dy == 0 {
		return p.Animations[A]
	}
	if dx < 0 && dy < 0 {
		return p.Animations[AW]
	} else {
		return p.Animations[Idle]
	}
}
