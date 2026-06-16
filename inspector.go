package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	panelW    = 400
	panelH    = 740
	inspNodeR = float32(7)
)

var inputLabels = []string{
	"food d", "food s", "food c",
	"mate d", "mate s", "mate c",
	"energy", "speed ", "repro?", "bias  ",
}

var outputLabels = []string{"turn", "thrust", "repro", "mitos"}

func drawNNPanel(screen *ebiten.Image, c *Creature) {
	if c == nil || !c.Alive || c.LastActivations == nil {
		return
	}

	px := float32(ViewWidth - panelW - 8)
	py := float32(30)
	pw := float32(panelW)
	ph := float32(panelH)

	vector.DrawFilledRect(screen, px, py, pw, ph, color.RGBA{8, 8, 22, 225}, false)
	vector.StrokeRect(screen, px, py, pw, ph, 1, color.RGBA{60, 80, 140, 200}, false)

	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf(" Neural Net  Gen:%d  Age:%d  E:%.0f", c.Gen, c.Age, c.Energy),
		int(px), int(py)+2)
	vector.StrokeLine(screen, px, py+14, px+pw, py+14, 1, color.RGBA{60, 80, 140, 160}, false)

	nn := c.Brain
	sizes := nn.Sizes
	numLayers := len(sizes)
	acts := c.LastActivations

	leftPad := float32(68)
	rightPad := float32(52)
	topPad := float32(22)
	contentW := pw - leftPad - rightPad
	contentH := ph - topPad - 14
	contentX := px + leftPad
	contentY := py + topPad + 4

	layerXs := make([]float32, numLayers)
	for i := range layerXs {
		layerXs[i] = contentX + float32(i)*contentW/float32(numLayers-1)
	}

	maxNodes := 0
	for _, s := range sizes {
		if s > maxNodes {
			maxNodes = s
		}
	}
	nodeSpacing := float32(math.Min(float64(contentH)/float64(maxNodes), 22))

	type pt struct{ x, y float32 }
	nodePos := make([][]pt, numLayers)
	for l, s := range sizes {
		nodePos[l] = make([]pt, s)
		totalH := float32(s-1) * nodeSpacing
		startY := contentY + contentH/2 - totalH/2
		for n := 0; n < s; n++ {
			nodePos[l][n] = pt{layerXs[l], startY + float32(n)*nodeSpacing}
		}
	}

	// Connections
	for l := 0; l < len(nn.Layers); l++ {
		layer := nn.Layers[l]
		for j := range layer.Weights {
			for k := range layer.Weights[j] {
				w := layer.Weights[j][k]
				abs := math.Abs(w)
				if abs < 0.08 {
					continue
				}
				alpha := uint8(math.Min(abs*75, 65))
				from := nodePos[l][k]
				to := nodePos[l+1][j]
				var col color.RGBA
				if w > 0 {
					col = color.RGBA{60, 190, 100, alpha}
				} else {
					col = color.RGBA{210, 60, 60, alpha}
				}
				vector.StrokeLine(screen, from.x, from.y, to.x, to.y, 1, col, false)
			}
		}
	}

	// Nodes
	for l, s := range sizes {
		for n := 0; n < s; n++ {
			pos := nodePos[l][n]
			var act float64
			if l < len(acts) && n < len(acts[l]) {
				act = acts[l][n]
			}
			vector.DrawFilledCircle(screen, pos.x, pos.y, inspNodeR, activationColor(act), false)
			vector.StrokeCircle(screen, pos.x, pos.y, inspNodeR, 1, color.RGBA{160, 160, 200, 100}, false)
		}
	}

	// Input labels
	for n, label := range inputLabels {
		if n >= len(nodePos[0]) {
			break
		}
		pos := nodePos[0][n]
		act := 0.0
		if len(acts) > 0 && n < len(acts[0]) {
			act = acts[0][n]
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%.2f", label, act),
			int(px)+2, int(pos.y)-4)
	}

	// Output labels
	for n, label := range outputLabels {
		if n >= len(nodePos[numLayers-1]) {
			break
		}
		pos := nodePos[numLayers-1][n]
		act := 0.0
		if len(acts) > 0 && n < len(acts[numLayers-1]) {
			act = acts[numLayers-1][n]
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s %.2f", label, act),
			int(pos.x)+int(inspNodeR)+3, int(pos.y)-4)
	}

	ebitenutil.DebugPrintAt(screen, "click empty space to deselect",
		int(px)+2, int(py+ph)-12)
}

func activationColor(v float64) color.RGBA {
	v = math.Max(-1, math.Min(1, v))
	if v >= 0 {
		g := uint8(40 + v*215)
		return color.RGBA{35, g, 55, 255}
	}
	r := uint8(40 + (-v)*215)
	return color.RGBA{r, 35, 55, 255}
}
