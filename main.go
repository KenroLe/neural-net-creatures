package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	world         *World
	stepsPerFrame int
	paused        bool
	selected      *Creature
	prevMouseLeft bool
	prevKeyP      bool
	camX, camY    float64
}

func NewGame() *Game {
	return &Game{
		world:         NewWorld(),
		stepsPerFrame: 1,
	}
}

func (g *Game) Update() error {
	keyP := ebiten.IsKeyPressed(ebiten.KeyP)
	if keyP && !g.prevKeyP {
		g.paused = !g.paused
	}
	g.prevKeyP = keyP

	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		g.stepsPerFrame = min(g.stepsPerFrame+1, 10)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		g.stepsPerFrame = max(g.stepsPerFrame-1, 1)
	}

	// Camera movement.
	const camSpeed = 6.0
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camY -= camSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY += camSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= camSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += camSpeed
	}
	g.camX = math.Max(0, math.Min(g.camX, float64(WorldWidth-ViewWidth)))
	g.camY = math.Max(0, math.Min(g.camY, float64(WorldHeight-ViewHeight)))

	// Click to select a creature; click empty space to deselect.
	mouseLeft := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if mouseLeft && !g.prevMouseLeft {
		mx, my := ebiten.CursorPosition()
		wx, wy := float64(mx)+g.camX, float64(my)+g.camY
		g.selected = nil
		bestDist := CreatureRadius * 3
		for _, c := range g.world.Creatures {
			if !c.Alive {
				continue
			}
			dx := wx - c.X
			dy := wy - c.Y
			if d := math.Sqrt(dx*dx + dy*dy); d < bestDist {
				bestDist = d
				g.selected = c
			}
		}
	}
	g.prevMouseLeft = mouseLeft

	if g.selected != nil && !g.selected.Alive {
		g.selected = nil
	}

	if !g.paused {
		for i := 0; i < g.stepsPerFrame; i++ {
			g.world.Update()
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background.
	screen.Fill(color.RGBA{12, 12, 22, 255})

	// Draw food as small bright green dots.
	for _, f := range g.world.Foods {
		if !f.Alive {
			continue
		}
		sx, sy := float32(f.X-g.camX), float32(f.Y-g.camY)
		if sx < -10 || sx > ViewWidth+10 || sy < -10 || sy > ViewHeight+10 {
			continue
		}
		vector.DrawFilledCircle(screen, sx, sy, 3.5, color.RGBA{70, 210, 90, 220}, false)
	}

	// Draw creatures.
	for _, c := range g.world.Creatures {
		if !c.Alive {
			continue
		}
		drawCreature(screen, c, g.camX, g.camY)
	}

	// Selection ring.
	if g.selected != nil && g.selected.Alive {
		cx := float32(g.selected.X - g.camX)
		cy := float32(g.selected.Y - g.camY)
		vector.StrokeCircle(screen, cx, cy, float32(CreatureRadius)+6, 2, color.RGBA{255, 240, 80, 230}, false)
	}

	// HUD.
	alive := len(g.world.AliveCreatures())
	status := ""
	if g.paused {
		status = " [PAUSED]"
	}
	hud := fmt.Sprintf(
		"Tick: %d  Creatures: %d  Food: %d  Gen: %d  Born: %d  Speed: %dx  TPS: %.0f%s\n"+
			"[P] pause   [+/-] speed   [WASD] camera  cam:(%.0f,%.0f)",
		g.world.Tick,
		alive,
		g.world.AliveFoodCount(),
		g.world.MaxGen,
		g.world.BornTotal,
		g.stepsPerFrame,
		ebiten.ActualTPS(),
		status,
		g.camX, g.camY,
	)
	ebitenutil.DebugPrint(screen, hud)

	drawNNPanel(screen, g.selected)
}

func drawCreature(screen *ebiten.Image, c *Creature, camX, camY float64) {
	cx, cy := float32(c.X-camX), float32(c.Y-camY)
	if cx < -30 || cx > ViewWidth+30 || cy < -30 || cy > ViewHeight+30 {
		return
	}
	r := float32(CreatureRadius)

	energyRatio := float32(c.Energy / MaxEnergy)

	// Outer energy ring — dim version of creature colour, sized by energy.
	ringR := r*0.4 + r*1.4*energyRatio
	ringCol := color.RGBA{
		R: uint8(c.R * 80),
		G: uint8(c.G * 80),
		B: uint8(c.B * 80),
		A: 60,
	}
	vector.DrawFilledCircle(screen, cx, cy, ringR, ringCol, false)

	// Glow when reproducing recently.
	if c.GlowTicks > 0 {
		glowAlpha := uint8(float32(c.GlowTicks) / float32(40) * 180)
		glowCol := color.RGBA{255, 255, 200, glowAlpha}
		vector.DrawFilledCircle(screen, cx, cy, r+5, glowCol, false)
	}

	// Main body.
	bodyCol := c.Color()
	vector.DrawFilledCircle(screen, cx, cy, r, bodyCol, false)

	// White outline for readability.
	vector.StrokeCircle(screen, cx, cy, r, 1, color.RGBA{255, 255, 255, 60}, false)

	// Direction indicator.
	headX := cx + float32(math.Cos(c.Angle))*r*1.4
	headY := cy + float32(math.Sin(c.Angle))*r*1.4
	vector.StrokeLine(screen, cx, cy, headX, headY, 2, color.RGBA{255, 255, 255, 180}, false)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return ViewWidth, ViewHeight
}

func main() {
	ebiten.SetWindowSize(ViewWidth, ViewHeight)
	ebiten.SetWindowTitle("Neural Net Creatures")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
