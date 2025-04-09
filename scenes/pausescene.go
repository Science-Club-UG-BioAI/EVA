package scenes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

type PauseScene struct {
	loaded bool
}

func NewPauseScene() *PauseScene {
	return &PauseScene{
		loaded: false,
	}
}

// Draw implements Scene.
func (p *PauseScene) Draw(screen *ebiten.Image) {
	img, _, err := ebitenutil.NewImageFromFile("assets/images/start_background.png")
	if err != nil {
		ebitenutil.DebugPrint(screen, "Failed to load background image")
		return
	}
	screen.DrawImage(img, nil)

	// Draw "EVA" text
	evaText := "EVA"
	evaColor := color.White
	evaBounds := text.BoundString(basicfont.Face7x13, evaText)
	evaTextWidth := evaBounds.Dx()
	evaTextHeight := evaBounds.Dy()

	screenWidth, screenHeight := screen.Size()
	evaScaleFactor := 4.0
	evaX := (float64(screenWidth) - float64(evaTextWidth)*evaScaleFactor) / 2
	evaY := (float64(screenHeight) - float64(evaTextHeight)*evaScaleFactor) / 2 - 100

	evaOp := &ebiten.DrawImageOptions{}
	evaOp.GeoM.Scale(evaScaleFactor, evaScaleFactor)
	evaOp.GeoM.Translate(evaX, evaY)

	evaTextImage := ebiten.NewImage(evaTextWidth, evaTextHeight)
	text.Draw(evaTextImage, evaText, basicfont.Face7x13, 0, evaTextHeight, evaColor)
	screen.DrawImage(evaTextImage, evaOp)

	// styling for the pause text
	message := "Press escape to unpause"
	textColor := color.White
	bounds := text.BoundString(basicfont.Face7x13, message)
	textWidth := bounds.Dx()
	textHeight := bounds.Dy()

	scaleFactor := 2.0
	x := (float64(screenWidth) - float64(textWidth)*scaleFactor) / 2
	y := evaY + float64(evaTextHeight)*evaScaleFactor + 40

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleFactor, scaleFactor)
	op.GeoM.Translate(x, y)

	textImage := ebiten.NewImage(textWidth, textHeight)
	text.Draw(textImage, message, basicfont.Face7x13, 0, textHeight, textColor)
	screen.DrawImage(textImage, op)
}

// FirstLoad implements Scene.
func (p *PauseScene) FirstLoad() {
	p.loaded = true
}

// IsLoaded implements Scene.
func (p *PauseScene) IsLoaded() bool {
	return p.loaded
}

// OnEnter implements Scene.
func (p *PauseScene) OnEnter() {
}

// OnExit implements Scene.
func (p *PauseScene) OnExit() {
}

// Update implements Scene.
func (p *PauseScene) Update() SceneId {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return GameSceneId
	}

	return PauseSceneId
}

var _ Scene = (*PauseScene)(nil)
