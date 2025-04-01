package main

import (
	"image"
	"log"
	"projectEVA/entities"

	"github.com/hajimehoshi/ebiten/v2"
)

var windowWidth float64 = 960
var windowHeight float64 = 540
var gameWidth float64 = 4800
var gameHeight float64 = 4800
var vitaminDuration float64 = 0

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(image.Rect(
			int(sprite.X),
			int(sprite.Y),
			int(sprite.X)+32,
			int(sprite.Y)+32)) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - 32
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}
func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(image.Rect(
			int(sprite.X),
			int(sprite.Y),
			int(sprite.X)+32,
			int(sprite.Y)+32)) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - 32
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}

func main() {
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowTitle("ProjectEVA")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
