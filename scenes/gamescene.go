package scenes

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"projectEVA/animations"
	"projectEVA/camera"
	"projectEVA/components"
	"projectEVA/constants"
	"projectEVA/entities"
	"projectEVA/spritesheet"
	"projectEVA/tilemap"
	"projectEVA/tileset"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type GameScene struct {
	loaded             bool
	gamePause          bool
	gameOver           bool
	caloryCount        bool
	player             *entities.Player
	playerSpriteSheet  *spritesheet.SpriteSheet
	enemies            []*entities.Enemy
	enemySpriteSheet   *spritesheet.SpriteSheet
	vitamins           []*entities.Vitamin
	vitaminSpriteSheet *spritesheet.SpriteSheet
	vitaminDuration    float64
	tilemapJSON        *tilemap.TilemapJSON
	tilesets           []tileset.Tileset
	tilemapImg         *ebiten.Image
	cam                *camera.Camera
	colliders          []image.Rectangle
}

func NewGameScene() *GameScene {
	return &GameScene{
		gamePause:          false,
		gameOver:           false,
		caloryCount:        true,
		player:             nil,
		playerSpriteSheet:  nil,
		enemies:            make([]*entities.Enemy, 0),
		enemySpriteSheet:   nil,
		vitamins:           make([]*entities.Vitamin, 0),
		vitaminSpriteSheet: nil,
		vitaminDuration:    0,
		tilemapJSON:        nil,
		tilesets:           nil,
		tilemapImg:         nil,
		cam:                nil,
		colliders:          make([]image.Rectangle, 0),
		loaded:             false,
	}
}

func (g *GameScene) IsLoaded() bool {
	return g.loaded
}

func (g *GameScene) Draw(screen *ebiten.Image) {

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

	for _, sprite := range g.enemies {
		opts.GeoM.Reset()
		if sprite.Type != 2 {
			opts.GeoM.Scale(sprite.Size, sprite.Size)
		}
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		enemyFrame := 0
		activeAnim := sprite.ActiveAnimation(sprite.Type)
		if activeAnim != nil {
			enemyFrame = activeAnim.Frame()
		}
		screen.DrawImage(
			sprite.Img.SubImage(
				g.enemySpriteSheet.Rect(enemyFrame),
			).(*ebiten.Image),
			&opts,
		)
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("EnemyHP:  %v", sprite.CombatComp.Health()), int(sprite.X)+int(g.cam.X), int(sprite.Y)+int(g.cam.Y))
		opts.GeoM.Reset()
	}
	opts.GeoM.Reset()

	for _, sprite := range g.vitamins {
		opts.GeoM.Reset()
		opts.GeoM.Scale(sprite.Size, sprite.Size)
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		vitaminFrame := 0
		activeAnim := sprite.ActiveAnimation(sprite.Type)
		if activeAnim != nil {
			vitaminFrame = activeAnim.Frame()
		}
		screen.DrawImage(
			sprite.Img.SubImage(
				g.vitaminSpriteSheet.Rect(vitaminFrame),
			).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
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
	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("Player Properties: \n Position(%0.1f, %0.1f)\n Calories: %0.0f/1000\n Diet: %v\n Speed: %0.1f\n Efficiency: %0.1f\n HP: %0.1f\n SpeedMultiplier: %0.1f\n EfficiencyMultiplier: %0.1f\n TempHP: %0.1f\n Vitamin Duration: %0.1f",
			g.player.X, g.player.Y, g.player.Calories, g.player.Diet, g.player.Speed, g.player.Efficiency, g.player.CombatComp.Health(), g.player.SpeedMultiplier, g.player.EfficiencyMultiplier, g.player.TempHP, g.vitaminDuration))
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Game State: \n Game Pause: %v\n Game Over: %v", g.gamePause, g.gameOver), 0, 300)
	if g.gameOver {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("GAME OVER\n"), 480, 270)
	}

}

