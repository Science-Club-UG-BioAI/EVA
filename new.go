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

var windowWidth float64 = 960
var windowHeight float64 = 540
var gameWidth float64 = 3200
var gameHeight float64 = 3200
var vitaminDuration float64 = 0

type Game struct {
	player          *entities.Player
	enemies         []*entities.Enemy
	vitamins        []*entities.Vitamin
	tilemapJSON     *TilemapJSON
	tilemapImg      *ebiten.Image
	cam             *Camera
	animation_frame float64
}

func (g *Game) Update() error {

	// Player movement
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.X += (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.X -= (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.Y -= (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.Y += (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed
	}

	// enemy behavior
	for _, sprite := range g.enemies {

		if sprite.FollowsPLayer {
			if sprite.X+32 < g.player.X+16 {
				sprite.X += 0.1
			} else if sprite.X+32 > g.player.X+16 {
				sprite.X -= 0.1
			}
			if sprite.Y+32 < g.player.Y+16 {
				sprite.Y += 0.1
			} else if sprite.Y+32 > g.player.Y+16 {
				sprite.Y -= 0.1
			}
		}
	}

	// vitamin behavior
	for _, vitamin := range g.vitamins {
		// when vitamin picked (probably gonna move to function file in the future)
		if (g.player.X >= vitamin.X && g.player.X <= vitamin.X+32) && (g.player.Y >= vitamin.Y && g.player.Y <= vitamin.Y+32) {
			fmt.Printf("Picked Vitamin")
			g.player.TempSpeed, g.player.TempHP, g.player.TempEfficiency = 1, 1, 0
			g.player.TempSpeed = vitamin.Speed
			g.player.TempEfficiency = vitamin.Efficiency
			g.player.TempHP = vitamin.TempHP
			if vitamin.StopCalory {
				// Stop calory function
			}
			vitaminDuration = vitamin.Duration * 60
		}
	}
	// vitamine countdown
	if vitaminDuration > 0 {
		vitaminDuration--
	} else if vitaminDuration <= 0 {
		g.player.TempSpeed, g.player.TempHP, g.player.TempEfficiency = 1, 1, 0
		vitaminDuration = 0
	}

	// Infinite map illusion
	if g.player.X >= gameWidth-windowWidth {
		g.player.X = 0 + windowWidth + 1
	}
	if g.player.X <= 0+windowWidth {
		g.player.X = gameWidth - windowWidth - 1
	}
	if g.player.Y >= gameHeight-windowHeight {
		g.player.Y = 0 + windowHeight + 1
	}
	if g.player.Y <= 0+windowHeight {
		g.player.Y = gameHeight - windowHeight - 1
	}

	g.cam.FollowTarget(g.player.X+16, g.player.Y+16, windowWidth, windowHeight)
	g.cam.Constrain(gameWidth, gameHeight, windowWidth, windowHeight)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{128, 180, 255, 255})
	opts := ebiten.DrawImageOptions{}

	// loop over layers
	for _, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {
			x := index % layer.Width
			y := index / layer.Height

			x *= 16
			y *= 16

			srcX := (id - 1) % 12
			srcY := (id - 1) / 12

			srcX *= 16
			srcY *= 16

			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(g.cam.X, g.cam.Y)

			screen.DrawImage(
				g.tilemapImg.SubImage(
					image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image),
				&opts,
			)
			opts.GeoM.Reset()
		}
	}

	if g.animation_frame < 29 {
		g.animation_frame += 0.2
	} else {
		println("X: ", g.player.X)
		println("Y: ", g.player.Y)
		println("Speed: ", g.player.Speed)
		println("Efficiency: ", g.player.Efficiency)
		println("HP: ", g.player.HP)
		g.animation_frame = 0
	}
	opts.GeoM.Reset()

	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

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
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(32*int(g.animation_frame), 32*sprite.Type, 32*int(g.animation_frame)+32, 32*sprite.Type+32),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
	opts.GeoM.Reset()
	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.cam.X, g.cam.Y)

	screen.DrawImage(
		g.player.Img.SubImage(
			image.Rect(32*int(g.animation_frame), 0, 32*int(g.animation_frame)+32, 32),
		).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}

func main() {
	ebiten.SetWindowSize(1920, 1089)
	ebiten.SetWindowTitle("ProjectEVA")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Load Images
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
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
	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/Water+.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("assets/maps/EVAmap.json")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   (gameWidth / 2) + 16,
				Y:   (gameHeight / 2) + 16,
			},
			Speed:          5,
			Efficiency:     1,
			HP:             10,
			TempSpeed:      1,
			TempEfficiency: 1,
			TempHP:         0,
		},
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: enemiesImg,
					X:   windowWidth,
					Y:   windowHeight,
				},
				HP:            0,
				FollowsPLayer: false,
			},
		},

		vitamins: []*entities.Vitamin{
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (gameWidth / 2) + 16 + 100,
					Y:   (gameHeight / 2) + 16 + 100,
				},
				Speed:      0.5,
				Efficiency: 1.5,
				TempHP:     0,
				Duration:   3,
				StopCalory: false,
				Type:       0, // blue
			},
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (gameWidth / 2) + 16 + 100,
					Y:   (gameHeight / 2) + 16 + 200,
				},
				Speed:      1.5,
				Efficiency: 0.5,
				TempHP:     0,
				Duration:   3,
				StopCalory: false,
				Type:       1, // red
			},
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (gameWidth / 2) + 16 + 200,
					Y:   (gameHeight / 2) + 16 + 100,
				},
				Speed:      0,
				Efficiency: 0,
				TempHP:     3,
				Duration:   3,
				StopCalory: false,
				Type:       2, // green
			},
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (gameWidth / 2) + 16 + 200,
					Y:   (gameHeight / 2) + 16 + 200,
				},
				Speed:      1,
				Efficiency: 0,
				TempHP:     0,
				Duration:   3,
				StopCalory: true,
				Type:       3, // bronze
			},
		},
		tilemapJSON:     tilemapJSON,
		tilemapImg:      tilemapImg,
		cam:             NewCamera(0.0, 0.0),
		animation_frame: 0,
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
