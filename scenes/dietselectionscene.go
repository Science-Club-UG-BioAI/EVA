package scenes

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

var PlayerDiet int

type DietSelectionScene struct {
	loaded              bool
	backgroundImage     *ebiten.Image
	carnivoreButtonRect Rect
	omnivoreButtonRect  Rect
	herbivoreButtonRect Rect
	selectedDiet        string
	startButtonRect     Rect
}

func NewDietSelectionScene() *DietSelectionScene {
	return &DietSelectionScene{
		carnivoreButtonRect: Rect{X: 100, Y: 150, Width: 200, Height: 50},
		omnivoreButtonRect:  Rect{X: 100, Y: 220, Width: 200, Height: 50},
		herbivoreButtonRect: Rect{X: 100, Y: 290, Width: 200, Height: 50},
		startButtonRect:     Rect{X: 380, Y: 400, Width: 200, Height: 50},
	}
}

func (s *DietSelectionScene) FirstLoad() {
	s.loaded = true
	img, _, err := ebitenutil.NewImageFromFile("assets/images/start_background.png")
	if err != nil {
		log.Fatalf("failed to load background image: %v", err)
	}
	s.backgroundImage = img
}

func (s *DietSelectionScene) IsLoaded() bool {
	return s.loaded
}

func (s *DietSelectionScene) Update() SceneId {
	// Handle button clicks
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()

		if cursorX >= s.carnivoreButtonRect.X && cursorX <= s.carnivoreButtonRect.X+s.carnivoreButtonRect.Width &&
			cursorY >= s.carnivoreButtonRect.Y && cursorY <= s.carnivoreButtonRect.Y+s.carnivoreButtonRect.Height {
			s.selectedDiet = "Carnivore"
			PlayerDiet = 0
		} else if cursorX >= s.omnivoreButtonRect.X && cursorX <= s.omnivoreButtonRect.X+s.omnivoreButtonRect.Width &&
			cursorY >= s.omnivoreButtonRect.Y && cursorY <= s.omnivoreButtonRect.Y+s.omnivoreButtonRect.Height {
			s.selectedDiet = "Omnivore"
			PlayerDiet = 2
		} else if cursorX >= s.herbivoreButtonRect.X && cursorX <= s.herbivoreButtonRect.X+s.herbivoreButtonRect.Width &&
			cursorY >= s.herbivoreButtonRect.Y && cursorY <= s.herbivoreButtonRect.Y+s.herbivoreButtonRect.Height {
			s.selectedDiet = "Herbivore"
			PlayerDiet = 1
		}
	}

	// Handle "START" button click
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		if cursorX >= s.startButtonRect.X && cursorX <= s.startButtonRect.X+s.startButtonRect.Width &&
			cursorY >= s.startButtonRect.Y && cursorY <= s.startButtonRect.Y+s.startButtonRect.Height {
			if s.selectedDiet != "" {
				return GameSceneId // Transition to the game scene
			}
		}
	}

	// if needed this id can be updated to transition to another scene
	return DietSelectionSceneId
}

