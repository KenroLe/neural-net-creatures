package main

import (
	"image/color"
	"math"
	"math/rand"
)

const (
	WorldWidth     = 1200
	WorldHeight    = 800
	MaxSpeed       = 3.0
	MaxEnergy      = 200.0
	InitEnergy     = 100.0
	ReproEnergy    = 160.0 // energy required to be eligible to reproduce
	ReproCost      = 70.0  // energy spent per parent when reproducing
	MoveCost       = 0.04  // energy per tick per speed unit
	IdleCost       = 0.015 // base energy drain per tick
	FoodEnergy     = 35.0
	ViewRadius     = 160.0
	CreatureRadius = 8.0
	MaxCooldown    = 300 // ticks before a creature can reproduce again
)

// Inputs to the neural net (must match NewNeuralNet sizes[0]).
// 0  nearest food distance (normalised 0-1, 1 = nothing visible)
// 1  sin of angle to nearest food (relative to heading)
// 2  cos of angle to nearest food
// 3  nearest creature distance (normalised)
// 4  sin of angle to nearest creature
// 5  cos of angle to nearest creature
// 6  energy ratio (0-1)
// 7  current speed (normalised)
// 8  want-to-reproduce flag (1 if cooldown==0 and energy>ReproEnergy)
// 9  bias constant
const NumInputs = 10

// Outputs (must match last layer size):
// 0  turn rate   (tanh → −1..1)
// 1  thrust      (tanh → −1..1, mapped to 0..1 speed)
// 2  reproduction desire (tanh > 0.5 triggers attempt)
const NumOutputs = 3

type Creature struct {
	X, Y          float64
	Angle         float64
	Speed         float64
	Energy        float64
	Age           int
	Brain         *NeuralNet
	R, G, B       float64 // genome colour 0-1
	ReproCooldown int
	Alive         bool
	Gen           int
	// Brief glow after reproducing (ticks remaining)
	GlowTicks int
}

func NewCreature(x, y float64) *Creature {
	return &Creature{
		X:      x,
		Y:      y,
		Angle:  rand.Float64() * 2 * math.Pi,
		Energy: InitEnergy,
		Brain:  NewNeuralNet([]int{NumInputs, 16, 12, NumOutputs}),
		R:      rand.Float64(),
		G:      rand.Float64(),
		B:      rand.Float64(),
		Alive:  true,
	}
}

func (c *Creature) Color() color.RGBA {
	alpha := uint8(180 + uint8(c.Energy/MaxEnergy*75))
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: alpha,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (c *Creature) sense(foods []*Food, creatures []*Creature) []float64 {
	nearFoodDist := ViewRadius
	nearFoodAngle := 0.0
	for _, f := range foods {
		if !f.Alive {
			continue
		}
		dx, dy := f.X-c.X, f.Y-c.Y
		d := math.Sqrt(dx*dx + dy*dy)
		if d < nearFoodDist {
			nearFoodDist = d
			nearFoodAngle = math.Atan2(dy, dx)
		}
	}

	nearCDist := ViewRadius
	nearCAngle := 0.0
	for _, o := range creatures {
		if o == c || !o.Alive {
			continue
		}
		dx, dy := o.X-c.X, o.Y-c.Y
		d := math.Sqrt(dx*dx + dy*dy)
		if d < nearCDist {
			nearCDist = d
			nearCAngle = math.Atan2(dy, dx)
		}
	}

	relFood := nearFoodAngle - c.Angle
	relC := nearCAngle - c.Angle

	wantRepro := 0.0
	if c.ReproCooldown == 0 && c.Energy >= ReproEnergy {
		wantRepro = 1.0
	}

	return []float64{
		nearFoodDist / ViewRadius,
		math.Sin(relFood),
		math.Cos(relFood),
		nearCDist / ViewRadius,
		math.Sin(relC),
		math.Cos(relC),
		c.Energy / MaxEnergy,
		c.Speed / MaxSpeed,
		wantRepro,
		1.0, // bias
	}
}

func (c *Creature) Update(foods []*Food, creatures []*Creature) {
	if !c.Alive {
		return
	}

	inputs := c.sense(foods, creatures)
	out := c.Brain.Forward(inputs)

	// out[0]: turn  (−1..1)
	// out[1]: thrust (mapped from tanh to 0..1)
	// out[2]: reproduce desire
	c.Angle += out[0] * 0.18
	thrust := (out[1] + 1) * 0.5
	c.Speed = thrust * MaxSpeed

	c.X += math.Cos(c.Angle) * c.Speed
	c.Y += math.Sin(c.Angle) * c.Speed

	// Wrap around the world edges.
	c.X = math.Mod(c.X+WorldWidth, WorldWidth)
	c.Y = math.Mod(c.Y+WorldHeight, WorldHeight)

	c.Energy -= IdleCost + c.Speed*MoveCost
	c.Age++

	if c.ReproCooldown > 0 {
		c.ReproCooldown--
	}
	if c.GlowTicks > 0 {
		c.GlowTicks--
	}

	if c.Energy <= 0 {
		c.Alive = false
	}
}

func (c *Creature) WantsToReproduce(out []float64) bool {
	return c.Energy >= ReproEnergy && c.ReproCooldown == 0 && out[2] > 0.5
}

func Reproduce(a, b *Creature) *Creature {
	childBrain := Crossover(a.Brain, b.Brain)
	childBrain.Mutate(0.08, 0.25)

	// Colour inherits from both parents with a small mutation.
	mutR := (rand.Float64() - 0.5) * 0.15
	mutG := (rand.Float64() - 0.5) * 0.15
	mutB := (rand.Float64() - 0.5) * 0.15

	child := &Creature{
		X:      (a.X+b.X)/2 + (rand.Float64()-0.5)*20,
		Y:      (a.Y+b.Y)/2 + (rand.Float64()-0.5)*20,
		Angle:  rand.Float64() * 2 * math.Pi,
		Energy: InitEnergy,
		Brain:  childBrain,
		R:      clamp01(a.R*0.5 + b.R*0.5 + mutR),
		G:      clamp01(a.G*0.5 + b.G*0.5 + mutG),
		B:      clamp01(a.B*0.5 + b.B*0.5 + mutB),
		Alive:  true,
		Gen:    max(a.Gen, b.Gen) + 1,
	}

	a.Energy -= ReproCost
	b.Energy -= ReproCost
	a.ReproCooldown = MaxCooldown
	b.ReproCooldown = MaxCooldown
	a.GlowTicks = 40
	b.GlowTicks = 40

	return child
}
