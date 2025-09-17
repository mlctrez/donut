package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed donut.png
var donutPNG []byte

const (
	donutScale    = 0.5 // Configuration: scale factor for the donut (1.0 = original size, 2.0 = double size, etc.)
	initialDonuts = 6   // Configuration: initial number of donuts to display
	maxDonuts     = 50  // Maximum number of donuts allowed
	minDonuts     = 1   // Minimum number of donuts allowed
)

type Donut struct {
	x, y          float64
	vx, vy        float64
	rotation      float64 // Rotation angle in radians
	rotationSpeed float64 // Rotation speed in radians per frame
}

type Game struct {
	donutImage   *ebiten.Image
	donutWidth   float64
	donutHeight  float64
	donuts       []Donut
	screenWidth  int
	screenHeight int
	numDonuts    int // Current number of donuts
}

func (g *Game) Update() error {
	// Check for the escape key to exit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Handle plus key to add more donuts
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		if g.numDonuts < maxDonuts {
			g.numDonuts++
			g.donuts = createDonuts(g.screenWidth, g.screenHeight, g.donutWidth, g.donutHeight, g.numDonuts)
		}
	}

	// Handle minus key to remove donuts
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		if g.numDonuts > minDonuts {
			g.numDonuts--
			g.donuts = createDonuts(g.screenWidth, g.screenHeight, g.donutWidth, g.donutHeight, g.numDonuts)
		}
	}

	// Update each donut
	for i := range g.donuts {
		donut := &g.donuts[i]

		// Update position
		donut.x += donut.vx
		donut.y += donut.vy

		// Update rotation
		donut.rotation += donut.rotationSpeed

		// Bounce off edges
		if donut.x <= 0 || donut.x >= float64(g.screenWidth)-g.donutWidth {
			donut.vx = -donut.vx
			if donut.x <= 0 {
				donut.x = 0
			} else {
				donut.x = float64(g.screenWidth) - g.donutWidth
			}
		}
		if donut.y <= 0 || donut.y >= float64(g.screenHeight)-g.donutHeight {
			donut.vy = -donut.vy
			if donut.y <= 0 {
				donut.y = 0
			} else {
				donut.y = float64(g.screenHeight) - g.donutHeight
			}
		}
	}

	// Check for collisions between donuts
	g.handleDonutCollisions()

	return nil
}

// handleDonutCollisions checks for and resolves collisions between donuts
func (g *Game) handleDonutCollisions() {
	radius := g.donutWidth / 2 // Assuming width == height for circular donuts

	for i := 0; i < len(g.donuts); i++ {
		for j := i + 1; j < len(g.donuts); j++ {
			donut1 := &g.donuts[i]
			donut2 := &g.donuts[j]

			// Calculate center positions
			center1X := donut1.x + radius
			center1Y := donut1.y + radius
			center2X := donut2.x + radius
			center2Y := donut2.y + radius

			// Check if donuts are colliding
			if g.areDonutsColliding(center1X, center1Y, center2X, center2Y, radius) {
				g.resolveCollision(donut1, donut2, center1X, center1Y, center2X, center2Y)
			}
		}
	}
}

// areDonutsColliding checks if two circular donuts are overlapping
func (g *Game) areDonutsColliding(x1, y1, x2, y2, radius float64) bool {
	dx := x2 - x1
	dy := y2 - y1
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < (radius * 2) // Two circles collide when distance < sum of radii
}

