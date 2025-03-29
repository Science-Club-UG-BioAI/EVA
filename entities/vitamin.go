package entities

// import "github.com/hajimehoshi/ebiten/v2"

type Vitamin struct {
	*Sprite
	Speed, Efficiency, TempHP, Duration float64
	StopCalory                          bool
	Type                                int
}
