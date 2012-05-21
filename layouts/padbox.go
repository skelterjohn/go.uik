package layouts

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type PadConfig struct {
	MinX, MinY, MaxX, MaxY float64
}

type PadLayout struct {
	block *uik.Block

	config PadConfig
	hint   uik.SizeHint
}

func NewPadBox(config PadConfig, block *uik.Block) (pb *Layouter) {
	return NewLayouter(NewPadLayout(config, block))
}

func NewPadLayout(config PadConfig, block *uik.Block) (p *PadLayout) {
	p = new(PadLayout)

	p.config = config
	p.block = block
	return
}
func (p *PadLayout) SetHint(block *uik.Block, hint uik.SizeHint) {
	if block == p.block {
		p.hint = hint
	}
}
func (p *PadLayout) GetHint() (hint uik.SizeHint) {
	hint.MinSize.X = p.hint.MinSize.X + p.config.MinX + p.config.MaxX
	hint.MinSize.Y = p.hint.MinSize.Y + p.config.MinY + p.config.MaxY
	hint.PreferredSize.X = p.hint.PreferredSize.X + p.config.MinX + p.config.MaxX
	hint.PreferredSize.Y = p.hint.PreferredSize.Y + p.config.MinY + p.config.MaxY
	hint.MaxSize.X = p.hint.MaxSize.X + p.config.MinX + p.config.MaxX
	hint.MaxSize.Y = p.hint.MaxSize.Y + p.config.MinY + p.config.MaxY
	return
}
func (p *PadLayout) GetLayout(size geom.Coord) (l Layout) {
	l = make(Layout)
	l[p.block] = geom.Rect{
		Min: geom.Coord{p.config.MinX, p.config.MinY},
		Max: geom.Coord{size.X - p.config.MaxX, size.Y - p.config.MaxY},
	}
	return
}
func (p *PadLayout) SetAdd(add func(*uik.Block)) {
	add(p.block)
}
