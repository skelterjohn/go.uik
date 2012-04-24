package layouts

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image/color"
	"image/draw"
	"math"
)

type Anchor float64

const (
	AnchorMin Anchor = 1 << iota
	AnchorMax
)

type BlockData struct {
	Block          *uik.Block
	GridX, GridY   int
	ExtraX, ExtraY int
}

type GridConfig struct {
}

type table struct {
	data map[int]map[int]BlockData
}

func (t *table) get(x, y int) (bd BlockData, ok bool) {
	if t.data == nil {
		ok = false
		return
	}
	col, ok := t.data[x]
	if !ok {
		return
	}
	bd, ok = col[y]
	return
}

func (t *table) set(x, y int, bd BlockData) {
	if t.data == nil {
		t.data = map[int]map[int]BlockData{}
	}
	col := t.data[x]
	if col == nil {
		col = map[int]BlockData{}
		t.data[x] = col
	}
	col[y] = bd
}

type Grid struct {
	uik.Foundation

	children          map[*uik.Block]bool
	childrenSizeHints map[*uik.Block]uik.SizeHint
	childrenBlockData map[*uik.Block]BlockData
	config            GridConfig

	table table

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

	go g.handleEvents()

	return
}

func (g *Grid) Initialize() {
	g.Foundation.Initialize()
	g.DrawOp = draw.Over

	g.children = map[*uik.Block]bool{}
	g.childrenSizeHints = map[*uik.Block]uik.SizeHint{}
	g.childrenBlockData = map[*uik.Block]BlockData{}

	g.add = make(chan BlockData, 1)
	g.Add = g.add
	g.remove = make(chan *uik.Block, 1)
	g.Remove = g.remove
	g.setConfig = make(chan GridConfig, 1)
	g.SetConfig = g.setConfig
	g.getConfig = make(chan GridConfig, 1)
	g.GetConfig = g.getConfig

	g.Paint = func(gc draw2d.GraphicContext) {
		g.draw(gc)
	}
}

func (g *Grid) addBlock(bd BlockData) {
	g.AddBlock(bd.Block)
	g.children[bd.Block] = true
	g.childrenBlockData[bd.Block] = bd
	g.regrid()
}

func (g *Grid) remBlock(b *uik.Block) {
	if !g.children[b] {
		return
	}
	delete(g.childrenSizeHints, b)
	delete(g.childrenBlockData, b)
	g.regrid()
}

func (g *Grid) makePreferences() {
	var sizeHint uik.SizeHint

	maxX, maxY := 0, 0
	for _, bd := range g.childrenBlockData {
		if bd.GridX+bd.ExtraX > maxX {
			maxX = bd.GridX + bd.ExtraX
		}
		if bd.GridY+bd.ExtraY > maxY {
			maxY = bd.GridY + bd.ExtraY
		}
	}

	widths := make([]float64, maxX+1)
	heights := make([]float64, maxY+1)

	for _, bd := range g.childrenBlockData {
		sh, ok := g.childrenSizeHints[bd.Block]
		if !ok {
			continue
		}
		widths[bd.GridX] = math.Max(widths[bd.GridX], sh.PreferredSize.X)
		heights[bd.GridY] = math.Max(heights[bd.GridY], sh.PreferredSize.Y)
	}

	minXs := make([]float64, maxX)
	for i, w := range widths[:len(widths)-1] {
		if i != 0 {
			minXs[i] += minXs[i-1]
		}
		minXs[i] += w
	}
	minYs := make([]float64, maxY)
	for i, h := range heights[:len(heights)-1] {
		if i != 0 {
			minYs[i] += minYs[i-1]
		}
		minYs[i] += h
	}

	for _, bd := range g.childrenBlockData {
		minX := 0.0
		if bd.GridX != 0 {
			minX = minXs[bd.GridX-1]
		}
		minY := 0.0
		if bd.GridY != 0 {
			minY = minYs[bd.GridY-1]
		}
		bounds := geom.Rect{
			geom.Coord{minX, minY},
			geom.Coord{minX + widths[bd.GridX], minY + heights[bd.GridY]},
		}

		if bounds.Max.X > sizeHint.PreferredSize.X {
			sizeHint.PreferredSize.X = bounds.Max.X
		}
		if bounds.Max.Y > sizeHint.PreferredSize.Y {
			sizeHint.PreferredSize.Y = bounds.Max.Y
		}
	}

	sizeHint.MinSize = sizeHint.PreferredSize
	sizeHint.MaxSize = sizeHint.PreferredSize
	g.SetSizeHint(sizeHint)
}

func (g *Grid) regrid() {

	maxX, maxY := 0, 0
	for _, bd := range g.childrenBlockData {
		if bd.GridX+bd.ExtraX > maxX {
			maxX = bd.GridX + bd.ExtraX
		}
		if bd.GridY+bd.ExtraY > maxY {
			maxY = bd.GridY + bd.ExtraY
		}
	}

	widths := make([]float64, maxX+1)
	heights := make([]float64, maxY+1)

	for _, bd := range g.childrenBlockData {
		sh, ok := g.childrenSizeHints[bd.Block]
		if !ok {
			continue
		}
		widths[bd.GridX] = math.Max(widths[bd.GridX], sh.PreferredSize.X)
		heights[bd.GridY] = math.Max(heights[bd.GridY], sh.PreferredSize.Y)
	}

	minXs := make([]float64, maxX)
	for i, w := range widths[:len(widths)-1] {
		if i != 0 {
			minXs[i] += minXs[i-1]
		}
		minXs[i] += w
	}
	minYs := make([]float64, maxY)
	for i, h := range heights[:len(heights)-1] {
		if i != 0 {
			minYs[i] += minYs[i-1]
		}
		minYs[i] += h
	}

	for _, bd := range g.childrenBlockData {
		minX := 0.0
		if bd.GridX != 0 {
			minX = minXs[bd.GridX-1]
		}
		minY := 0.0
		if bd.GridY != 0 {
			minY = minYs[bd.GridY-1]
		}
		bounds := geom.Rect{
			geom.Coord{minX, minY},
			geom.Coord{minX + widths[bd.GridX], minY + heights[bd.GridY]},
		}

		g.ChildrenBounds[bd.Block] = bounds
		bd.Block.EventsIn <- uik.ResizeEvent{
			Size: geom.Coord{widths[bd.GridX], heights[bd.GridY]},
		}
	}
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
}

func (g *Grid) handleEvents() {
	for {
		select {
		case e := <-g.Events:
			switch e := e.(type) {
			case uik.ResizeEvent:
				g.Size = e.Size
				g.regrid()
				g.PaintAndComposite()
			default:
				g.Foundation.HandleEvent(e)
			}
		case bsh := <-g.BlockSizeHints:
			if !g.children[bsh.Block] {
				// Do I know you?
				break
			}
			g.childrenSizeHints[bsh.Block] = bsh.SizeHint

			g.makePreferences()
		case e := <-g.Redraw:
			g.DoRedraw(e)
		case e := <-g.CompositeBlockRequests:
			g.DoCompositeBlockRequest(e)
			// go uik.ShowBuffer("grid", g.Buffer)
		case g.config = <-g.setConfig:
			g.makePreferences()
		case g.getConfig <- g.config:
		case bd := <-g.add:
			g.addBlock(bd)
		case b := <-g.remove:
			g.remBlock(b)
		}
	}
}
