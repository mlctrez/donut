package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed donut.png
var donutPNG []byte

const (
	donutScale = 0.5 // Configuration: scale factor for the donut (1.0 = original size, 2.0 = double size, etc.)
)

type Game struct {
	donutImage   *ebiten.Image
	donutWidth   float64
	donutHeight  float64
	x, y         float64
	vx, vy       float64
	screenWidth  int
	screenHeight int
}

func (g *Game) Update() error {
	// Check for the escape key to exit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Update position
	g.x += g.vx
	g.y += g.vy

	// Bounce off edges
	if g.x <= 0 || g.x >= float64(g.screenWidth)-g.donutWidth {
		g.vx = -g.vx
		if g.x <= 0 {
			g.x = 0
		} else {
			g.x = float64(g.screenWidth) - g.donutWidth
		}
	}
	if g.y <= 0 || g.y >= float64(g.screenHeight)-g.donutHeight {
		g.vy = -g.vy
		if g.y <= 0 {
			g.y = 0
		} else {
			g.y = float64(g.screenHeight) - g.donutHeight
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{A: 255}) // Black background

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(donutScale, donutScale)
	op.GeoM.Translate(g.x, g.y)
	screen.DrawImage(g.donutImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Update screen dimensions when the window is resized
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func loadDonutImage() (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(donutPNG))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func main() {
	donutImage, err := loadDonutImage()
	if err != nil {
		log.Fatal("Failed to load donut.png:", err)
	}

	// Calculate scaled dimensions
	bounds := donutImage.Bounds()
	donutWidth := float64(bounds.Dx()) * donutScale
	donutHeight := float64(bounds.Dy()) * donutScale

	// Start with default dimensions - Layout method will update with actual window size
	game := &Game{
		donutImage:   donutImage,
		donutWidth:   donutWidth,
		donutHeight:  donutHeight,
		x:            100,
		y:            100,
		vx:           3,
		vy:           2,
		screenWidth:  800, // Default width, will be updated by Layout
		screenHeight: 600, // Default height, will be updated by Layout
	}

	// Don't set a specific window size - let it use the system default or fullscreen
	ebiten.SetWindowTitle("Donut Screensaver")
	ebiten.SetFullscreen(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
