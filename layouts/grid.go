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
	"encoding/json"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"io"
	"log"
	"math"
)

func VBox(config GridConfig, blocks ...*uik.Block) (l *Layouter) {
	g := NewGridEngine(config)
	l = NewLayouter(g)
	for i, b := range blocks {
		g.Add(b, GridComponent{
			GridX: 0, GridY: i,
			AnchorX: AnchorMin,
		})
	}
	return
}

func HBox(config GridConfig, blocks ...*uik.Block) (l *Layouter) {
	g := NewGridEngine(config)
	l = NewLayouter(g)
	for i, b := range blocks {
		g.Add(b, GridComponent{
			GridX: i, GridY: 0,
			AnchorY: AnchorMin,
		})
	}
	return
}

type Anchor uint8

const (
	AnchorMin Anchor = 1 << iota
	AnchorMax
)

type GridComponent struct {
	// The coordinates for the top-left of the block's placement
	GridX, GridY int
	// How many extra columns and rows the block occupies
	ExtraX, ExtraY int
	// AnchorX and AnchorY get the bit flags from AnchorMin and AnchorMax. The
	// zero value means they will float in the center.
	AnchorX, AnchorY Anchor
	// The zero-values for MinSize, PreferredSize and MaxSize tell the grid to ignore them
	MinSize, PreferredSize, MaxSize geom.Coord
}

type blockConfigPair struct {
	block  *uik.Block
	config GridComponent
}

type blockNamePair struct {
	block *uik.Block
	name  string
}

type removeBlock *uik.Block

type GridConfig struct {
	Components map[string]GridComponent
}

func ReadGridConfig(r io.Reader) (cfg GridConfig, err error) {
	dec := json.NewDecoder(r)
	err = dec.Decode(&cfg)
	return
}

type GridEngine struct {
	layouter *Layouter

	childrenHints          map[*uik.Block]uik.SizeHint
	childrenGridComponents map[*uik.Block]GridComponent
	config                 GridConfig

	vflex, hflex   *flex
	velems, helems map[*uik.Block]*elem

	configs chan interface{}
}

func NewGrid(cfg GridConfig) (l *Layouter) {
	g := NewGridEngine(cfg)
	l = NewLayouter(g)
	return
}

func NewGridEngine(config GridConfig) (g *GridEngine) {
	g = new(GridEngine)
	g.config = config

	g.childrenHints = make(map[*uik.Block]uik.SizeHint)
	g.childrenGridComponents = map[*uik.Block]GridComponent{}

	g.hflex = &flex{}
	g.vflex = &flex{}

	g.helems = map[*uik.Block]*elem{}
	g.velems = map[*uik.Block]*elem{}

	return
}

func (g *GridEngine) SetLayouter(layouter *Layouter) {
	g.layouter = layouter
}

func (g *GridEngine) Add(b *uik.Block, bd GridComponent) {
	g.layouter.Config(blockConfigPair{b, bd})
}

func (g *GridEngine) AddName(name string, b *uik.Block) {
	// uik.Report(g.layouter.ID, "add", name)
	g.layouter.Config(blockNamePair{b, name})
}

func (g *GridEngine) Remove(b *uik.Block) {
	g.layouter.Config(removeBlock(b))
}

func (g *GridEngine) SetConfig(cfg GridConfig) {
	g.layouter.Config(cfg)
}

func (g *GridEngine) addBlock(block *uik.Block, bd GridComponent) {

	g.layouter.AddBlock(block)
	g.childrenGridComponents[block] = bd

	helem := &elem{
		index: bd.GridX,
		extra: bd.ExtraX,
	}
	g.helems[block] = helem
	velem := &elem{
		index: bd.GridY,
		extra: bd.ExtraY,
	}
	g.velems[block] = velem

	g.hflex.add(helem)
	g.vflex.add(velem)
}