func (g *GameScene) FirstLoad() {
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
	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/Water+.png")
	if err != nil {
		log.Fatal(err)
	}
	tilemapJSON, err := tilemap.NewTilemapJSON("assets/maps/EVAmap.json")
	if err != nil {
		log.Fatal(err)
	}
	tilesets, err := tilemapJSON.GenTilesets()
	if err != nil {
		log.Fatal(err)
	}
	// Load sptitesheets
	playerSpriteSheet := spritesheet.NewSpriteSheet(30, 1, constants.Tilesize)
	vitaminSpriteSheet := spritesheet.NewSpriteSheet(30, 1, constants.Tilesize)
	enemySpriteSheet := spritesheet.NewSpriteSheet(30, 1, constants.Tilesize)

	g.player = &entities.Player{
		Sprite: &entities.Sprite{
			Img:  playerImg,
			X:    (constants.GameWidth / 2) + 16,
			Y:    (constants.GameHeight / 2) + 16,
			Size: 1,
		},
		Calories:             500.00,
		Speed:                5,
		Efficiency:           1,
		SpeedMultiplier:      1,
		EfficiencyMultiplier: 1,
		TempHP:               0,
		Size:                 1,
		Animations: map[entities.PlayerState]*animations.Animation{
			// entities.W:    animations.NewAnimation(0, 29, 1, 5.0),
			// entities.WD:   animations.NewAnimation(30, 59, 1, 5.0),
			// entities.D:    animations.NewAnimation(60, 89, 1, 5.0),
			// entities.DS:   animations.NewAnimation(90, 119, 1, 5.0),
			// entities.S:    animations.NewAnimation(120, 149, 1, 5.0),
			// entities.SA:   animations.NewAnimation(150, 179, 1, 5.0),
			// entities.A:    animations.NewAnimation(180, 209, 1, 5.0),
			// entities.AW:   animations.NewAnimation(210, 239, 1, 5.0),
			// entities.Idle: animations.NewAnimation(240, 269, 1, 5.0),
			entities.W:    animations.NewAnimation(0, 0, 1, 5.0),
			entities.WD:   animations.NewAnimation(30, 30, 1, 5.0),
			entities.D:    animations.NewAnimation(60, 60, 1, 5.0),
			entities.DS:   animations.NewAnimation(90, 90, 1, 5.0),
			entities.S:    animations.NewAnimation(120, 120, 1, 5.0),
			entities.SA:   animations.NewAnimation(150, 150, 1, 5.0),
			entities.A:    animations.NewAnimation(180, 180, 1, 5.0),
			entities.AW:   animations.NewAnimation(210, 210, 1, 5.0),
			entities.Idle: animations.NewAnimation(240, 240, 1, 5.0),
		},
		CombatComp: components.NewPlayerCombat(3, 1, 1000),
		Diet:       0,
	}
	g.playerSpriteSheet = playerSpriteSheet

	g.enemies = []*entities.Enemy{
		{
			Sprite: &entities.Sprite{
				Img:  enemiesImg,
				X:    (constants.GameWidth / 2) - 500,
				Y:    (constants.GameHeight / 2) - 500,
				Size: 1,
			},
			Animations: map[entities.EnemyState]*animations.Animation{
				entities.Meat:      animations.NewAnimation(0, 29, 1, 5.0),
				entities.Plant:     animations.NewAnimation(30, 59, 1, 5.0),
				entities.Agressive: animations.NewAnimation(60, 89, 1, 5.0),
			},
			Follows:    true,
			CombatComp: components.NewEnemyCombat(10, 1, 1000),
			Type:       2,
		},
	}
	g.enemySpriteSheet = enemySpriteSheet
	g.vitamins = []*entities.Vitamin{
		{
			Sprite: &entities.Sprite{
				Img:  vitaminesImg,
				X:    (constants.GameWidth / 2) + 16 + 100,
				Y:    (constants.GameHeight / 2) + 16 + 100,
				Size: constants.VitaminSize,
			},
			Animations: map[entities.VitaminState]*animations.Animation{
				entities.Blue:   animations.NewAnimation(0, 29, 1, 5.0),
				entities.Red:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Green:  animations.NewAnimation(60, 89, 1, 5.0),
				entities.Bronze: animations.NewAnimation(90, 119, 1, 5.0),
			},
			CombatComp: components.NewEnemyCombat(1, 0, 0),
			Speed:      0.5,
			Efficiency: 0.5,
			TempHP:     0,
			Duration:   3,
			StopCalory: false,
			Type:       0, // blue
		},
		{
			Sprite: &entities.Sprite{
				Img:  vitaminesImg,
				X:    (constants.GameWidth / 2) + 16 + 100,
				Y:    (constants.GameHeight / 2) + 16 + 200,
				Size: constants.VitaminSize,
			},
			Animations: map[entities.VitaminState]*animations.Animation{
				entities.Blue:   animations.NewAnimation(0, 29, 1, 5.0),
				entities.Red:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Green:  animations.NewAnimation(60, 89, 1, 5.0),
				entities.Bronze: animations.NewAnimation(90, 119, 1, 5.0),
			},
			CombatComp: components.NewEnemyCombat(1, 0, 0),
			Speed:      1.5,
			Efficiency: 1.5,
			TempHP:     0,
			Duration:   3,
			StopCalory: false,
			Type:       1, // red
		},
		{
			Sprite: &entities.Sprite{
				Img:  vitaminesImg,
				X:    (constants.GameWidth / 2) + 16 + 200,
				Y:    (constants.GameHeight / 2) + 16 + 100,
				Size: constants.VitaminSize,
			},
			Animations: map[entities.VitaminState]*animations.Animation{
				entities.Blue:   animations.NewAnimation(0, 29, 1, 5.0),
				entities.Red:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Green:  animations.NewAnimation(60, 89, 1, 5.0),
				entities.Bronze: animations.NewAnimation(90, 119, 1, 5.0),
			},
			CombatComp: components.NewEnemyCombat(1, 0, 0),
			Speed:      1,
			Efficiency: 1,
			TempHP:     3,
			Duration:   3,
			StopCalory: false,
			Type:       2, // green
		},
		{
			Sprite: &entities.Sprite{
				Img:  vitaminesImg,
				X:    (constants.GameWidth / 2) + 16 + 200,
				Y:    (constants.GameHeight / 2) + 16 + 200,
				Size: constants.VitaminSize,
			},
			Animations: map[entities.VitaminState]*animations.Animation{
				entities.Blue:   animations.NewAnimation(0, 29, 1, 5.0),
				entities.Red:    animations.NewAnimation(30, 59, 1, 5.0),
				entities.Green:  animations.NewAnimation(60, 89, 1, 5.0),
				entities.Bronze: animations.NewAnimation(90, 119, 1, 5.0),
			},
			CombatComp: components.NewEnemyCombat(1, 0, 0),
			Speed:      1,
			Efficiency: 1,
			TempHP:     0,
			Duration:   3,
			StopCalory: true,
			Type:       3, // bronze
		},
	}
	g.vitaminSpriteSheet = vitaminSpriteSheet
	g.tilemapJSON = tilemapJSON
	g.tilemapImg = tilemapImg
	g.tilesets = tilesets
	g.cam = camera.NewCamera(0.0, 0.0)
	g.colliders = []image.Rectangle{
		image.Rect(int(constants.GameWidth/2)+300, int(constants.GameHeight/2)+300, int(constants.GameWidth/2)+332, int(constants.GameHeight/2)+332),
	}

	g.loaded = true
}

