/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package layouts

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type PadConfig struct {
	MinX, MinY, MaxX, MaxY float64
}

type PadLayout struct {
	layouter *Layouter

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

func (p *PadLayout) SetLayouter(layouter *Layouter) {
	p.layouter = layouter
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
func (p *PadLayout) ConfigUnsafe(cfg interface{}) {
	switch x := cfg.(type) {
	case PadConfig:
		p.config = x
		p.layouter.Invalidate()
	case *uik.Block:
		if x == p.block {
			break
		}
		p.layouter.RemoveBlock(p.block)
		p.block = x
		p.layouter.AddBlock(p.block)
	}
}

func (p *PadLayout) SetBlock(block *uik.Block) {
	p.layouter.Config(block)
}
func (p *PadLayout) SetConfig(cfg PadConfig) {
	p.layouter.Config(cfg)
}
