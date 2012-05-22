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

package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.uik/layouts"
	"github.com/skelterjohn/go.wde"
	"image/color"
)

type RadioSelection struct {
	Index  int
	Option string
}

type SelectionListener chan RadioSelection

type Radio struct {
	uik.Foundation

	options    []string
	setOptions chan []string
	SetOptions chan<- []string
	getOptions chan []string
	GetOptions <-chan []string

	selection    int
	setSelection chan int
	SetSelection chan<- int
	getSelection chan int
	GetSelection <-chan int

	buttons     []*Button
	buttonsDone []chan bool

	selectionListeners      map[SelectionListener]bool
	addSelectionListener    chan SelectionListener
	AddSelectionListener    chan<- SelectionListener
	removeSelectionListener chan SelectionListener
	RemoveSelectionListener <-chan SelectionListener

	radioGrid *layouts.Grid
}

func NewRadio(options []string) (r *Radio) {
	r = new(Radio)
	r.Initialize()

	if uik.ReportIDs {
		uik.Report(r.ID, "radio")
	}

	go r.HandleEvents()

	r.SetOptions <- options

	return
}

func (r *Radio) Initialize() {
	r.Foundation.Initialize()

	r.setOptions = make(chan []string, 1)
	r.SetOptions = r.setOptions

	r.setSelection = make(chan int, 1)
	r.SetSelection = r.setSelection

	r.selectionListeners = map[SelectionListener]bool{}
	r.addSelectionListener = make(chan SelectionListener, 1)
	r.AddSelectionListener = r.addSelectionListener
	r.removeSelectionListener = make(chan SelectionListener, 1)
	r.RemoveSelectionListener = r.removeSelectionListener

	r.selection = -1

	r.radioGrid = layouts.NewGrid(layouts.GridConfig{})
	r.AddBlock(&r.radioGrid.Block)
	r.radioGrid.Paint = nil

	r.Paint = func(gc draw2d.GraphicContext) {
		gc.SetFillColor(color.Black)
		bbounds := r.Bounds()
		safeRect(gc, bbounds.Min, bbounds.Max)
		gc.Fill()
	}
}

func (r *Radio) HandleEvents() {
	for {
		select {
		case e := <-r.UserEvents:
			r.HandleEvent(e)
		case e := <-r.ResizeEvents:
			r.Foundation.DoResizeEvent(e)
			r.PlaceBlock(&r.radioGrid.Block, geom.Rect{Max: r.Size})
		case options := <-r.setOptions:
			r.makeButtons(options)
		case r.selection = <-r.setSelection:
			r.updateButtons()
			for selLis := range r.selectionListeners {
				selLis <- RadioSelection{
					Index:  r.selection,
					Option: r.options[r.selection],
				}
			}
		case r.getSelection <- r.selection:
		case bsh := <-r.BlockSizeHints:
			r.ChildrenHints[bsh.Block] = bsh.SizeHint
			if bsh.Block != &r.radioGrid.Block {
				// who is this?
				break
			}
			// uik.Report("gotpr", r.Block.ID, bsh.Block.ID, bsh.SizeHint)
			// uik.Report("prefs", r.Block.ID, bsh.SizeHint)
			r.SetSizeHint(bsh.SizeHint)
		case inv := <-r.BlockInvalidations:
			r.Invalidate(inv.Bounds...)
		case selLis := <-r.addSelectionListener:
			r.selectionListeners[selLis] = true
		case selLis := <-r.removeSelectionListener:
			if r.selectionListeners[selLis] {
				delete(r.selectionListeners, selLis)
			}
		}
	}
}

func (r *Radio) HandleEvent(e interface{}) {
	switch e := e.(type) {
	default:
		r.Foundation.HandleEvent(e)
	}
}

func (r *Radio) makeButtons(options []string) {
	// see if the options are actually different
	changed := len(r.options) != len(options)
	if !changed {
		for i := range r.options {
			if r.options[i] != options[i] {
				changed = true
			}
		}
	}
	if !changed {
		return
	}
	r.options = options

	// remove old buttons
	for _, b := range r.buttons {
		r.RemoveBlock(&b.Block)
	}
	for _, d := range r.buttonsDone {
		d <- true
	}

	r.buttons = make([]*Button, len(r.options))
	r.buttonsDone = make([]chan bool, len(r.options))
	for i, option := range r.options {
		ob := NewButton(option)
		r.buttons[i] = ob
		r.buttonsDone[i] = make(chan bool, 1)

		pb := layouts.NewPadBox(layouts.PadConfig{
			MinX: 2, MinY: 2,
			MaxX: 2, MaxY: 2,
		}, &ob.Block)

		r.radioGrid.Add(&pb.Block, layouts.BlockData{
			Block: &pb.Block,
			GridX: 0, GridY: i,
			AnchorX: layouts.AnchorMin | layouts.AnchorMax,
			AnchorY: layouts.AnchorMin | layouts.AnchorMax,
		})

		clicker := make(chan wde.Button, 1)
		go func(clicker chan wde.Button, index int, done chan bool) {
			for {
				select {
				case <-clicker:
					r.SetSelection <- index
				case <-done:
					return
				}
			}
		}(clicker, i, r.buttonsDone[i])
		ob.AddClicker <- clicker
	}
	r.updateButtons()
}

func (r *Radio) updateButtons() {
	for i, b := range r.buttons {
		if i == r.selection {
			b.SetConfig(ButtonConfig{
				Color: color.RGBA{110, 110, 110, 255},
			})
		} else {
			b.SetConfig(ButtonConfig{})
		}
	}
}
