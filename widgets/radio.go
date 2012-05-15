package widgets

import (
	"github.com/skelterjohn/go.uik"
)

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
}

func NewRadio(options []string) (r *Radio) {
	r = &Radio{
		options:   options,
		selection: -1,
	}
	r.Initialize()

	go r.HandleEvents()

	return
}

func (r *Radio) Initialize() {
	r.Foundation.Initialize()

	r.setOptions = make(chan []string, 1)
	r.SetOptions = r.setOptions

	r.setSelection = make(chan int, 1)
	r.SetSelection = r.setSelection
}

func (r *Radio) HandleEvents() {
	for {
		select {
		case e := <-r.UserEvents:
			r.HandleEvent(e)
		}
	}
}

func (r *Radio) HandleEvent(e interface{}) {
	switch e := e.(type) {
	default:
		r.Foundation.HandleEvent(e)
	}
}
