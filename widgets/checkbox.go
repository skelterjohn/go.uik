package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image/color"
)

type Checkbox struct {
	uik.Foundation

	state, pressed bool
}

func NewCheckbox(size geom.Coord) (c *Checkbox) {
	c = new(Checkbox)
	c.Initialize()
	c.Size = size

	go c.handleEvents()

	c.Paint = func(gc draw2d.GraphicContext) {
		c.draw(gc)
	}

	c.SetSizeHint(uik.SizeHint{
		MinSize: size,
		PreferredSize: size,
		MaxSize: size,
	})

	return
}

func (c *Checkbox) draw(gc draw2d.GraphicContext) {
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	if c.pressed {
		gc.SetFillColor(color.RGBA{155, 0, 0, 255})
	} else {
		gc.SetFillColor(color.RGBA{255, 0, 0, 255})
	}

	// Draw background rect
	x, y := gc.LastPoint()
	gc.MoveTo(0, 0)
	gc.LineTo(c.Size.X, 0)
	gc.LineTo(c.Size.X, c.Size.Y)
	gc.LineTo(0, c.Size.Y)
	gc.Close()
	gc.FillStroke()

	// Draw inner rect
	if c.state {
		gc.SetFillColor(color.Black)
		gc.MoveTo(5, 5)
		gc.LineTo(c.Size.X-5, 5)
		gc.LineTo(c.Size.X-5, c.Size.Y-5)
		gc.LineTo(5, c.Size.Y-5)
		gc.Close()
		gc.FillStroke()
	}

	gc.MoveTo(x, y)
}

func (c *Checkbox) handleEvents() {
	for {
		select {
		case e := <-c.Events:
			switch e.(type) {
			case uik.MouseDownEvent:
				c.pressed = true
				c.DoRedraw(uik.RedrawEvent{c.Bounds()})
			case uik.MouseUpEvent:
				c.pressed = false
				c.state = !c.state
				c.DoRedraw(uik.RedrawEvent{c.Bounds()})
			}
		case e := <-c.Redraw:
			c.DoRedraw(e)
		case cbr := <-c.CompositeBlockRequests:
			c.DoPaint(c.PrepareBuffer())
			c.DoCompositeBlockRequest(cbr)
		}
	}
}
