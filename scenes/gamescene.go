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
	"projectEVA/data"
	"projectEVA/entities"
	"projectEVA/spritesheet"
	"projectEVA/tilemap"
	"projectEVA/tileset"

	"github.com/gbatagian/deepsort"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// testowanie populacji
var population []*data.Genom
var currentGenom *data.Genom
var currentGenIndex int
var generation int = 1

// Limit czasu trwania życia genomu (w sekundach i klatkach)
const GenomLifetimeInSeconds = 120
const FramesPerSecond = 60
const GenomLifetimeFrames = GenomLifetimeInSeconds * FramesPerSecond

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
	foodEaten          int
	enemyKilled        int
	timePassed         int
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
		enemyKilled:        0,
		foodEaten:          0,
		timePassed:         0,
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
		fmt.Sprintf("Game State: \n Game Pause: %v\n Game Over: %v\n Score: %v\n Enemies on map: %v\n Food on map: %v\n Vitamins on map: %v", g.gamePause, g.gameOver, SCORE, numberOfEnemies, numberOfFood, len(g.vitamins)), 0, 300)
	if currentGenom != nil {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Genom: %d/%d\nGeneracja: %d\nFitness: %.2f",
				currentGenIndex+1, len(population), generation, currentGenom.Fitness),
			10, 500)
	}

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
	// vitaminesImg, _, err := ebitenutil.NewImageFromFile("assets/images/vitamines.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// enemiesImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/Water.png")
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
			entities.W:    animations.NewAnimation(0, 29, 1, 5.0),
			entities.WD:   animations.NewAnimation(30, 59, 1, 5.0),
			entities.D:    animations.NewAnimation(60, 89, 1, 5.0),
			entities.DS:   animations.NewAnimation(90, 119, 1, 5.0),
			entities.S:    animations.NewAnimation(120, 149, 1, 5.0),
			entities.SA:   animations.NewAnimation(150, 179, 1, 5.0),
			entities.A:    animations.NewAnimation(180, 209, 1, 5.0),
			entities.AW:   animations.NewAnimation(210, 239, 1, 5.0),
			entities.Idle: animations.NewAnimation(240, 269, 1, 5.0),
			// entities.W:    animations.NewAnimation(0, 0, 1, 5.0),
			// entities.WD:   animations.NewAnimation(30, 30, 1, 5.0),
			// entities.D:    animations.NewAnimation(60, 60, 1, 5.0),
			// entities.DS:   animations.NewAnimation(90, 90, 1, 5.0),
			// entities.S:    animations.NewAnimation(120, 120, 1, 5.0),
			// entities.SA:   animations.NewAnimation(150, 150, 1, 5.0),
			// entities.A:    animations.NewAnimation(180, 180, 1, 5.0),
			// entities.AW:   animations.NewAnimation(210, 210, 1, 5.0),
			// entities.Idle: animations.NewAnimation(240, 240, 1, 5.0),
		},
		CombatComp: components.NewPlayerCombat(3, 1, 6000),
		Diet:       0,
		Dmg:        1,
		MaxHealth:  3,
	}
	g.playerSpriteSheet = playerSpriteSheet

	g.enemies = []*entities.Enemy{}
	g.enemySpriteSheet = enemySpriteSheet
	g.vitamins = []*entities.Vitamin{}
	g.vitaminSpriteSheet = vitaminSpriteSheet
	g.tilemapJSON = tilemapJSON
	g.tilemapImg = tilemapImg
	g.tilesets = tilesets
	g.cam = camera.NewCamera(0.0, 0.0)
	g.colliders = []image.Rectangle{}

	g.foodEaten = 0
	g.enemyKilled = 0
	g.timePassed = 0

	//tworzenie ai do testow - START
	population = []*data.Genom{}
	sharedHistory := &data.InnovationHistory{}
	for i := 0; i < 100; i++ {
		g := &data.Genom{
			NumInputs:  15,
			NumOutputs: 8,
			//			TotalNodes:       23, //uwazac bo createnetwork tutaj dodaje - nie jest to wgl potrzebne tbh
			Nodes:            []*data.Node{},
			ConnCreationRate: 1.0,
			IH:               sharedHistory,
			Fitness:          0,
		}
		g.CreateNetwork()
		fmt.Printf("\nGENOM #%d\n", i)
		for _, c := range g.Connections {
			fmt.Printf("Połączenie: In=%d (Type %d) → Out=%d (Type %d), Waga=%.2f\n",
				c.InNode.ID, c.InNode.Type,
				c.OutNode.ID, c.OutNode.Type,
				c.Weight,
			)
		}
		population = append(population, g)
	}
	currentGenIndex = 0
	currentGenom = population[currentGenIndex]
	// fmt.Println("Test fitness:", testGenom.EvaluateFitness(120, 3, 56, 32, 2)) //sprawdzanie dzialania funkcji fitness
	// fmt.Printf("Utworzono populację z %d genomów\n", len(population)) //sprawdzanie czy populacja zostala stworzona
	//print sprawdzajacy polaczenia w kazdym genomie
	fmt.Println("=== PODGLĄD POPULACJI ===")
	for i, genom := range population {
		fmt.Printf("GENOM %d:\n", i)
		fmt.Printf("  NODES:\n")
		for _, node := range genom.Nodes {
			fmt.Printf("    Node %d (Type: %d)\n", node.ID, node.Type)
		}
		fmt.Printf("  CONNECTIONS:\n")
		for _, conn := range genom.Connections {
			fmt.Printf("    [%d] %d -> %d | Weight: %.2f | Enabled: %v\n",
				conn.Innovation, conn.InNode.ID, conn.OutNode.ID, conn.Weight, conn.Enabled)
		}
		fmt.Println("----------------------------------")
	}
	//AI - koniec
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
		//testowanie do ai - start
		if currentGenom != nil {
			g.ControlByAI(currentGenom)
		}
		//testowanie do ai - koniec
		if g.caloryCount {
			g.player.Calories -= 0.1 * g.player.Efficiency * g.player.EfficiencyMultiplier
			g.timePassed += 1
		}

		// Evolution &BALANCE
		if g.player.Calories >= 1000 {
			if g.timePassed < 3600 {
				g.player.Speed += 1
				g.player.Efficiency += 0.1
			} else {
				g.player.Speed -= 1
				g.player.Efficiency -= 0.1
			}

			if g.enemyKilled > 2 {
				g.player.Dmg += 1
			} else {
				g.player.MaxHealth += 1
			}

			if g.foodEaten > 10 {
				g.player.Efficiency -= 0.1
				g.player.MaxHealth += 1
			} else {
				g.player.Efficiency -= 0.1
				g.player.Speed += 1
			}
			g.player.CombatComp = components.NewPlayerCombat(g.player.MaxHealth+g.player.TempHP, g.player.Dmg, 6000)
			g.foodEaten = 0
			g.enemyKilled = 0
			g.timePassed = 0
			g.player.Calories = 500
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
		// enemy behavior
		deadEnemies := make(map[int]struct{})
		numberOfEnemies = 0
		numberOfFood = 0
		for index, enemy := range g.enemies {
			if enemy.Type != 2 {
				numberOfFood++
			} else {
				numberOfEnemies++
			}
			activeAnim := enemy.ActiveAnimation(enemy.Type)
			if activeAnim != nil {
				activeAnim.Update()
			}
			enemy.Dx = 0.0
			enemy.Dy = 0.0

			if enemy.Follows {
				enemy.FollowsTarget(g.player.Sprite, constants.EnemyPlayerVision)
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
			// CheckCollisionHorizontal(enemy.Sprite, []image.Rectangle{pRect})
			enemy.Y += enemy.Dy
			CheckCollisionVertical(enemy.Sprite, g.colliders)
			// CheckCollisionVertical(enemy.Sprite, []image.Rectangle{pRect})

			// enemy eating food
			if enemy.Type == 2 {
				for index2, food := range g.enemies {
					if food.Type == 0 {
						enemy.CombatComp.Update()
						fRect := image.Rect(
							int(food.X),
							int(food.Y),
							int(food.X+(constants.Tilesize*food.Size)),
							int(food.Y+(constants.Tilesize*food.Size)),
						)
						if rect.Overlaps(fRect) {
							if enemy.CombatComp.Attack() {
								food.CombatComp.Damage(enemy.CombatComp.AttackPower())
								if food.CombatComp.Health() <= 0 {
									deadEnemies[index2] = struct{}{}
								}
							}
						}
					}
				}
			}
			if rect.Overlaps(pRect) {

				// enemy attack player
				if enemy.CombatComp.Attack() {
					g.player.CombatComp.Damage(enemy.CombatComp.AttackPower())
					if g.player.CombatComp.Health() <= 0 {
						g.gameOver = true
						// Game over screen here
					}
				}
				// player attack enemy
				if g.player.Diet == enemy.Type || g.player.Diet == 2 || enemy.Type == 2 {
					if g.player.CombatComp.Attack() {
						enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())
						if enemy.CombatComp.Health() <= 0 {
							if enemy.Type == 2 {
								if g.player.Diet == 2 {
									g.player.Calories += 100
									SCORE += 100
								} else {
									g.player.Calories += 200
									SCORE += 200
								}
								g.enemyKilled += 1
							} else {
								if g.player.Diet == 2 {
									g.player.Calories += 25
									SCORE += 25
								} else {
									g.player.Calories += 50
									SCORE += 50
								}
								g.foodEaten += 1
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
					g.player.CombatComp = components.NewPlayerCombat(g.player.MaxHealth+g.player.TempHP, g.player.Dmg, 3000)
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
		} else if g.vitaminDuration == 0 {
			g.caloryCount = true
			g.player.SpeedMultiplier = 1
			g.player.EfficiencyMultiplier = 1
			g.player.TempHP = 0
			g.vitaminDuration = 0
			g.player.CombatComp = components.NewPlayerCombat(g.player.MaxHealth+g.player.TempHP, g.player.Dmg, 6000)
			g.vitaminDuration--
		}

		// Teleport map edge
		if g.player.X >= constants.GameWidth {
			g.player.X = 1
		}
		if g.player.X <= 0 {
			g.player.X = constants.GameWidth - 1
		}
		if g.player.Y >= constants.GameHeight {
			g.player.Y = 1
		}
		if g.player.Y <= 0 {
			g.player.Y = constants.GameHeight - 1
		}
		for _, sprite := range g.enemies {
			if sprite.X >= constants.GameWidth {
				sprite.X = 1
			}
			if sprite.X <= 0 {
				sprite.X = constants.GameWidth - 1
			}
			if sprite.Y >= constants.GameHeight {
				sprite.Y = 1
			}
			if sprite.Y <= 0 {
				sprite.Y = constants.GameHeight - 1
			}
		}
		g.cam.FollowTarget(g.player.X+(constants.Tilesize/2), g.player.Y+(constants.Tilesize/2), constants.WindowWidth, constants.WindowHeight)
		g.cam.Constrain(constants.GameWidth, constants.GameHeight, constants.WindowWidth, constants.WindowHeight)

		// Food spawning
		if numberOfFood < constants.FoodLimit {
			chanceForFood := rand.IntN(2)
			if chanceForFood%2 == 0 {
				enemiesImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
				if err != nil {
					log.Fatal(err)
				}
				newFood := &entities.Enemy{Sprite: &entities.Sprite{
					Img:  enemiesImg,
					X:    float64(randRange(0, constants.GameWidth)),
					Y:    float64(randRange(0, constants.GameHeight)),
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
						X:    float64(randRange(0, constants.GameWidth)),
						Y:    float64(randRange(0, constants.GameHeight)),
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

		// Enemy spawning
		if numberOfEnemies < constants.EnemyLimit {
			enemiesImg, _, err := ebitenutil.NewImageFromFile("assets/images/enemies.png")
			if err != nil {
				log.Fatal(err)
			}
			newEnemy := &entities.Enemy{
				Sprite: &entities.Sprite{
					Img:  enemiesImg,
					X:    float64(randRange(0, constants.GameWidth)),
					Y:    float64(randRange(0, constants.GameHeight)),
					Size: 1,
				},
				Animations: map[entities.EnemyState]*animations.Animation{
					entities.Meat:      animations.NewAnimation(0, 29, 1, 5.0),
					entities.Plant:     animations.NewAnimation(30, 59, 1, 5.0),
					entities.Agressive: animations.NewAnimation(60, 89, 1, 5.0),
				},
				Follows:    true,
				CombatComp: components.NewEnemyCombat(float64(randRange(int(g.player.MaxHealth*0.9), int(g.player.MaxHealth*1.1))), math.Max(1, float64(randRange(int(g.player.Dmg*0.9), int(g.player.Dmg*1.1)))), 3000),
				Type:       2,
				Speed:      float64(randRange(int(g.player.Speed*0.9), int(g.player.Speed*1.1))),
			}
			g.enemies = append(g.enemies, newEnemy)
		}

		// AI VARS ?
		PLAYERCALORIES = g.player.Calories
		PLAYERHP = g.player.CombatComp.Health()
		PLAYERDMG = g.player.Dmg
		PLAYERSPEED = g.player.Speed
		PLAYEREFFICIENCY = g.player.Efficiency
		PLAYERX = g.player.X
		PLAYERY = g.player.Y
		ENEMIES = make([][3]float64, 0)
		NEARFOODS = make([][]float64, 0)
		NEARVITAMINS = make([][]float64, 0)
		for _, enemy := range g.enemies {
			if enemy.Type == 2 {
				dystans := math.Sqrt(math.Pow(g.player.X-enemy.X, 2) + math.Pow(g.player.Y-enemy.Y, 2))
				kat := g.player.X - enemy.X/g.player.Y - enemy.Y
				hp := enemy.CombatComp.Health()
				tablica := [3]float64{dystans, kat, hp}
				ENEMIES = append(ENEMIES, tablica)
			}
			if enemy.Type == 0 && (g.player.Diet == 0 || g.player.Diet == 2) {
				dystans := math.Sqrt(math.Pow(g.player.X-enemy.X, 2) + math.Pow(g.player.Y-enemy.Y, 2))
				kat := g.player.X - enemy.X/g.player.Y - enemy.Y
				tablica := []float64{dystans, kat}
				NEARFOODS = append(NEARFOODS, tablica)
			}
			if enemy.Type == 1 && (g.player.Diet == 0 || g.player.Diet == 2) {
				dystans := math.Sqrt(math.Pow(g.player.X-enemy.X, 2) + math.Pow(g.player.Y-enemy.Y, 2))
				kat := g.player.X - enemy.X/g.player.Y - enemy.Y
				tablica := []float64{dystans, kat}
				NEARFOODS = append(NEARFOODS, tablica)
			}
		}
		deepsort.DeepSort(&NEARFOODS, []float64{0})
		newNEARFOODS := make([][]float64, 0)
		for index, _ := range NEARFOODS {
			if index < 10 {
				newNEARFOODS = append(newNEARFOODS, NEARFOODS[index])
			} else {
				break
			}
		}
		deepsort.DeepSort(&NEARVITAMINS, []float64{0})
		newNEARVITAMINS := make([][]float64, 0)
		for index, _ := range NEARVITAMINS {
			if index < 10 {
				newNEARVITAMINS = append(newNEARVITAMINS, NEARVITAMINS[index])
			} else {
				break
			}
		}
		NEARFOODS = newNEARFOODS
		//zapewnia nie danie pustej tablicy (wypluwa wtedy puste outputy)
		if len(NEARFOODS) > 0 {
			if len(NEARFOODS[0]) > 0 {
				fmt.Println(NEARFOODS[0][0])
			}
		}
		NEARVITAMINS = newNEARVITAMINS
		//zapewnia nie danie pustej tablicy (wypluwa wtedy puste outputy)
		if len(NEARVITAMINS) > 0 {
			if len(NEARVITAMINS[0]) > 0 {
				fmt.Println(NEARVITAMINS[0][0])
			}
		}
		// println(ENEMIES[0][0])
	}
	// dane do funkcji kosztu
	//przechodzenie po genomach - start
	if g.gameOver || g.timePassed >= GenomLifetimeFrames {
		fitness := currentGenom.EvaluateFitness(SCORE, g.foodEaten, g.enemyKilled, g.timePassed, g.player.CombatComp.Health())
		fmt.Printf("Genom %d fitness: %f\n", currentGenIndex, fitness)

		currentGenIndex++
		if currentGenIndex < len(population) {
			currentGenom = population[currentGenIndex]
			g.ResetGameState()
		} else {
			fmt.Println("Koniec pokolenia")
			// TODO: tu dodamy selekcję, krzyżowanie, mutację
			g.gamePause = true
		}
	}
	//przchodzenie po genomach - koniec
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

var numberOfFood int = 0
var numberOfEnemies int = 0
var numberOfVitamins int = 0

var PLAYERCALORIES float64 = 0
var SCORE int = 0
var PLAYERHP float64 = 0
var PLAYERDMG float64 = 0
var PLAYERSPEED float64 = 0
var PLAYEREFFICIENCY float64 = 0
var PLAYERX float64 = 0
var PLAYERY float64 = 0

var ENEMIES []([3]float64) = make([][3]float64, 0)
var NEARFOODS []([]float64) = make([][]float64, 0)
var NEARVITAMINS []([]float64) = make([][]float64, 0)

//laczenie AI z gra

func (g *GameScene) ControlByAI(genom *data.Genom) {
	print("=== AI CONTROL ===")
	inputs := g.PrepareInputs()
	fmt.Printf("INPUTS to NEAT: %v\n", inputs)
	outputs := genom.Forward(inputs)
	fmt.Printf("OUTPUTS z NEAT: %v (len: %d)\n", outputs, len(outputs))
	if len(outputs) < 8 {
		return
	}
	directions := [8][2]float64{
		{0, -1},  // ↑
		{0, 1},   // ↓
		{-1, 0},  // ←
		{1, 0},   // →
		{-1, -1}, // ↖
		{1, -1},  // ↗
		{1, 1},   // ↘
		{-1, 1},  // ↙
	}
	// szukamy kierunku o najwyższym output
	bestIndex := 0
	bestValue := outputs[0]
	for i, val := range outputs {
		if val > bestValue {
			bestValue = val
			bestIndex = i
		}
	}

	moveScale := (0.1 + 2*math.Log(1+g.player.Speed)) * g.player.SpeedMultiplier
	dir := directions[bestIndex]

	length := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1])
	if length != 0 {
		dir[0] /= length
		dir[1] /= length
	}
	g.player.Dx = (dir[0]*2 - 1) * moveScale
	g.player.Dy = (dir[1]*2 - 1) * moveScale
}

// przygotowanie inputow dla NEATA
func (g *GameScene) PrepareInputs() []float64 {
	inputs := []float64{
		float64(SCORE) / 10.0,
		PLAYERHP / 10.0,
		PLAYERDMG / 10.0,
		PLAYERSPEED / 10.0,
		PLAYEREFFICIENCY / 10.0,
		PLAYERX / float64(constants.GameWidth),
		PLAYERY / float64(constants.GameHeight),
		PLAYERCALORIES / 10.0,
	}

	if len(NEARFOODS) > 0 {
		distance := NEARFOODS[0][0] / 500.0
		angle := NEARFOODS[0][1] / 180.0
		inputs = append(inputs, distance, angle)
	} else {
		inputs = append(inputs, 0.0, 0.0)

	}

	if len(NEARVITAMINS) > 0 {
		distance := NEARVITAMINS[0][0] / 500.0
		angle := NEARFOODS[0][1] / 180.0
		inputs = append(inputs, distance, angle)
	} else {
		inputs = append(inputs, 0.0, 0.0)
	}

	if len(ENEMIES) > 0 {
		distance := ENEMIES[0][0] / 500.0
		angle := ENEMIES[0][1] / 180.0
		enemyHP := ENEMIES[0][2] / 10.0
		inputs = append(inputs, distance, angle, enemyHP)
	} else {
		inputs = append(inputs, 0.0, 0.0, 0.0)
	}
	fmt.Printf("INPUTS to NEAT: %+v\n", inputs)
	return inputs
}

// funkcja resetujaca gre dla ai
func (g *GameScene) ResetGameState() {
	// Reset playera
	g.player.X = (constants.GameWidth / 2) + 16
	g.player.Y = (constants.GameHeight / 2) + 16
	g.player.Calories = 500.0
	g.player.Speed = 5
	g.player.Efficiency = 1
	g.player.SpeedMultiplier = 1
	g.player.EfficiencyMultiplier = 1
	g.player.TempHP = 0
	g.player.Diet = 0
	g.player.Dmg = 1
	g.player.MaxHealth = 3
	g.player.CombatComp = components.NewPlayerCombat(3, 1, 6000)

	// Reset mapy i przeciwników
	g.enemies = []*entities.Enemy{}
	g.vitamins = []*entities.Vitamin{}
	g.foodEaten = 0
	g.enemyKilled = 0
	g.timePassed = 0
	g.vitaminDuration = 0
	g.caloryCount = true
	g.gameOver = false

	// Reset kamery
	g.cam = camera.NewCamera(0.0, 0.0)

	// Reset zmiennych globalnych NEAT-a:
	SCORE = 0
	numberOfFood = 0
	numberOfEnemies = 0
	PLAYERHP = 0
	PLAYERDMG = 0
	PLAYERSPEED = 0
	PLAYEREFFICIENCY = 0
	PLAYERX = 0
	PLAYERY = 0
	ENEMIES = make([][3]float64, 0)
	NEARFOODS = make([][]float64, 0)
}
