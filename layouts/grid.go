/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package layouts

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image/color"
	"image/draw"
	"math"
)

func VBox(config GridConfig, blocks ...*uik.Block) (g *Grid) {
	g = NewGrid(config)
	for i, b := range blocks {
		g.Add <- BlockData{
			Block: b,
			GridX: 0, GridY: i,
			AnchorX: AnchorMin,
		}
	}
	return
}

func HBox(config GridConfig, blocks ...*uik.Block) (g *Grid) {
	g = NewGrid(config)
	for i, b := range blocks {
		g.Add <- BlockData{
			Block: b,
			GridX: i, GridY: 0,
			AnchorY: AnchorMin,
		}
	}
	return
}

type Anchor uint8

const (
	AnchorMin Anchor = 1 << iota
	AnchorMax
)

type BlockData struct {
	Block                           *uik.Block
	GridX, GridY                    int
	ExtraX, ExtraY                  int
	AnchorX, AnchorY                Anchor
	MinSize, PreferredSize, MaxSize geom.Coord
}

type GridConfig struct {
}

type Grid struct {
	uik.Foundation

	children          map[*uik.Block]bool
	childrenBlockData map[*uik.Block]BlockData
	config            GridConfig

	vflex, hflex *flex

	Add       chan<- BlockData
	add       chan BlockData
	Remove    chan<- *uik.Block
	remove    chan *uik.Block
	SetConfig chan<- GridConfig
	setConfig chan GridConfig
	GetConfig <-chan GridConfig
	getConfig chan GridConfig
}

func NewGrid(cfg GridConfig) (g *Grid) {
	g = new(Grid)

	g.config = cfg

	g.Initialize()
	if uik.ReportIDs {
		uik.Report(g.ID, "grid")
	}

	go g.handleEvents()

	return
}

func (g *Grid) Initialize() {
	g.Foundation.Initialize()
	g.DrawOp = draw.Over

	g.children = map[*uik.Block]bool{}
	g.childrenBlockData = map[*uik.Block]BlockData{}

	g.add = make(chan BlockData, 1)
	g.Add = g.add
	g.remove = make(chan *uik.Block, 1)
	g.Remove = g.remove
	g.setConfig = make(chan GridConfig, 1)
	g.SetConfig = g.setConfig
	g.getConfig = make(chan GridConfig, 1)
	g.GetConfig = g.getConfig

	// g.Paint = func(gc draw2d.GraphicContext) {
	// 	g.draw(gc)
	// }
}

func (g *Grid) addBlock(bd BlockData) {
	g.AddBlock(bd.Block)
	g.children[bd.Block] = true
	g.childrenBlockData[bd.Block] = bd
	g.vflex = nil
	g.regrid()
}

func (g *Grid) remBlock(b *uik.Block) {
	if !g.children[b] {
		return
	}
	delete(g.ChildrenHints, b)
	delete(g.childrenBlockData, b)
	g.vflex = nil
	g.regrid()
}

func (g *Grid) reflex() {
	if g.vflex != nil {
		return
	}
	g.hflex = &flex{}
	g.vflex = &flex{}
	for _, bd := range g.childrenBlockData {
		csh := g.ChildrenHints[bd.Block]

		helem := elem{
			index:    bd.GridX,
			extra:    bd.ExtraX,
			minSize:  csh.MinSize.X,
			prefSize: csh.PreferredSize.X,
			maxSize:  math.Inf(1),
		}
		if bd.MinSize.X != 0 {
			helem.minSize = math.Max(bd.MinSize.X, helem.minSize)
		}
		if bd.PreferredSize.X != 0 {
			helem.prefSize = bd.PreferredSize.X
		}
		if bd.MaxSize.X != 0 {
			helem.maxSize = math.Min(bd.MaxSize.X, helem.maxSize)
		}
		helem.prefSize = math.Min(helem.maxSize, math.Max(helem.minSize, helem.prefSize))
		g.hflex.add(&helem)

		velem := elem{
			index:    bd.GridY,
			extra:    bd.ExtraY,
			minSize:  csh.MinSize.Y,
			prefSize: csh.PreferredSize.Y,
			maxSize:  math.Inf(1),
		}
		if bd.MinSize.Y != 0 {
			velem.minSize = math.Max(bd.MinSize.Y, velem.minSize)
		}
		if bd.PreferredSize.Y != 0 {
			velem.prefSize = bd.PreferredSize.Y
		}
		if bd.MaxSize.Y != 0 {
			velem.maxSize = math.Min(bd.MaxSize.Y, velem.maxSize)
		}
		velem.prefSize = math.Min(velem.maxSize, math.Max(velem.minSize, velem.prefSize))
		g.vflex.add(&velem)
	}
}

func (g *Grid) makePreferences() {
	var sizeHint uik.SizeHint
	g.reflex()
	hmin, hpref, hmax := g.hflex.makePrefs()
	vmin, vpref, vmax := g.vflex.makePrefs()
	sizeHint.MinSize = geom.Coord{hmin, vmin}
	sizeHint.PreferredSize = geom.Coord{hpref, vpref}
	sizeHint.MaxSize = geom.Coord{hmax, vmax}
	g.SetSizeHint(sizeHint)
}

