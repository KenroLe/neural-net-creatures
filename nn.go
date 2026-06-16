package main

import (
	"math"
	"math/rand"
)

type Layer struct {
	Weights [][]float64 // [out][in]
	Biases  []float64
}

type NeuralNet struct {
	Layers []Layer
	Sizes  []int
}

func NewNeuralNet(sizes []int) *NeuralNet {
	nn := &NeuralNet{Sizes: sizes}
	for i := 0; i < len(sizes)-1; i++ {
		in, out := sizes[i], sizes[i+1]
		scale := math.Sqrt(2.0 / float64(in))
		layer := Layer{
			Weights: make([][]float64, out),
			Biases:  make([]float64, out),
		}
		for j := 0; j < out; j++ {
			layer.Weights[j] = make([]float64, in)
			for k := 0; k < in; k++ {
				layer.Weights[j][k] = rand.NormFloat64() * scale
			}
		}
		nn.Layers = append(nn.Layers, layer)
	}
	return nn
}

func (nn *NeuralNet) Forward(input []float64) []float64 {
	current := input
	for _, layer := range nn.Layers {
		next := make([]float64, len(layer.Biases))
		for j := range next {
			sum := layer.Biases[j]
			for k, v := range current {
				sum += layer.Weights[j][k] * v
			}
			next[j] = math.Tanh(sum)
		}
		current = next
	}
	return current
}

func (nn *NeuralNet) Clone() *NeuralNet {
	clone := &NeuralNet{Sizes: nn.Sizes}
	for _, layer := range nn.Layers {
		cl := Layer{
			Biases:  make([]float64, len(layer.Biases)),
			Weights: make([][]float64, len(layer.Weights)),
		}
		copy(cl.Biases, layer.Biases)
		for j, w := range layer.Weights {
			cl.Weights[j] = make([]float64, len(w))
			copy(cl.Weights[j], w)
		}
		clone.Layers = append(clone.Layers, cl)
	}
	return clone
}

// Crossover: uniform gene mixing between two parents.
func Crossover(a, b *NeuralNet) *NeuralNet {
	child := a.Clone()
	for l := range child.Layers {
		for j := range child.Layers[l].Weights {
			if rand.Float64() < 0.5 {
				child.Layers[l].Biases[j] = b.Layers[l].Biases[j]
			}
			for k := range child.Layers[l].Weights[j] {
				if rand.Float64() < 0.5 {
					child.Layers[l].Weights[j][k] = b.Layers[l].Weights[j][k]
				}
			}
		}
	}
	return child
}

func (nn *NeuralNet) Mutate(rate, strength float64) {
	for l := range nn.Layers {
		for j := range nn.Layers[l].Weights {
			if rand.Float64() < rate {
				nn.Layers[l].Biases[j] += rand.NormFloat64() * strength
			}
			for k := range nn.Layers[l].Weights[j] {
				if rand.Float64() < rate {
					nn.Layers[l].Weights[j][k] += rand.NormFloat64() * strength
				}
			}
		}
	}
}
