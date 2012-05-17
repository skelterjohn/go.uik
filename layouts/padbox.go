package layouts

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type PadConfig struct {
	MinX, MinY, MaxX, MaxY float64
}

type PadBox struct {
	uik.Foundation

	block  *uik.Block
	config PadConfig
}

func NewPadBox(config PadConfig, block *uik.Block) (p *PadBox) {
	p = new(PadBox)
	p.config = config
	p.block = block

	p.Initialize()

	p.AddBlock(block)

	go p.HandleEvents()

	return
}

func (p *PadBox) Initialize() {
	p.Foundation.Initialize()

	p.Paint = func(gc draw2d.GraphicContext) {

	}
}

func (p *PadBox) HandleEvents() {
	for {
		select {
		case e := <-p.UserEvents:
			switch e := e.(type) {
			default:
				p.Foundation.HandleEvent(e)
			}
		case e := <-p.BlockInvalidations:
			p.DoBlockInvalidation(e)

		case bsh := <-p.BlockSizeHints:
			if !p.Children[bsh.Block] {
				// Do I know you?
				break
			}

			sh := bsh.SizeHint
			sh.MinSize.X += p.config.MinX + p.config.MaxX
			sh.PreferredSize.X += p.config.MinX + p.config.MaxX
			sh.MaxSize.X += p.config.MinX + p.config.MaxX

			p.SetSizeHint(sh)

		case e := <-p.ResizeEvents:

			p.DoResizeEvent(e)

			cbounds := geom.Rect{Max: p.Size}
			cbounds.Min.X += p.config.MinX
			cbounds.Min.Y += p.config.MinY
			cbounds.Max.X -= p.config.MaxX
			cbounds.Max.Y -= p.config.MaxY

			p.PlaceBlock(p.block, cbounds)
		}
	}
}
