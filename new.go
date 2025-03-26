package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"projectEVA/entities"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	player   *entities.Player
	enemies  []*entities.Enemy
	vitamins []*entities.Vitamin
	Cam      *Camera
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.X += 1 + 2*(math.Log(g.player.Speed))

	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.X -= 1 + 2*(math.Log(g.player.Speed))

	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.Y -= 1 + 2*(math.Log(g.player.Speed))

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.Y += 1 + 2*(math.Log(g.player.Speed))
	}

	// enemy behavior
	for _, sprite := range g.enemies {

		if sprite.FollowsPLayer {
			if sprite.X < g.player.X {
				sprite.X += 0.1
			} else if sprite.X > g.player.X {
				sprite.X -= 0.1
			}
			if sprite.Y < g.player.Y {
				sprite.Y += 0.1
			} else if sprite.Y > g.player.Y {
				sprite.Y -= 0.1
			}
		}
	}

	for _, vitamin := range g.vitamins {
		if (g.player.X >= vitamin.X && g.player.X <= vitamin.X+32) && (g.player.Y >= vitamin.Y && g.player.Y <= vitamin.Y+32) {
			g.player.Speed += vitamin.Speed
			g.player.Efficiency *= vitamin.Efficiency
			fmt.Printf("Picked Vitamin")
			vitamin.X = -32
			vitamin.Y = -32
		}
	}

	// g.Cam.FollowTarget(g.player.X+16, g.player.Y+16, 320, 240)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{128, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	// opts.GeoM.Translate(g.Cam.X, g.Cam.Y)

	screen.DrawImage(
		g.player.Img.SubImage(
			image.Rect(0, 0, 32, 32),
		).(*ebiten.Image),
		&opts,
	)

	opts.GeoM.Reset()

	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		// opts.GeoM.Translate(g.Cam.X, g.Cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(256, 448, 320, 512),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
	opts.GeoM.Reset()

	for _, sprite := range g.vitamins {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		// opts.GeoM.Translate(g.Cam.X, g.Cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(0, 0, 32, 32),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
	opts.GeoM.Reset()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("ProjectEVA")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Load Images
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/player.png")
	if err != nil {
		log.Fatal(err)
	}
	vitaminesImg, _, err := ebitenutil.NewImageFromFile("assets/images/vitamines.png")
	if err != nil {
		log.Fatal(err)
	}
	enemiesImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   100,
				Y:   100,
			},
			Speed:      1,
			Efficiency: 1,
			HP:         10,
		},
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: enemiesImg,
					X:   0,
					Y:   0,
				},
				HP:            0,
				FollowsPLayer: false,
			},
		},

		vitamins: []*entities.Vitamin{
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   50,
					Y:   100,
				},
				Speed:      0.5,
				Efficiency: 1.5,
				TempHP:     0,
				Duration:   1,
				StopCalory: false,
			},
		},
		// Cam: NewCamera(0.0, 0.0),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
