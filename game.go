package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"projectEVA/animations"
	"projectEVA/components"
	"projectEVA/constants"
	"projectEVA/entities"
	"projectEVA/spritesheet"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	gamePause         bool
	caloryCount       bool
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
	// Load sptitesheets
	playerSpriteSheet := spritesheet.NewSpriteSheet(30, 1, constants.Tilesize)

	// Init of everything
	return &Game{

		// basic bools
		gamePause:   false,
		caloryCount: true,

		// Create player
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   (constants.GameWidth / 2) + 16,
				Y:   (constants.GameHeight / 2) + 16,
			},
			Calories:             500.00,
			Speed:                5,
			Efficiency:           1,
			SpeedMultiplier:      1,
			EfficiencyMultiplier: 1,
			TempHP:               0,
			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Up:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Down:  animations.NewAnimation(0, 29, 1, 5.0),
				entities.Left:  animations.NewAnimation(0, 29, 1, 5.0),
				entities.Right: animations.NewAnimation(0, 29, 1, 5.0),
				entities.Idle:  animations.NewAnimation(0, 29, 1, 5.0),
			},
			CombatComp: components.NewBasicCombat(3, 1),
		},
		playerSpriteSheet: playerSpriteSheet,

		// Create enemies
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: enemiesImg,
					X:   (constants.GameWidth / 2),
					Y:   (constants.GameHeight / 2),
				},
				Follows:    true,
				CombatComp: components.NewBasicCombat(10, 1),
			},
		},

		// Create vitamis
		vitamins: []*entities.Vitamin{
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (constants.GameWidth / 2) + 16 + 100,
					Y:   (constants.GameHeight / 2) + 16 + 200,
				},
				Speed:      0.5,
				Efficiency: 0.5,
				TempHP:     0,
				Duration:   3,
				StopCalory: false,
				Type:       0, // blue
			},
			{
				Sprite: &entities.Sprite{
					Img: vitaminesImg,
					X:   (constants.GameWidth / 2) + 16 + 100,
					Y:   (constants.GameHeight / 2) + 16 + 200,
				},
				Speed:      1.5,
				Efficiency: 1.5,
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
		vitaminDuration: 0,

		// Create other entities
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.gamePause = !g.gamePause
	}
	if !g.gamePause {
		// Calories
		if g.caloryCount {
			g.player.Calories -= 0.1
		}
		// Player movement
		g.player.Dx = 0.0
		g.player.Dy = 0.0

		if ebiten.IsKeyPressed(ebiten.KeyD) {
			g.player.Dx = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier

		}
		if ebiten.IsKeyPressed(ebiten.KeyA) {
			g.player.Dx = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier

		}
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			g.player.Dy = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier

		}
		if ebiten.IsKeyPressed(ebiten.KeyS) {
			g.player.Dy = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier
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

			if sprite.Follows {
				sprite.FollowsTarget(g.player.Sprite)
			}

			sprite.X += sprite.Dx
			CheckCollisionHorizontal(sprite.Sprite, g.colliders)

			sprite.Y += sprite.Dy
			CheckCollisionVertical(sprite.Sprite, g.colliders)
		}

		clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
		cX, cY := ebiten.CursorPosition()
		cX -= int(g.cam.X)
		cY -= int(g.cam.Y)

		deadEnemies := make(map[int]struct{})
		for index, enemy := range g.enemies {
			rect := image.Rect(
				int(enemy.X),
				int(enemy.Y),
				int(enemy.X)+constants.Tilesize,
				int(enemy.Y)+constants.Tilesize,
			)

			// if g.player.X > float64(rect.Min.X) && g.player.X < float64(rect.Max.X) && g.player.Y > float64(rect.Min.Y) && g.player.Y < float64(rect.Max.Y) {
			if cX > rect.Min.X && cX < rect.Max.X && cY > rect.Min.Y && cY < rect.Max.Y {
				if clicked {
					enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())
				}
				if enemy.CombatComp.Health() <= 0 {
					deadEnemies[index] = struct{}{}
				}
			}
		}
		if len(deadEnemies) > 0 {
			newEnemies := make([]*entities.Enemy, 0)
			for index, enemy := range g.enemies {
				if _, exists := deadEnemies[index]; !exists {
					newEnemies = append(newEnemies, enemy)
				}
			}
			g.enemies = newEnemies
		}

		// vitamin behavior
		for _, vitamin := range g.vitamins {
			// when vitamin picked (probably gonna move to function file in the future)
			// temp 'colision'
			if (g.player.X >= vitamin.X && g.player.X <= vitamin.X+constants.Tilesize) && (g.player.Y >= vitamin.Y && g.player.Y <= vitamin.Y+constants.Tilesize) {
				fmt.Printf("Picked Vitamin")
				g.player.SpeedMultiplier, g.player.TempHP, g.player.EfficiencyMultiplier = 1, 1, 0
				g.player.SpeedMultiplier = vitamin.Speed
				g.player.EfficiencyMultiplier = vitamin.Efficiency
				g.player.TempHP = vitamin.TempHP
				if vitamin.StopCalory {
					g.caloryCount = false
				}
				g.vitaminDuration = vitamin.Duration * 60
			}
		}
		// vitamine countdown
		if g.vitaminDuration > 0 {
			g.vitaminDuration--
		} else if g.vitaminDuration <= 0 {
			g.caloryCount = true
			g.player.SpeedMultiplier = 1
			g.player.EfficiencyMultiplier = 1
			g.player.TempHP = 0
			g.vitaminDuration = 0
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
	}
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

			x *= constants.Tilesize
			y *= constants.Tilesize

			img := g.tilesets[layerIndex].Img(id)

			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy()) + constants.Tilesize))
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
				image.Rect(constants.Tilesize, 0, constants.Tilesize+constants.Tilesize, constants.Tilesize),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("EnemyHP:  %v", sprite.CombatComp.Health()), int(sprite.X)+int(g.cam.X), int(sprite.Y)+int(g.cam.Y))
	}
	opts.GeoM.Reset()

	for _, sprite := range g.vitamins {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(constants.Tilesize, constants.Tilesize*sprite.Type, constants.Tilesize+constants.Tilesize, constants.Tilesize*sprite.Type+constants.Tilesize),
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
		fmt.Sprintf("Player Properties: \n Position(%0.1f, %0.1f)\n Calories: %0.0f/1000\n Speed: %0.1f\n Efficiency: %0.1f\n HP: %0.1f\n SpeedMultiplier: %0.1f\n EfficiencyMultiplier: %0.1f\n TempHP: %0.1f\n Vitamin Duration: %0.1f",
			g.player.X, g.player.Y, g.player.Calories, g.player.Speed, g.player.Efficiency, g.player.CombatComp.Health(), g.player.SpeedMultiplier, g.player.EfficiencyMultiplier, g.player.TempHP, g.vitaminDuration))
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Game State: \n Game Pause: %v", g.gamePause), 0, 300)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}
