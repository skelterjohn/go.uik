package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"image/color"
	"github.com/skelterjohn/geom"
)

type Checkbox struct {
	Foundation

	state bool
}

func NewCheckbox(size geom.Coord) (c *Checkbox) {
	c = new(Checkbox)
	c.Initialize()
	c.Size = size

	go c.handleEvents()

	c.Paint = func(gc draw2d.GraphicContext) {
		c.draw(gc)
	}

	return
}

func (c *Checkbox) draw(gc draw2d.GraphicContext) {
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	gc.SetFillColor(color.RGBA{255,0,0,255})

	// Draw background rect
	x, y := gc.LastPoint()
	gc.MoveTo(0,0)
	gc.LineTo(c.Size.X, 0)
	gc.LineTo(c.Size.X, c.Size.Y)
	gc.LineTo(0, c.Size.Y)
	gc.Close()
	gc.FillStroke()

	// Draw inner rect
	if c.state {
		gc.SetFillColor(color.Black)
		gc.MoveTo(5,5)
		gc.LineTo(c.Size.X-5, 5)
		gc.LineTo(c.Size.X-5, c.Size.Y-5)
		gc.LineTo(5, c.Size.Y-5)
		gc.Close()
		gc.FillStroke()
	}

	gc.MoveTo(x,y)
}

func (c *Checkbox) handleEvents() {
	c.ListenedChannels[c.MouseDownEvents] = true

	for {
		select {
		case <-c.MouseDownEvents:
			if c.state {
				c.state = false
			} else {
				c.state = true
			}
			c.DoRedraw(RedrawEvent{c.Bounds()})
		case e := <-c.Redraw:
			c.DoRedraw(e)
		case cbr := <-c.CompositeBlockRequests:
			c.DoPaint(c.PrepareBuffer())
			c.DoCompositeBlockRequest(cbr)
		}
	}
}
