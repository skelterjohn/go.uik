package layouts

import (
	"github.com/skelterjohn/go.uik"
)

type PadConfig struct {
	MinX, MinY, MaxX, MaxY float64
}

type PadBox struct {
	uik.Foundation

	block  *uik.Block
	config PadConfig

	setConfig, getConfig chan PadConfig
}

func NewPadBox(config PadConfig, block *uik.Block) (p *PadBox) {
	p = new(PadBox)
	p.config = config
	p.block = block

	p.Initialize()

	if uik.ReportIDs {
		uik.Report(p.ID, "padbox")
	}

	p.AddBlock(block)

	go p.HandleEvents()

	return
}

func (p *PadBox) Initialize() {
	p.Foundation.Initialize()

	p.setConfig = make(chan PadConfig, 1)
	p.getConfig = make(chan PadConfig, 1)

	p.Paint = nil
}

func (p *PadBox) SetConfig(cfg PadConfig) {
	p.setConfig <- cfg
}

func (p *PadBox) GetConfig() (cfg PadConfig) {
	cfg = <-p.getConfig
	return
}

func (p *PadBox) repad() {
	cbounds := p.Bounds()
	cbounds.Min.X += p.config.MinX
	cbounds.Min.Y += p.config.MinY
	cbounds.Max.X -= p.config.MaxX
	cbounds.Max.Y -= p.config.MaxY
	p.PlaceBlock(p.block, cbounds)
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
			p.repad()

		case p.getConfig <- p.config:

		case p.config = <-p.setConfig:
			p.repad()
		}
	}
}
