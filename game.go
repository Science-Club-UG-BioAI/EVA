package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"projectEVA/animations"
	"projectEVA/constants"
	"projectEVA/entities"
	"projectEVA/spritesheet"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	player            *entities.Player
	playerSpriteSheet *spritesheet.SpriteSheet
	enemies           []*entities.Enemy
	vitamins          []*entities.Vitamin
	vitaminDuration   float64
	tilemapJSON       *TilemapJSON
	tilesets          []Tileset
	tilemapImg        *ebiten.Image
	cam               *Camera
	colliders         []image.Rectangle
}

func NewGame() *Game {
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

	tilesets, err := tilemapJSON.GenTilesets()
	if err != nil {
		log.Fatal(err)
	}

	playerSpriteSheet := spritesheet.NewSpriteSheet(30, 1, 32)

	return &Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   (constants.GameWidth / 2) + 16,
				Y:   (constants.GameHeight / 2) + 16,
			},
			Speed:          5,
			Efficiency:     1,
			HP:             10,
			TempSpeed:      1,
			TempEfficiency: 1,
			TempHP:         0,
			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Up:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Down:  animations.NewAnimation(0, 29, 1, 5.0),
				entities.Left:  animations.NewAnimation(0, 29, 1, 5.0),
				entities.Right: animations.NewAnimation(0, 29, 1, 5.0),
				entities.Idle:  animations.NewAnimation(0, 29, 1, 5.0),
			},
		},
		playerSpriteSheet: playerSpriteSheet,
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: enemiesImg,
					X:   (constants.GameWidth / 2),
					Y:   (constants.GameHeight / 2),
				},
				HP:            0,
				FollowsPLayer: true,
			},
		},

		vitamins: []*entities.Vitamin{

			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (constants.GameWidth / 2) + 16 + 100,
					Y:   (constants.GameHeight / 2) + 16 + 200,
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
					X:   (constants.GameWidth / 2) + 16 + 200,
					Y:   (constants.GameHeight / 2) + 16 + 100,
				},
				Speed:      1,
				Efficiency: 1,
				TempHP:     3,
				Duration:   3,
				StopCalory: false,
				Type:       2, // green
			},
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (constants.GameWidth / 2) + 16 + 200,
					Y:   (constants.GameHeight / 2) + 16 + 200,
				},
				Speed:      1,
				Efficiency: 1,
				TempHP:     0,
				Duration:   3,
				StopCalory: true,
				Type:       3, // bronze
			},
		},
		tilemapJSON: tilemapJSON,
		tilemapImg:  tilemapImg,
		tilesets:    tilesets,
		cam:         NewCamera(0.0, 0.0),
		colliders: []image.Rectangle{
			image.Rect(int(constants.GameWidth/2)+300, int(constants.GameHeight/2)+300, int(constants.GameWidth/2)+332, int(constants.GameHeight/2)+332),
		},
	}

}

func (g *Game) Update() error {
	g.player.Dx = 0.0
	g.player.Dy = 0.0
	// Player movement
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.Dx = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.Dx = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.Dy = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.Dy = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.TempSpeed
	}

	g.player.X += g.player.Dx
	CheckCollisionHorizontal(g.player.Sprite, g.colliders)

	g.player.Y += g.player.Dy
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		activeAnim.Update()
	}

	// enemy behavior
	for _, sprite := range g.enemies {

		sprite.Dx = 0.0
		sprite.Dy = 0.0

		if sprite.FollowsPLayer {
			if sprite.X < g.player.X {
				sprite.Dx = 1
			} else if sprite.X > g.player.X {
				sprite.Dx = -1
			}
			if sprite.Y < g.player.Y {
				sprite.Dy = 1
			} else if sprite.Y > g.player.Y {
				sprite.Dy = -1
			}
		}

		sprite.X += sprite.Dx
		CheckCollisionHorizontal(sprite.Sprite, g.colliders)

		sprite.Y += sprite.Dy
		CheckCollisionVertical(sprite.Sprite, g.colliders)
	}

	// vitamin behavior
	for _, vitamin := range g.vitamins {
		// when vitamin picked (probably gonna move to function file in the future)
		// temp 'colision'
		if (g.player.X >= vitamin.X && g.player.X <= vitamin.X+32) && (g.player.Y >= vitamin.Y && g.player.Y <= vitamin.Y+32) {
			fmt.Printf("Picked Vitamin")
			g.player.TempSpeed, g.player.TempHP, g.player.TempEfficiency = 1, 1, 0
			g.player.TempSpeed = vitamin.Speed
			g.player.TempEfficiency = vitamin.Efficiency
			g.player.TempHP = vitamin.TempHP
			if vitamin.StopCalory {
				// Stop calory function
			}
			g.vitaminDuration = vitamin.Duration * 60
		}
	}
	// vitamine countdown
	if vitaminDuration > 0 {
		vitaminDuration--
	} else if vitaminDuration <= 0 {
		g.player.TempSpeed, g.player.TempEfficiency, g.player.TempHP = 1, 1, 0
		vitaminDuration = 0
	}

	// Infinite map illusion
	if g.player.X >= constants.GameWidth-constants.WindowWidth {
		g.player.X = 0 + constants.WindowWidth + 1
	}
	if g.player.X <= 0+constants.WindowWidth {
		g.player.X = constants.GameWidth - constants.WindowWidth - 1
	}
	if g.player.Y >= constants.GameHeight-constants.WindowHeight {
		g.player.Y = 0 + constants.WindowHeight + 1
	}
	if g.player.Y <= 0+constants.WindowHeight {
		g.player.Y = constants.GameHeight - constants.WindowHeight - 1
	}

	g.cam.FollowTarget(g.player.X+16, g.player.Y+16, constants.WindowWidth, constants.WindowHeight)
	g.cam.Constrain(constants.GameWidth, constants.GameHeight, constants.WindowWidth, constants.WindowHeight)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{128, 180, 255, 255})
	opts := ebiten.DrawImageOptions{}

	// loop over layers
	for layerIndex, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {

			if id == 0 {
				continue
			}

			x := index % layer.Width
			y := index / layer.Height

			x *= 32
			y *= 32

			img := g.tilesets[layerIndex].Img(id)

			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy()) + 32))
			opts.GeoM.Translate(g.cam.X, g.cam.Y)

			screen.DrawImage(img, &opts)
			opts.GeoM.Reset()
		}
	}
	opts.GeoM.Reset()

	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(32, 0, 32+32, 32),
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
				image.Rect(32, 32*sprite.Type, 32+32, 32*sprite.Type+32),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
	opts.GeoM.Reset()
	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.cam.X, g.cam.Y)

	playerFrame := 0
	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		playerFrame = activeAnim.Frame()
	}

	screen.DrawImage(
		g.player.Img.SubImage(
			g.playerSpriteSheet.Rect(playerFrame),
		).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()

	for _, colider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(colider.Min.X)+float32(g.cam.X),
			float32(colider.Min.Y)+float32(g.cam.Y),
			float32(colider.Dx()),
			float32(colider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			true,
		)
	}
	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("Player Properties\n Position(%0.0f, %0.0f)\n Speed: %v\n Efficiency: %v\n HP: %v\n TempSpeed: %v\n TempEfficiency: %v\n TempHP: %v\n ",
			g.player.X, g.player.Y, g.player.Speed, g.player.Efficiency, g.player.HP, g.player.TempSpeed, g.player.TempEfficiency, g.player.TempHP))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}