func (g *GridEngine) remBlock(b *uik.Block) {
	g.layouter.RemoveBlock(b)

	delete(g.childrenHints, b)
	delete(g.childrenGridComponents, b)

	g.hflex.rem(g.helems[b])
	g.vflex.rem(g.velems[b])

	g.layouter.Invalidate()
}

func (g *GridEngine) reflex(b *uik.Block) {
	bd := g.childrenGridComponents[b]
	sh := g.childrenHints[b]

	helem := g.helems[b]
	helem.minSize = sh.MinSize.X
	helem.prefSize = sh.PreferredSize.X
	helem.maxSize = math.Inf(1) //sh.MaxSize.X
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
	helem.fix()

	velem := g.velems[b]
	velem.minSize = sh.MinSize.Y
	velem.prefSize = sh.PreferredSize.Y
	velem.maxSize = math.Inf(1) //sh.MaxSize.Y
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
	velem.fix()
}

func (g *GridEngine) SetHint(block *uik.Block, hint uik.SizeHint) {
	g.childrenHints[block] = hint

	g.reflex(block)
	g.layouter.Invalidate()
}

func (g *GridEngine) GetHint() (hint uik.SizeHint) {
	hmin, hpref, hmax := g.hflex.makePrefs()
	vmin, vpref, vmax := g.vflex.makePrefs()
	hint.MinSize = geom.Coord{hmin, vmin}
	hint.PreferredSize = geom.Coord{hpref, vpref}
	hint.MaxSize = geom.Coord{hmax, vmax}

	return
}

func (g *GridEngine) GetLayout(size geom.Coord) (layout Layout) {

	layout = make(Layout)

	_, minXs, maxXs := g.hflex.constrain(size.X)
	_, minYs, maxYs := g.vflex.constrain(size.Y)

	// if g.Block.ID == 2 {
	// 	uik.Report("regrid", g.Block.ID, g.Size, whs, wvs)
	// }

	for child, csh := range g.childrenHints {
		bd := g.childrenGridComponents[child]
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
			switch bd.AnchorX {
			case 0:
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			case AnchorMin:
				gridBounds.Max.X -= diff
			case AnchorMax:
				gridBounds.Min.X += diff
			case AnchorMin | AnchorMax:
			}
		}
		if gridSizeY > csh.MaxSize.Y {
			diff := gridSizeY - csh.MaxSize.Y
			switch bd.AnchorY {
			case 0:
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			case AnchorMin:
				gridBounds.Max.Y -= diff
			case AnchorMax:
				gridBounds.Min.Y += diff
			case AnchorMin | AnchorMax:
			}
		}

		gridSizeX, gridSizeY = gridBounds.Size()
		if gridSizeX > csh.PreferredSize.X {
			diff := gridSizeX - csh.PreferredSize.X
			switch bd.AnchorX {
			case 0:
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			case AnchorMin:
				gridBounds.Max.X -= diff
			case AnchorMax:
				gridBounds.Min.X += diff
			case AnchorMin | AnchorMax:
			}
		}
		if gridSizeY > csh.PreferredSize.Y {
			diff := gridSizeY - csh.PreferredSize.Y
			switch bd.AnchorY {
			case 0:
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			case AnchorMin:
				gridBounds.Max.Y -= diff
			case AnchorMax:
				gridBounds.Min.Y += diff
			case AnchorMin | AnchorMax:
			}
		}

		layout[child] = gridBounds

	}

	return
}

func (g *GridEngine) ConfigUnsafe(cfg interface{}) {
	// uik.Report(g.layouter.ID, "cfg", cfg)
	switch cfg := cfg.(type) {
	case GridConfig:
		g.config = cfg
	case blockConfigPair:
		g.addBlock(cfg.block, cfg.config)
	case blockNamePair:
		componentConfig, ok := g.config.Components[cfg.name]
		if !ok {
			log.Print("GridEngine with Layouter", g.layouter.ID, ": unknown name", cfg.name)
			return
		}
		g.addBlock(cfg.block, componentConfig)
	}
}