func (g *Grid) regrid() {

	g.reflex()

	_, minXs, maxXs := g.hflex.constrain(g.Size.X)
	_, minYs, maxYs := g.vflex.constrain(g.Size.Y)

	for child, csh := range g.ChildrenHints {
		bd := g.childrenBlockData[child]
		gridBounds := geom.Rect{
			Min: geom.Coord{
				X: minXs[bd.GridX],
				Y: minYs[bd.GridY],
			},
			Max: geom.Coord{
				X: maxXs[bd.GridX+bd.ExtraX],
				Y: maxYs[bd.GridY+bd.ExtraY],
			},
		}

		gridSizeX, gridSizeY := gridBounds.Size()
		if gridSizeX > csh.MaxSize.X {
			diff := gridSizeX - csh.MaxSize.X
			if bd.AnchorX&AnchorMin != 0 && bd.AnchorX&AnchorMax != 0 {
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			} else if bd.AnchorX&AnchorMin != 0 {
				gridBounds.Max.X -= diff
			} else if bd.AnchorX&AnchorMax != 0 {
				gridBounds.Min.X += diff
			}
		}
		if gridSizeY > csh.MaxSize.Y {
			diff := gridSizeY - csh.MaxSize.Y
			if bd.AnchorY&AnchorMin == 0 && bd.AnchorY&AnchorMax == 0 {
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			} else if bd.AnchorY&AnchorMin != 0 {
				gridBounds.Max.Y -= diff
			} else if bd.AnchorY&AnchorMax != 0 {
				gridBounds.Min.Y += diff
			}
		}

		gridSizeX, gridSizeY = gridBounds.Size()
		if gridSizeX > csh.PreferredSize.X {
			diff := gridSizeX - csh.PreferredSize.X
			if bd.AnchorX&AnchorMin != 0 && bd.AnchorX&AnchorMax != 0 {
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			} else if bd.AnchorX&AnchorMin != 0 {
				gridBounds.Max.X -= diff
			} else if bd.AnchorX&AnchorMax != 0 {
				gridBounds.Min.X += diff
			}
		}
		if gridSizeY > csh.PreferredSize.Y {
			diff := gridSizeY - csh.PreferredSize.Y
			if bd.AnchorY&AnchorMin == 0 && bd.AnchorY&AnchorMax == 0 {
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			} else if bd.AnchorY&AnchorMin != 0 {
				gridBounds.Max.Y -= diff
			} else if bd.AnchorY&AnchorMax != 0 {
				gridBounds.Min.Y += diff
			}
		}

		g.ChildrenBounds[child] = gridBounds

		gridSizeX, gridSizeY = gridBounds.Size()
		child.UserEventsIn <- uik.ResizeEvent{
			Size: geom.Coord{gridSizeX, gridSizeY},
		}
	}

	g.Invalidate()
}

func safeRect(path draw2d.GraphicContext, min, max geom.Coord) {
	x1, y1 := min.X, min.Y
	x2, y2 := max.X, max.Y
	x, y := path.LastPoint()
	path.MoveTo(x1, y1)
	path.LineTo(x2, y1)
	path.LineTo(x2, y2)
	path.LineTo(x1, y2)
	path.Close()
	path.MoveTo(x, y)
}

func (g *Grid) draw(gc draw2d.GraphicContext) {
	gc.Clear()
	gc.SetFillColor(color.RGBA{150, 150, 150, 255})
	safeRect(gc, geom.Coord{0, 0}, g.Size)
	gc.FillStroke()

	g.reflex()

	_, minXs, _ := g.hflex.constrain(g.Size.X)
	for _, x := range minXs[1:] {
		gc.MoveTo(x, 0)
		gc.LineTo(x, g.Size.Y)

	}
	_, minYs, _ := g.vflex.constrain(g.Size.Y)
	for _, y := range minYs[1:] {
		gc.MoveTo(0, y)
		gc.LineTo(g.Size.X, y)

	}
	gc.Stroke()
	// _, _, maxYs := g.vflex.constrain(g.Size.Y)
}

func (g *Grid) handleEvents() {
	for {
		select {
		case e := <-g.UserEvents:
			switch e := e.(type) {
			case uik.ResizeEvent:
				g.Size = e.Size
				g.regrid()
				g.Invalidate()
			default:
				g.Foundation.HandleEvent(e)
			}
		case bsh := <-g.BlockSizeHints:
			if !g.children[bsh.Block] {
				// Do I know you?
				break
			}
			g.ChildrenHints[bsh.Block] = bsh.SizeHint

			g.vflex = nil

			g.makePreferences()
			g.regrid()
		case e := <-g.BlockInvalidations:
			g.DoBlockInvalidation(e)
			// go uik.ShowBuffer("grid", g.Buffer)
		case g.config = <-g.setConfig:
			g.vflex = nil
			g.makePreferences()
		case g.getConfig <- g.config:
		case bd := <-g.add:
			g.addBlock(bd)
			g.vflex = nil
			g.makePreferences()
		case b := <-g.remove:
			g.remBlock(b)
			g.vflex = nil
			g.makePreferences()
		}
	}
}
