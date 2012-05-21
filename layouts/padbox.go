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

	add        chan *uik.Block
	remove     chan *uik.Block
	configs    chan interface{}
	invalidate chan bool
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
func (p *PadLayout) SetConfigUnsafe(cfg interface{}) {
	config, ok := cfg.(PadConfig)
	if ok {
		p.config = config
		p.invalidate <- true
	}
}
func (p *PadLayout) SetAddChan(add chan *uik.Block) {
	p.add = add
	add <- p.block
}
func (p *PadLayout) SetRemoveChan(remove chan *uik.Block) {
	p.remove = remove
}
func (p *PadLayout) SetConfigChan(configs chan interface{}) {
	p.configs = configs
}
func (p *PadLayout) SetInvalidateChan(invalidate chan bool) {
	p.invalidate = invalidate
}

func (p *PadLayout) SetBlock(block *uik.Block) {
	p.remove <- p.block
	p.block = block
	p.add <- p.block
}
