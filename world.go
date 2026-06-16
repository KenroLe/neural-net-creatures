package main

import (
	"math"
	"math/rand"
)

const (
	MaxFood       = 180
	MaxCreatures  = 400
	InitCreatures = 40
	FoodPerTick   = 2 // new food items spawned each tick
	MinPopulation = 8 // repopulate with random creatures if below this
)

type Food struct {
	X, Y  float64
	Alive bool
}

type World struct {
	Creatures []*Creature
	Foods     []*Food
	Tick      int

	// Stats
	MaxGen    int
	BornTotal int
	DiedTotal int
}

func NewWorld() *World {
	w := &World{}

	for i := 0; i < InitCreatures; i++ {
		c := NewCreature(rand.Float64()*WorldWidth, rand.Float64()*WorldHeight)
		w.Creatures = append(w.Creatures, c)
	}

	for i := 0; i < MaxFood/2; i++ {
		w.Foods = append(w.Foods, &Food{
			X:     rand.Float64() * WorldWidth,
			Y:     rand.Float64() * WorldHeight,
			Alive: true,
		})
	}

	w.BornTotal = InitCreatures
	return w
}

func (w *World) Update() {
	w.Tick++

	// Spawn food up to cap.
	foodAlive := 0
	for _, f := range w.Foods {
		if f.Alive {
			foodAlive++
		}
	}
	for i := 0; i < FoodPerTick && foodAlive < MaxFood; i++ {
		w.Foods = append(w.Foods, &Food{
			X:     rand.Float64() * WorldWidth,
			Y:     rand.Float64() * WorldHeight,
			Alive: true,
		})
		foodAlive++
	}

	alive := w.AliveCreatures()

	// Update all creatures (sense + move).
	for _, c := range alive {
		c.Update(w.Foods, alive)
	}

	// Eat food (squared-distance check avoids sqrt).
	eatRadius2 := (CreatureRadius + 5) * (CreatureRadius + 5)
	for _, c := range alive {
		if !c.Alive {
			continue
		}
		for _, f := range w.Foods {
			if !f.Alive {
				continue
			}
			dx, dy := f.X-c.X, f.Y-c.Y
			if dx*dx+dy*dy < eatRadius2 {
				f.Alive = false
				c.Energy = math.Min(c.Energy+FoodEnergy, MaxEnergy)
			}
		}
	}

	// Reproduction: O(n²) proximity check.
	var newborns []*Creature
	for i := 0; i < len(alive); i++ {
		a := alive[i]
		if !a.Alive {
			continue
		}
		aOut := a.Brain.Forward(a.sense(w.Foods, alive))
		if !a.WantsToReproduce(aOut) {
			continue
		}
		for j := i + 1; j < len(alive); j++ {
			b := alive[j]
			if !b.Alive {
				continue
			}
			bOut := b.Brain.Forward(b.sense(w.Foods, alive))
			if !b.WantsToReproduce(bOut) {
				continue
			}
			dx, dy := b.X-a.X, b.Y-a.Y
			if dx*dx+dy*dy < (CreatureRadius*3)*(CreatureRadius*3) {
				if len(alive)+len(newborns) < MaxCreatures {
					child := Reproduce(a, b)
					newborns = append(newborns, child)
					w.BornTotal++
					if child.Gen > w.MaxGen {
						w.MaxGen = child.Gen
					}
				}
			}
		}
	}
	w.Creatures = append(w.Creatures, newborns...)

	// Count deaths and compact dead entries periodically.
	if w.Tick%200 == 0 {
		var live []*Creature
		for _, c := range w.Creatures {
			if c.Alive {
				live = append(live, c)
			} else {
				w.DiedTotal++
			}
		}
		w.Creatures = live

		var liveF []*Food
		for _, f := range w.Foods {
			if f.Alive {
				liveF = append(liveF, f)
			}
		}
		w.Foods = liveF
	}

	// Emergency repopulation to keep the sim running.
	if len(alive)+len(newborns) < MinPopulation {
		for i := 0; i < InitCreatures/2; i++ {
			c := NewCreature(rand.Float64()*WorldWidth, rand.Float64()*WorldHeight)
			w.Creatures = append(w.Creatures, c)
			w.BornTotal++
		}
	}
}

func (w *World) AliveCreatures() []*Creature {
	var out []*Creature
	for _, c := range w.Creatures {
		if c.Alive {
			out = append(out, c)
		}
	}
	return out
}

func (w *World) AliveFoodCount() int {
	n := 0
	for _, f := range w.Foods {
		if f.Alive {
			n++
		}
	}
	return n
}
