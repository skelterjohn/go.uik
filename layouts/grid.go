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
	childrenBlockData map[*uik.Block]BlockData
	config            GridConfig

	table        table
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

	g.Paint = func(gc draw2d.GraphicContext) {
		g.draw(gc)
	}
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
		g.hflex.add(&elem{
			index:    bd.GridX,
			extra:    bd.ExtraX,
			minSize:  csh.MinSize.X,
			prefSize: csh.PreferredSize.X,
			maxSize:  math.Inf(1),
		})
		g.vflex.add(&elem{
			index:    bd.GridY,
			extra:    bd.ExtraY,
			minSize:  csh.MinSize.Y,
			prefSize: csh.PreferredSize.Y,
			maxSize:  math.Inf(1),
		})
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
			gridBounds.Max.X = gridBounds.Min.X + csh.MaxSize.X
		}
		if gridSizeY > csh.MaxSize.Y {
			gridBounds.Max.Y = gridBounds.Min.Y + csh.MaxSize.Y
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

			g.regrid()
		case e := <-g.BlockInvalidations:
			g.DoBlockInvalidation(e)
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