// resolveCollision handles the physics of two donuts colliding
func (g *Game) resolveCollision(donut1, donut2 *Donut, center1X, center1Y, center2X, center2Y float64) {
	// Calculate collision vector
	dx := center2X - center1X
	dy := center2Y - center1Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Avoid division by zero
	if distance == 0 {
		dx = 1
		dy = 0
		distance = 1
	}

	// Normalize collision vector
	nx := dx / distance
	ny := dy / distance

	// Separate the donuts so they don't overlap
	radius := g.donutWidth / 2
	overlap := (radius * 2) - distance
	separationX := nx * overlap * 0.5
	separationY := ny * overlap * 0.5

	donut1.x -= separationX
	donut1.y -= separationY
	donut2.x += separationX
	donut2.y += separationY

	// Calculate relative velocity
	dvx := donut2.vx - donut1.vx
	dvy := donut2.vy - donut1.vy

	// Calculate relative velocity along collision normal
	dvn := dvx*nx + dvy*ny

	// Don't resolve if velocities are separating
	if dvn > 0 {
		return
	}

	// Collision impulse (assuming equal mass and elastic collision)
	impulse := 2 * dvn / 2 // divided by 2 because we have 2 objects of equal mass

	// Update velocities
	donut1.vx += impulse * nx
	donut1.vy += impulse * ny
	donut2.vx -= impulse * nx
	donut2.vy -= impulse * ny
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{A: 255}) // Black background

	// Draw each donut
	for _, donut := range g.donuts {
		op := &ebiten.DrawImageOptions{}

		// Apply transformations in the correct order for rotation around center:
		// 1. Scale the image
		op.GeoM.Scale(donutScale, donutScale)

		// 2. Translate to center the rotation point (move origin to center of scaled image)
		op.GeoM.Translate(-g.donutWidth/2, -g.donutHeight/2)

		// 3. Rotate around the origin (which is now at the center)
		op.GeoM.Rotate(donut.rotation)

		// 4. Translate back and then to final position
		op.GeoM.Translate(g.donutWidth/2, g.donutHeight/2)
		op.GeoM.Translate(donut.x, donut.y)

		screen.DrawImage(g.donutImage, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Update screen dimensions when the window is resized
	if g.screenWidth != outsideWidth || g.screenHeight != outsideHeight {
		g.screenWidth = outsideWidth
		g.screenHeight = outsideHeight
		// Recreate donuts with new screen dimensions
		g.donuts = createDonuts(outsideWidth, outsideHeight, g.donutWidth, g.donutHeight, g.numDonuts)
	}
	return outsideWidth, outsideHeight
}

func loadDonutImage() (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(donutPNG))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func createDonuts(screenWidth, screenHeight int, donutWidth, donutHeight float64, numDonuts int) []Donut {
	donuts := make([]Donut, numDonuts)

	// Define the center area where donuts will spawn (middle 50% of screen)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	spawnRadius := math.Min(float64(screenWidth), float64(screenHeight)) * 0.25

	for i := 0; i < numDonuts; i++ {
		// Random position near center
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * spawnRadius
		x := centerX + math.Cos(angle)*distance - donutWidth/2
		y := centerY + math.Sin(angle)*distance - donutHeight/2

		// Ensure donuts stay within screen bounds
		if x < 0 {
			x = 0
		} else if x > float64(screenWidth)-donutWidth {
			x = float64(screenWidth) - donutWidth
		}
		if y < 0 {
			y = 0
		} else if y > float64(screenHeight)-donutHeight {
			y = float64(screenHeight) - donutHeight
		}

		// Random velocity with consistent dx/dy components like the original
		// Generate random vx and vy independently to ensure good movement in both directions
		vx := 1.5 + rand.Float64()*3.0 // Between 1.5 and 4.5
		vy := 1.5 + rand.Float64()*3.0 // Between 1.5 and 4.5

		// Randomly make velocities negative to get different directions
		if rand.Float64() < 0.5 {
			vx = -vx
		}
		if rand.Float64() < 0.5 {
			vy = -vy
		}

		// Alternating rotation direction (clockwise vs counter-clockwise)
		rotationSpeed := 0.015 + rand.Float64()*0.02 // Base speed with some variation
		if i%2 == 1 {
			rotationSpeed = -rotationSpeed // Counter-clockwise for every other donut
		}

		donuts[i] = Donut{
			x:             x,
			y:             y,
			vx:            vx,
			vy:            vy,
			rotation:      rand.Float64() * 2 * math.Pi, // Random starting rotation
			rotationSpeed: rotationSpeed,
		}
	}

	return donuts
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
	screenWidth, screenHeight := 800, 600 // Default dimensions

	game := &Game{
		donutImage:   donutImage,
		donutWidth:   donutWidth,
		donutHeight:  donutHeight,
		donuts:       createDonuts(screenWidth, screenHeight, donutWidth, donutHeight, initialDonuts),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		numDonuts:    initialDonuts,
	}

	// Don't set a specific window size - let it use the system default or fullscreen
	ebiten.SetWindowTitle("Donut Screensaver")
	ebiten.SetFullscreen(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
