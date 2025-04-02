package entities

import "projectEVA/components"

// import "github.com/hajimehoshi/ebiten/v2"

type Enemy struct {
	*Sprite
	Follows    bool
	CombatComp *components.BasicCombat
}

func (e *Enemy) FollowsTarget(target *Sprite) {
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
}
