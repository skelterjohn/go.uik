package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.wde"
	"image/color"
	"image/draw"
	"math"
)

type Clicker chan wde.Button

type Button struct {
	uik.Foundation
	Label   *Label
	pressed bool

	Clickers      map[Clicker]bool
	AddClicker    chan Clicker
	RemoveClicker chan Clicker
}

func NewButton(size geom.Coord, label string) (b *Button) {
	b = new(Button)
	b.Size = size
	b.Initialize()

	// uik.Report(b.ID, "is button")

	b.Label.SetConfig <- LabelData{
		Text:     label,
		FontSize: 12,
		Color:    color.Black,
	}

	go b.handleEvents()

	return
}

func (b *Button) Initialize() {
	b.Foundation.Initialize()

	b.DrawOp = draw.Over

	b.Label = NewLabel(b.Size, LabelData{
		Text:     "",
		FontSize: 12,
		Color:    color.Black,
	})
	lbounds := b.Bounds()
	lbounds.Min.X += 1
	lbounds.Min.Y += 1
	lbounds.Max.X -= 1
	lbounds.Max.Y -= 1
	b.PlaceBlock(&b.Label.Block, lbounds)

	b.Clickers = map[Clicker]bool{}
	b.AddClicker = make(chan Clicker, 1)
	b.RemoveClicker = make(chan Clicker, 1)

	b.Paint = func(gc draw2d.GraphicContext) {
		b.draw(gc)
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

func (b *Button) draw(gc draw2d.GraphicContext) {
	gc.Clear()

	gc.SetStrokeColor(color.Black)
	if b.pressed {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
		safeRect(gc, geom.Coord{0, 0}, b.Size)
		gc.FillStroke()
	} else {
		gc.SetFillColor(color.RGBA{200, 200, 200, 255})
		safeRect(gc, geom.Coord{0, 0}, b.Size)
		gc.FillStroke()
	}
}

func (b *Button) handleEvents() {

	ld := <-b.Label.GetConfig
	ld.Text = "pressing!"

	for {
		select {
		case e := <-b.Events:
			switch e := e.(type) {
			case uik.MouseDownEvent:
				b.pressed = true
				b.Rebuffer()
			case uik.MouseUpEvent:
				b.pressed = false
				for c := range b.Clickers {
					select {
					case c <- e.Which:
					default:
					}
				}
				b.Rebuffer()
				// go uik.ShowBuffer("button buffer", b.Buffer)
			case uik.ResizeEvent:
				b.Foundation.HandleEvent(e)
				lbounds := b.Bounds()
				lbounds.Min.X += 1
				lbounds.Min.Y += 1
				lbounds.Max.X -= 1
				lbounds.Max.Y -= 1
				b.ChildrenBounds[&b.Label.Block] = lbounds
				b.Label.EventsIn <- uik.ResizeEvent{
					Size: b.Size,
				}
			default:
				b.Foundation.HandleEvent(e)
			}
		case e := <-b.Redraw:
			b.DoRedraw(e)
		case cbr := <-b.CompositeBlockRequests:
			b.DoCompositeBlockRequest(cbr)
		case c := <-b.AddClicker:
			b.Clickers[c] = true
		case c := <-b.RemoveClicker:
			if b.Clickers[c] {
				delete(b.Clickers, c)
			}
		case bsh := <-b.BlockSizeHints:
			sh := bsh.SizeHint
			sh.PreferredSize.X += 10
			sh.PreferredSize.Y += 10
			sh.MaxSize.X = math.Inf(1)
			sh.MaxSize.Y = math.Inf(1)
			b.SetSizeHint(sh)

		}
	}
}
