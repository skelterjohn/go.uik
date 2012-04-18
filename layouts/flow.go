package layouts

import (
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.wde"
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"image/color"
)

type Clicker chan wde.Button

type Button struct {
	uik.Foundation
	Label *Label
	pressed bool

	Click chan<- wde.Button
	click chan wde.Button

	Clickers map[Clicker]chan<- interface{}
	AddClicker chan Clicker
	RemoveClicker chan Clicker
}

func NewButton(size geom.Coord, label string) (b *Button) {
	b = new(Button)
	b.Initialize()
	b.Size = size

	b.Label = NewLabel(size, LabelData{
		Text: label,
		FontSize: 12,
		Color: color.Black,
	})
	lbounds := b.Bounds()
	lbounds.Min.X += 1
	lbounds.Min.Y += 1
	lbounds.Max.X -= 1
	lbounds.Max.Y -= 1
	b.PlaceBlock(&b.Label.Block, lbounds)


	b.click = make(chan wde.Button)
	b.Click = b.click

	b.Clickers = map[Clicker]chan<- interface{}{}
	b.AddClicker = make(chan Clicker)
	b.RemoveClicker = make(chan Clicker)

	go b.handleEvents()

	b.Paint = func(gc draw2d.GraphicContext) {
		b.draw(gc)
	}

	return
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
	} else {
		gc.SetFillColor(color.White)
	}
	safeRect(gc, geom.Coord{0, 0}, b.Size)
	gc.FillStroke()
}

func (b *Button) handleEvents() {
	b.ListenedChannels[b.MouseDownEvents] = true
	b.ListenedChannels[b.MouseUpEvents] = true

	ld := <-b.Label.GetConfig
	ld.Text = "pressing!"

	for {
		select {
		case <-b.MouseDownEvents:
			b.pressed = true
			b.Label.SetConfig<- ld
			b.DoRedraw(uik.RedrawEvent{b.Bounds()})
		case e := <-b.MouseUpEvents:
			b.pressed = false
			for c := range b.Clickers {
				c <- e.Which
			}
			b.DoRedraw(uik.RedrawEvent{b.Bounds()})
		case cbr := <-b.CompositeBlockRequests:
			b.DoPaint(b.PrepareBuffer())
			b.DoCompositeBlockRequest(cbr)
		case e := <-b.Redraw:
			b.DoRedraw(e)
		case c := <-b.AddClicker:
			clickHead := clickerPipe(c)
			b.Clickers[c] = clickHead
		case c := <-b.RemoveClicker:
			if _, ok := b.Clickers[c]; ok {
				delete(b.Clickers, c)
			}
		}
	}
}

func clickerPipe(c Clicker) (head chan interface{}) {
	head = make(chan interface{})
	tail := make(chan interface{})
	go uik.RingIQ(head, tail, 0)
	go func() {
		for click := range tail {
			c <- click.(wde.Button)
		}
	}()
	return
}
