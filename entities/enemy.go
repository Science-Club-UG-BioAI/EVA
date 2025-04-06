package entities

import (
	"math"
	"math/rand/v2"
	"projectEVA/animations"
	"projectEVA/components"
)

// import "github.com/hajimehoshi/ebiten/v2"

type EnemyState uint8

const (
	Meat EnemyState = iota
	Plant
	Agressive
)

type Enemy struct {
	*Sprite
	Follows    bool
	CombatComp *components.EnemyCombat
	Animations map[EnemyState]*animations.Animation
	Type       int
	Speed      float64
}

var directions = [2]int{-1, 1}

func (e *Enemy) FollowsTarget(target *Sprite, vision float64) {
	if math.Abs(target.X-e.Sprite.X) < vision && math.Abs(target.Y-e.Sprite.Y) < vision {
		if e.Sprite.X < target.X {
			e.Sprite.Dx = 1
		} else if e.Sprite.X > target.X {
			e.Sprite.Dx = -1
		}
		if e.Sprite.Y < target.Y {
			e.Sprite.Dy = 1
		} else if e.Sprite.Y > target.Y {
			e.Sprite.Dy = -1
		}
	} else {
		steps := rand.IntN(7)
		if steps == 0 {
			e.Sprite.Dx = (0.1 + 2*(math.Log(1+e.Speed)))
		}
		if steps == 1 {
			e.Sprite.Dx = -(0.1 + 2*(math.Log(1+e.Speed)))
		}
		if steps == 2 {
			e.Sprite.Dy = -(0.1 + 2*(math.Log(1+e.Speed)))
		}
		if steps == 3 {
			e.Sprite.Dy = (0.1 + 2*(math.Log(1+e.Speed)))
		}
		if steps == 4 {
			e.Sprite.Dx = (0.1 + 2*(math.Log(1+e.Speed))) / 1.4
			e.Sprite.Dy = -(0.1 + 2*(math.Log(1+e.Speed))) / 1.4
		}
		if steps == 5 {
			e.Sprite.Dx = -(0.1 + 2*(math.Log(1+e.Speed))) / 1.4
			e.Sprite.Dy = -(0.1 + 2*(math.Log(1+e.Speed))) / 1.4
		}
		if steps == 6 {
			e.Sprite.Dx = (0.1 + 2*(math.Log(1+e.Speed))) / 1.4
			e.Sprite.Dy = (0.1 + 2*(math.Log(1+e.Speed))) / 1.4
		}
		if steps == 7 {
			e.Sprite.Dx = -(0.1 + 2*(math.Log(1+e.Speed))) / 1.4
			e.Sprite.Dy = (0.1 + 2*(math.Log(1+e.Speed))) / 1.4
		}
	}
}
func (e *Enemy) ActiveAnimation(eType int) *animations.Animation {
	if eType == 0 {
		return e.Animations[Meat]
	}
	if eType == 1 {
		return e.Animations[Plant]
	} else {
		return e.Animations[Agressive]
	}
}
