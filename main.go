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
	world *World
	// Speed multiplier: sim steps per draw frame.
	stepsPerFrame int
	paused        bool
}

func NewGame() *Game {
	return &Game{
		world:         NewWorld(),
		stepsPerFrame: 1,
	}
}

func (g *Game) Update() error {
	// P toggles pause.
	if ebiten.IsKeyPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}
	// + / - adjust simulation speed.
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		g.stepsPerFrame = min(g.stepsPerFrame+1, 10)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		g.stepsPerFrame = max(g.stepsPerFrame-1, 1)
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
		vector.DrawFilledCircle(screen, float32(f.X), float32(f.Y), 3.5, color.RGBA{70, 210, 90, 220}, false)
	}

	// Draw creatures.
	for _, c := range g.world.Creatures {
		if !c.Alive {
			continue
		}
		drawCreature(screen, c)
	}

	// HUD.
	alive := len(g.world.AliveCreatures())
	status := ""
	if g.paused {
		status = " [PAUSED]"
	}
	hud := fmt.Sprintf(
		"Tick: %d  Creatures: %d  Food: %d  Gen: %d  Born: %d  Speed: %dx  TPS: %.0f%s\n"+
			"[P] pause   [+/-] speed",
		g.world.Tick,
		alive,
		g.world.AliveFoodCount(),
		g.world.MaxGen,
		g.world.BornTotal,
		g.stepsPerFrame,
		ebiten.ActualTPS(),
		status,
	)
	ebitenutil.DebugPrint(screen, hud)
}

func drawCreature(screen *ebiten.Image, c *Creature) {
	cx, cy := float32(c.X), float32(c.Y)
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
	return WorldWidth, WorldHeight
}

func main() {
	ebiten.SetWindowSize(WorldWidth, WorldHeight)
	ebiten.SetWindowTitle("Neural Net Creatures")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
