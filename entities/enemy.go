package entities

// import "github.com/hajimehoshi/ebiten/v2"

type Enemy struct {
	*Sprite
	HP            float64
	FollowsPLayer bool
}
