package uik

import (
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/go.wde/xgb"
	"code.google.com/p/draw2d/draw2d"
	"image/color"
	"fmt"
)

var WindowGenerator func (parent wde.Window, width, height int) (window wde.Window, err error)

func init() {
	WindowGenerator = func(parent wde.Window, width, height int) (window wde.Window, err error)  {
		window, err = xgb.NewWindow(width, height)
		return
	}
}

type EventFilter func(event interface{}) bool

type Size struct {
	W, H float64
}

type Point struct {
	X, Y float64
}

type Block interface {
	// constrain this block to a certain size
	SetSize(sz Size) (err error)
	GetSize() (sz Size)

	// draw this block in a context constrained by size and min at the origin
	Draw(gc draw2d.GraphicContext) (err error)

	// all events that occur in this block will be filtered and sent on ch
	Subscribe(ch chan<- interface{}, filter EventFilter)
}

type Foundation struct {
	Window wde.Window
	GC draw2d.GraphicContext
	Main Block
	events <-chan interface{}
}

func NewFoundation(sz Size) (f *Foundation, err error) {
	f = &Foundation{}
	f.Window, err = WindowGenerator(nil, int(sz.W), int(sz.H))
	if err != nil {
		return
	}
	f.GC = draw2d.NewGraphicContext(f.Window.Screen())
	f.events = f.Window.EventChan()
	go f.handleEvents()
	return
}

func (f *Foundation) handleEvents() {
	for e := range f.events {
		if _, ok := e.(wde.CloseEvent); ok {
			f.Window.Close()
		}
		fmt.Println(e)
	}
}

func (f *Foundation) SetSize(sz Size) (err error) {
	f.Window.SetSize(int(sz.W), int(sz.H))
	if f.Main != nil {
		f.Main.SetSize(sz)
	}
	return
}

func (f *Foundation) Draw() (err error) {
	f.GC.Clear()
	err = f.Main.Draw(f.GC)
	f.Window.FlushImage()
	return
}

func (f *Foundation) GetSize() (sz Size) {
	w, h := f.Window.Size()
	sz.W = float64(w)
	sz.H = float64(h)
	return
}

type Terminal struct {
	Size Size
}
var _ Block = &Terminal{}

func (t *Terminal) SetSize(sz Size) (err error) {
	t.Size = sz
	return
}

func (t *Terminal) GetSize() (sz Size) {
	sz = t.Size
	return
}

func (t *Terminal) Draw(gc draw2d.GraphicContext) (err error) {
	return
}

func (t *Terminal) Subscribe(ch chan<- interface{}, filter EventFilter) {

}

type Button struct {
	Terminal
	Label string
}

func NewButton(label string) (button *Button) {
	button = &Button{
		Label:label,
	}
	button.Size = Size{100, 30}
	return
}

func (b *Button) Draw(gc draw2d.GraphicContext) (err error) {
	gc.BeginPath()
	gc.SetFillColor(color.White)
	gc.SetStrokeColor(color.Black)
	draw2d.Rect(gc, 0, 0, b.Size.W, b.Size.H)
	gc.Stroke()
	gc.Close()
	return
}