func (s *DietSelectionScene) Draw(screen *ebiten.Image) {
	if s.backgroundImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(s.backgroundImage, op)
	}

	// Draw "CHOOSE YOUR DIET" text
	chooseDietText := "CHOOSE YOUR DIET"
	textColor := color.White
	bounds := text.BoundString(basicfont.Face7x13, chooseDietText)
	textWidth := bounds.Dx()
	textHeight := bounds.Dy()

	textImage := ebiten.NewImage(textWidth, textHeight)
	text.Draw(textImage, chooseDietText, basicfont.Face7x13, 0, textHeight, textColor)

	op := &ebiten.DrawImageOptions{}
	scaleFactor := 4.0
	op.GeoM.Scale(scaleFactor, scaleFactor)

	// Adjust "CHOOSE YOUR DIET" position 
	screenWidth, _ := screen.Size()
	x := (float64(screenWidth) - float64(textWidth)*scaleFactor) / 2
	y := float64(s.carnivoreButtonRect.Y) - float64(textHeight)*scaleFactor - 60 
	op.GeoM.Translate(x, y)

	screen.DrawImage(textImage, op)

	// Center diet buttons in a row and vertically in the middle of the screen
	totalButtonWidth := s.carnivoreButtonRect.Width + s.omnivoreButtonRect.Width + s.herbivoreButtonRect.Width + 40 
	startX := (screenWidth - totalButtonWidth) / 2
	_, screenHeight := screen.Size()
	buttonY := (screenHeight - s.carnivoreButtonRect.Height) / 2

	// Update button positions
	s.carnivoreButtonRect.X = startX
	s.carnivoreButtonRect.Y = buttonY

	s.omnivoreButtonRect.X = startX + s.carnivoreButtonRect.Width + 20
	s.omnivoreButtonRect.Y = buttonY

	s.herbivoreButtonRect.X = s.omnivoreButtonRect.X + s.omnivoreButtonRect.Width + 20
	s.herbivoreButtonRect.Y = buttonY

	// Draw DIET buttons
	s.drawButton(screen, s.carnivoreButtonRect, "Carnivore", color.RGBA{R: 249, G: 209, B: 66, A: 100}, "assets/images/meat.png", color.RGBA{R: 189, G: 77, B: 39, A: 255})
	s.drawButton(screen, s.omnivoreButtonRect, "Omnivore", color.RGBA{R: 249, G: 209, B: 66, A: 100}, "assets/images/all-foods.png", color.RGBA{R: 189, G: 77, B: 39, A: 255})
	s.drawButton(screen, s.herbivoreButtonRect, "Herbivore", color.RGBA{R: 249, G: 209, B: 66, A: 100}, "assets/images/plant.png", color.RGBA{R: 189, G: 77, B: 39, A: 255})

	// Draw "START" button
	buttonColor := color.RGBA{R: 249, G: 209, B: 66, A: 100}
	s.drawButton(screen, s.startButtonRect, "START", buttonColor, "", color.RGBA{R: 189, G: 77, B: 39, A: 255})


	// Display selected diet below the buttons + position it
	if s.selectedDiet != "" {
		selectedDietText := "Selected Diet: " + s.selectedDiet
		textColor := color.White
		bounds := text.BoundString(basicfont.Face7x13, selectedDietText)
		textWidth := bounds.Dx()
		textHeight := bounds.Dy()

		screenWidth, _ := screen.Size()
		textX := (screenWidth - textWidth) / 2
		textY := s.carnivoreButtonRect.Y + s.carnivoreButtonRect.Height + 40 // Position below buttons

		text.Draw(screen, selectedDietText, basicfont.Face7x13, textX, textY+textHeight, textColor)
	}
}

func (s *DietSelectionScene) drawButton(screen *ebiten.Image, rect Rect, label string, buttonColor color.Color, imagePath string, textColor color.Color) {
	radius := 4 // Radius for rounded corners

	// Draw button background (excluding corners)
	for dx := radius; dx < rect.Width-radius; dx++ {
		for dy := 0; dy < rect.Height; dy++ {
			screen.Set(rect.X+dx, rect.Y+dy, buttonColor)
		}
	}
	for dx := 0; dx < rect.Width; dx++ {
		for dy := radius; dy < rect.Height-radius; dy++ {
			screen.Set(rect.X+dx, rect.Y+dy, buttonColor)
		}
	}

	// Draw rounded edges
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy <= radius*radius {

				screen.Set(rect.X+radius+dx, rect.Y+radius+dy, buttonColor)

				screen.Set(rect.X+rect.Width-radius+dx, rect.Y+radius+dy, buttonColor)

				screen.Set(rect.X+radius+dx, rect.Y+rect.Height-radius+dy, buttonColor)

				screen.Set(rect.X+rect.Width-radius+dx, rect.Y+rect.Height-radius+dy, buttonColor)
			}
		}
	}

	var img *ebiten.Image
	var imgWidth, imgHeight int
	if imagePath != "" {
		var err error
		img, _, err = ebitenutil.NewImageFromFile(imagePath)
		if err != nil {
			log.Printf("failed to load image: %v", err)
			return
		}
		imgWidth, imgHeight = img.Bounds().Dx(), img.Bounds().Dy()
	}

	// Calculate total width of content (text + spacing + image)
	labelBounds := text.BoundString(basicfont.Face7x13, label)
	labelWidth := labelBounds.Dx()
	labelHeight := labelBounds.Dy()
	totalWidth := labelWidth
	if img != nil {
		totalWidth += imgWidth + 10 // Add spacing between text and image
	}

	// Center content within the button
	contentX := rect.X + (rect.Width-totalWidth)/2
	contentY := rect.Y + (rect.Height-labelHeight)/2

	// Draw the text
	text.Draw(screen, label, basicfont.Face7x13, contentX, contentY+labelHeight, textColor)

	// Draw the image to the right of the text
	if img != nil {
		op := &ebiten.DrawImageOptions{}
		imageX := contentX + labelWidth + 10 
		imageY := rect.Y + (rect.Height-imgHeight)/2
		op.GeoM.Translate(float64(imageX), float64(imageY))
		screen.DrawImage(img, op)
	}
}

func (s *DietSelectionScene) OnEnter() {}
func (s *DietSelectionScene) OnExit()  {}