func (g *GameScene) OnEnter() {
	g.gamePause = false
}

func (g *GameScene) OnExit() {
	g.gamePause = true
}

func (g *GameScene) Update() SceneId {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ExitSceneId
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return PauseSceneId
	}
	if !g.gamePause && !g.gameOver {
		// Calories
		if g.caloryCount {
			g.player.Calories -= 0.1 * g.player.Efficiency * g.player.EfficiencyMultiplier
		}
		// Player movement
		g.player.Dx = 0.0
		g.player.Dy = 0.0

		if ebiten.IsKeyPressed(ebiten.KeyD) && (!ebiten.IsKeyPressed(ebiten.KeyW) && !ebiten.IsKeyPressed(ebiten.KeyS)) {
			g.player.Dx = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) && (!ebiten.IsKeyPressed(ebiten.KeyW) && !ebiten.IsKeyPressed(ebiten.KeyS)) {
			g.player.Dx = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier
		}
		if ebiten.IsKeyPressed(ebiten.KeyW) && (!ebiten.IsKeyPressed(ebiten.KeyD) && !ebiten.IsKeyPressed(ebiten.KeyA)) {
			g.player.Dy = -(0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) && (!ebiten.IsKeyPressed(ebiten.KeyA) && !ebiten.IsKeyPressed(ebiten.KeyD)) {
			g.player.Dy = (0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier
		}
		if ebiten.IsKeyPressed(ebiten.KeyW) && ebiten.IsKeyPressed(ebiten.KeyD) {
			g.player.Dx = ((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
			g.player.Dy = -((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
		}
		if ebiten.IsKeyPressed(ebiten.KeyW) && ebiten.IsKeyPressed(ebiten.KeyA) {
			g.player.Dx = -((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
			g.player.Dy = -((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyD) {
			g.player.Dx = ((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
			g.player.Dy = ((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyA) {
			g.player.Dx = -((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
			g.player.Dy = ((0.1 + 2*(math.Log(1+g.player.Speed))) * g.player.SpeedMultiplier) / 1.4
		}

		g.player.X += g.player.Dx
		CheckCollisionHorizontal(g.player.Sprite, g.colliders)

		g.player.Y += g.player.Dy
		CheckCollisionVertical(g.player.Sprite, g.colliders)

		activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
		if activeAnim != nil {
			activeAnim.Update()
		}

		// clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
		// cX, cY := ebiten.CursorPosition()
		// cX -= int(g.cam.X)
		// cY -= int(g.cam.Y)
		// g.player.CombatComp.Update()
		pRect := image.Rect(
			int(g.player.X),
			int(g.player.Y),
			int(g.player.X+(constants.Tilesize*g.player.Size)),
			int(g.player.Y+(constants.Tilesize*g.player.Size)),
		)

		deadEnemies := make(map[int]struct{})
		for index, enemy := range g.enemies {
			activeAnim := enemy.ActiveAnimation(enemy.Type)
			if activeAnim != nil {
				activeAnim.Update()
			}
			enemy.Dx = 0.0
			enemy.Dy = 0.0

			if enemy.Follows {
				enemy.FollowsTarget(g.player.Sprite)
			}
			enemy.CombatComp.Update()
			g.player.CombatComp.Update()
			rect := image.Rect(
				int(enemy.X),
				int(enemy.Y),
				int(enemy.X+(constants.Tilesize*enemy.Size)),
				int(enemy.Y+(constants.Tilesize*enemy.Size)),
			)
			enemy.X += enemy.Dx
			CheckCollisionHorizontal(enemy.Sprite, g.colliders)
			CheckCollisionHorizontal(enemy.Sprite, []image.Rectangle{pRect})
			enemy.Y += enemy.Dy
			CheckCollisionVertical(enemy.Sprite, g.colliders)
			CheckCollisionVertical(enemy.Sprite, []image.Rectangle{pRect})

			if rect.Overlaps(pRect) {

				// enemy attack player
				if enemy.CombatComp.Attack() {
					g.player.CombatComp.Damage(enemy.CombatComp.AttackPower(), g.player.TempHP)
					if g.player.CombatComp.Health() <= 0 {
						g.gameOver = true
						// Game over screen here
					}
				}
				// player attack enemy
				if g.player.Diet == enemy.Type || enemy.Type == 2 {
					if g.player.CombatComp.Attack() {
						enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())
						if enemy.CombatComp.Health() <= 0 {
							if enemy.Type == 2 {
								g.player.Calories += 200
							} else {
								g.player.Calories += 50
							}
							deadEnemies[index] = struct{}{}
						}
					}
				}

			}
			// if g.player.X > float64(rect.Min.X) && g.player.X < float64(rect.Max.X) && g.player.Y > float64(rect.Min.Y) && g.player.Y < float64(rect.Max.Y) {
			// if cX > rect.Min.X && cX < rect.Max.X && cY > rect.Min.Y && cY < rect.Max.Y {
			// 	if clicked {
			// 		enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())
			// 	}
			// 	if enemy.CombatComp.Health() <= 0 {
			// 		deadEnemies[index] = struct{}{}
			// 	}
			// }
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
		deadVitamins := make(map[int]struct{})
		for index, vitamin := range g.vitamins {
			vitamin.CombatComp.Update()
			g.player.CombatComp.Update()
			rect := image.Rect(
				int(vitamin.X),
				int(vitamin.Y),
				int(vitamin.X+(constants.Tilesize*vitamin.Size)),
				int(vitamin.Y+(constants.Tilesize*vitamin.Size)),
			)

			if rect.Overlaps(pRect) {
				if g.player.CombatComp.Attack() {
					vitamin.CombatComp.Damage(1)
					deadVitamins[index] = struct{}{}
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
			activeAnim := vitamin.ActiveAnimation(vitamin.Type)
			if activeAnim != nil {
				activeAnim.Update()
			}
		}
		if len(deadVitamins) > 0 {
			newVitamins := make([]*entities.Vitamin, 0)
			for index, vitamin := range g.vitamins {
				if _, exists := deadVitamins[index]; !exists {
					newVitamins = append(newVitamins, vitamin)
				}
			}
			g.vitamins = newVitamins
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
		for _, sprite := range g.enemies {
			if sprite.X >= constants.GameWidth-constants.WindowWidth {
				sprite.X = 0 + constants.WindowWidth + 1
			}
			if sprite.X <= 0+constants.WindowWidth {
				sprite.X = constants.GameWidth - constants.WindowWidth - 1
			}
			if sprite.Y >= constants.GameHeight-constants.WindowHeight {
				sprite.Y = 0 + constants.WindowHeight + 1
			}
			if sprite.Y <= 0+constants.WindowHeight {
				sprite.Y = constants.GameHeight - constants.WindowHeight - 1
			}
		}
		g.cam.FollowTarget(g.player.X+(constants.Tilesize/2), g.player.Y+(constants.Tilesize/2), constants.WindowWidth, constants.WindowHeight)
		g.cam.Constrain(constants.GameWidth, constants.GameHeight, constants.WindowWidth, constants.WindowHeight)

		// Food spawning
		if len(g.enemies) < constants.FoodLimit {
			chanceForFood := rand.IntN(2)
			if chanceForFood%2 == 0 {
				enemiesImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
				if err != nil {
					log.Fatal(err)
				}
				newFood := &entities.Enemy{Sprite: &entities.Sprite{
					Img:  enemiesImg,
					X:    float64(randRange(0+constants.WindowWidth, constants.GameWidth-constants.WindowWidth)),
					Y:    float64(randRange(0+constants.WindowHeight, constants.GameHeight-constants.WindowHeight)),
					Size: constants.FoodSize,
				},
					Animations: map[entities.EnemyState]*animations.Animation{
						entities.Meat:      animations.NewAnimation(0, 29, 1, 5.0),
						entities.Plant:     animations.NewAnimation(30, 59, 1, 5.0),
						entities.Agressive: animations.NewAnimation(60, 89, 1, 5.0),
					},
					Follows:    false,
					CombatComp: components.NewEnemyCombat(1, 0, 30),
					Type:       rand.IntN(2),
				}
				g.enemies = append(g.enemies, newFood)
			}
		}

		// Vitamin spawning
		if len(g.vitamins) < constants.VitaminLimit {
			chanceForFood := rand.IntN(2)
			if chanceForFood%2 == 0 {
				vitaminesImg, _, err := ebitenutil.NewImageFromFile("assets/images/vitamines.png")
				if err != nil {
					log.Fatal(err)
				}

				Vspeed := 0.5
				Vefficiency := 0.5
				VtempHP := 0
				Vduration := 3
				VstopCalory := false
				VTypeSpawn := rand.IntN(3)

				if VTypeSpawn == 0 {
					Vspeed = 0.5
					Vefficiency = 0.5
					VtempHP = 0
					Vduration = 3
					VstopCalory = false
				}
				if VTypeSpawn == 1 {
					Vspeed = 1.5
					Vefficiency = 1.5
					VtempHP = 0
					Vduration = 3
					VstopCalory = false
				}
				if VTypeSpawn == 2 {
					Vspeed = 1
					Vefficiency = 1
					VtempHP = 10
					Vduration = 3
					VstopCalory = false
				}
				if VTypeSpawn == 3 {
					Vspeed = 1
					Vefficiency = 1
					VtempHP = 0
					Vduration = 3
					VstopCalory = true
				}
				newVitamin := &entities.Vitamin{
					Sprite: &entities.Sprite{
						Img:  vitaminesImg,
						X:    float64(randRange(0+constants.WindowWidth, constants.GameWidth-constants.WindowWidth)),
						Y:    float64(randRange(0+constants.WindowHeight, constants.GameHeight-constants.WindowHeight)),
						Size: constants.VitaminSize,
					},
					Animations: map[entities.VitaminState]*animations.Animation{
						entities.Blue:   animations.NewAnimation(0, 29, 1, 5.0),
						entities.Red:    animations.NewAnimation(30, 59, 1, 5.0),
						entities.Green:  animations.NewAnimation(60, 89, 1, 5.0),
						entities.Bronze: animations.NewAnimation(90, 119, 1, 5.0),
					},
					CombatComp: components.NewEnemyCombat(1, 0, 0),
					Speed:      Vspeed,
					Efficiency: Vefficiency,
					TempHP:     float64(VtempHP),
					Duration:   float64(Vduration),
					StopCalory: VstopCalory,
					Type:       VTypeSpawn,
				}
				g.vitamins = append(g.vitamins, newVitamin)
			}
		}
	}
	return GameSceneId

}

var _ Scene = (*GameScene)(nil)

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(image.Rect(
			int(sprite.X),
			int(sprite.Y),
			int(sprite.X)+constants.Tilesize,
			int(sprite.Y)+constants.Tilesize)) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - constants.Tilesize
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
			int(sprite.X)+constants.Tilesize,
			int(sprite.Y)+constants.Tilesize)) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - constants.Tilesize
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}
func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}
