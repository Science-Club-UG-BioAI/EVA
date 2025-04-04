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
}

var directions = [2]int{-1, 1}

func (e *Enemy) FollowsTarget(target *Sprite) {
	if math.Abs(target.X-e.Sprite.X) < 200 && math.Abs(target.Y-e.Sprite.Y) < 200 {
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
		steps := rand.IntN(100)
		directionX := float64(directions[rand.IntN(len(directions))])
		directionY := float64(directions[rand.IntN(len(directions))])
		for i := 0; i < steps; i++ {
			e.Sprite.Dx = directionX
			e.Sprite.Dy = directionY
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
