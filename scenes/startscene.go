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
    StartSceneIdd SceneIdd = iota
    CharacterSelectionSceneId
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
    
	// Obsługa kliknięcia przycisku START
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		if cursorX >= s.startButtonRect.X && cursorX <= s.startButtonRect.X+s.startButtonRect.Width &&
			cursorY >= s.startButtonRect.Y && cursorY <= s.startButtonRect.Y+s.startButtonRect.Height {
			return GameSceneId // Przełącz na GameScene
		}
	}

    return StartSceneId
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	if s.backgroundImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(s.backgroundImage, op)
	}

	// Tworzenie obrazu dla tekstu "EVA"
	evaColor := color.White
	textToCenter := "EVA"
	bounds := text.BoundString(basicfont.Face7x13, textToCenter)
	textWidth := bounds.Dx()
	textHeight := bounds.Dy()

	// Tworzenie obrazu tekstu
	textImage := ebiten.NewImage(textWidth, textHeight)
	text.Draw(textImage, textToCenter, basicfont.Face7x13, 0, textHeight, evaColor)

	// Skalowanie obrazu tekstu
	op := &ebiten.DrawImageOptions{}
	scaleFactor := 4.0 // Powiększenie tekstu 4x
	op.GeoM.Scale(scaleFactor, scaleFactor)

	// Wyśrodkowanie tekstu na ekranie
	screenWidth, screenHeight := screen.Size()
	x := (float64(screenWidth) - float64(textWidth)*scaleFactor) / 2
	y := (float64(screenHeight) - float64(textHeight)*scaleFactor) / 2
	op.GeoM.Translate(x, y)

	// Rysowanie powiększonego tekstu
	screen.DrawImage(textImage, op)

	// Przycisk "START" poniżej tekstu "EVA"
	s.startButtonRect.X = (screenWidth - s.startButtonRect.Width) / 2
	s.startButtonRect.Y = int(y + float64(textHeight)*scaleFactor + 60)

	// Rysowanie prostokątnego przycisku z zaokrąglonymi rogami
	buttonColor := color.RGBA{R: 249, G: 209, B: 66, A: 100}
	radius := 4 // Promień zaokrąglenia rogów

	// Rysowanie wypełnienia prostokąta (bez rogów)
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

	// Rysowanie rogów jako okręgów
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy <= radius*radius {
				// Lewy górny róg
				screen.Set(s.startButtonRect.X+radius+dx, s.startButtonRect.Y+radius+dy, buttonColor)
				// Prawy górny róg
				screen.Set(s.startButtonRect.X+s.startButtonRect.Width-radius+dx, s.startButtonRect.Y+radius+dy, buttonColor)
				// Lewy dolny róg
				screen.Set(s.startButtonRect.X+radius+dx, s.startButtonRect.Y+s.startButtonRect.Height-radius+dy, buttonColor)
				// Prawy dolny róg
				screen.Set(s.startButtonRect.X+s.startButtonRect.Width-radius+dx, s.startButtonRect.Y+s.startButtonRect.Height-radius+dy, buttonColor)
			}
		}
	}

	// Wyśrodkowanie napisu "START" na przycisku
	startText := "START"
	startBounds := text.BoundString(basicfont.Face7x13, startText)
	startTextWidth := startBounds.Dx()
	startTextHeight := startBounds.Dy()
	startTextX := s.startButtonRect.X + (s.startButtonRect.Width-startTextWidth)/2
	startTextY := s.startButtonRect.Y + (s.startButtonRect.Height+startTextHeight)/2

	// Rysowanie tekstu "START" na przycisku
	text.Draw(screen, startText, basicfont.Face7x13, startTextX, startTextY, color.RGBA{R: 189, G: 77, B: 39, A: 255})
}

func (s *StartScene) OnEnter() {}
func (s *StartScene) OnExit() {}