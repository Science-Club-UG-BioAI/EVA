package scenes

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

type SceneIdd int

const (
	CharacterSelectionSceneId = 1
	DietSelectionSceneId = 2 
)

type StartScene struct {
    loaded          bool
    backgroundImage *ebiten.Image
		startButtonRect Rect
}

type Rect struct {
	X, Y, Width, Height int
}

func NewStartScene() *StartScene {
	return &StartScene{
		startButtonRect: Rect{X: 100, Y: 150, Width: 100, Height: 30},
	}
}

func (s *StartScene) FirstLoad() {
    s.loaded = true
    img, _, err := ebitenutil.NewImageFromFile("assets/images/start_background.png")
    if (err != nil) {
        log.Fatalf("failed to load background image: %v", err)
    }
    s.backgroundImage = img
}

func (s *StartScene) IsLoaded() bool {
    return s.loaded
}

func (s *StartScene) Update() SceneId {
	// Handle "PLAY" button click
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		if cursorX >= s.startButtonRect.X && cursorX <= s.startButtonRect.X+s.startButtonRect.Width &&
			cursorY >= s.startButtonRect.Y && cursorY <= s.startButtonRect.Y+s.startButtonRect.Height {
			return DietSelectionSceneId // Transition to DietSelectionScene
		}
	}
	return StartSceneId // Remain in StartScene
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	if s.backgroundImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(s.backgroundImage, op)
	}

	// creating "EVA"
	evaColor := color.White
	textToCenter := "EVA"
	bounds := text.BoundString(basicfont.Face7x13, textToCenter)
	textWidth := bounds.Dx()
	textHeight := bounds.Dy()

	textImage := ebiten.NewImage(textWidth, textHeight)
	text.Draw(textImage, textToCenter, basicfont.Face7x13, 0, textHeight, evaColor)

	op := &ebiten.DrawImageOptions{}
	scaleFactor := 4.0 
	op.GeoM.Scale(scaleFactor, scaleFactor)

	// centering "EVA" 
	screenWidth, screenHeight := screen.Size()
	x := (float64(screenWidth) - float64(textWidth)*scaleFactor) / 2
	y := (float64(screenHeight) - float64(textHeight)*scaleFactor) / 2
	op.GeoM.Translate(x, y)

	// enlarging screen text
	screen.DrawImage(textImage, op)

	// "PLAY" below "EVA"
	s.startButtonRect.X = (screenWidth - s.startButtonRect.Width) / 2
	s.startButtonRect.Y = int(y + float64(textHeight)*scaleFactor + 60)

	// drawing rectangle with rounded edges
	buttonColor := color.RGBA{R: 249, G: 209, B: 66, A: 100}
	radius := 4 // used to round the edges

	// filling of button "PLAY"
	for dx := radius; dx < s.startButtonRect.Width-radius; dx++ {
		for dy := 0; dy < s.startButtonRect.Height; dy++ {
			screen.Set(s.startButtonRect.X+dx, s.startButtonRect.Y+dy, buttonColor)
		}
	}
	for dx := 0; dx < s.startButtonRect.Width; dx++ {
		for dy := radius; dy < s.startButtonRect.Height-radius; dy++ {
			screen.Set(s.startButtonRect.X+dx, s.startButtonRect.Y+dy, buttonColor)
		}
	}

	// rounding the edges of a button
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy <= radius*radius {

				screen.Set(s.startButtonRect.X+radius+dx, s.startButtonRect.Y+radius+dy, buttonColor)

				screen.Set(s.startButtonRect.X+s.startButtonRect.Width-radius+dx, s.startButtonRect.Y+radius+dy, buttonColor)

				screen.Set(s.startButtonRect.X+radius+dx, s.startButtonRect.Y+s.startButtonRect.Height-radius+dy, buttonColor)

				screen.Set(s.startButtonRect.X+s.startButtonRect.Width-radius+dx, s.startButtonRect.Y+s.startButtonRect.Height-radius+dy, buttonColor)
			}
		}
	}

	// centering text on a button
	startText := "PLAY"
	startBounds := text.BoundString(basicfont.Face7x13, startText)
	startTextWidth := startBounds.Dx()
	startTextHeight := startBounds.Dy()
	startTextX := s.startButtonRect.X + (s.startButtonRect.Width-startTextWidth)/2
	startTextY := s.startButtonRect.Y + (s.startButtonRect.Height+startTextHeight)/2

	text.Draw(screen, startText, basicfont.Face7x13, startTextX, startTextY, color.RGBA{R: 189, G: 77, B: 39, A: 255})
}

func (s *StartScene) OnEnter() {}
func (s *StartScene) OnExit() {}