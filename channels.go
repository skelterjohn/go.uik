package uik

import (
	"math"
)

// this type stolen from github.com/kylelemons/iq
type Ring struct {
	cap int
	cnt, i int
	data []interface{}
}

func (rb *Ring) Empty() bool {
	return rb.cnt == 0
}

func (rb *Ring) Peek() interface{} {
	return rb.data[rb.i]
}

func (rb *Ring) Enqueue(x interface{}) {
	if rb.cap != 0 && rb.cnt >= rb.cap {
		return
	}
	if rb.cnt >= len(rb.data) {
		rb.grow(2 * rb.cnt + 1)
	}
	rb.data[(rb.i + rb.cnt) % len(rb.data)] = x
	rb.cnt++
}

func (rb *Ring) Dequeue() {
	rb.cnt, rb.i = rb.cnt - 1, (rb.i + 1) % len(rb.data)
}

func (rb *Ring) grow(newSize int) {
	newData := make([]interface{}, newSize)

	n := copy(newData, rb.data[rb.i:])
	copy(newData[n:], rb.data[:rb.cnt-n])

	rb.i = 0
	rb.data = newData
}

// this function stolen from github.com/kylelemons/iq
func RingIQ(in <-chan interface{}, next chan<- interface{}, cap int) {
	var rb Ring
	rb.cap = cap
	defer func() {
		for !rb.Empty() {
			next <- rb.Peek()
			rb.Dequeue()
		}
		close(next)
	}()

	for {
		if rb.Empty() {
			v, ok := <-in
			if !ok {
				return
			}
			rb.Enqueue(v)
		}

		select {
		case next <- rb.Peek():
			rb.Dequeue()
		case v, ok := <-in:
			if !ok {
				return
			}
			rb.Enqueue(v)
		}
	}
}

func QueuePipe() (in chan<- interface{}, out <-chan interface{}) {
	inch := make(chan interface{})
	in = inch
	outch := make(chan interface{})
	out = outch
	go RingIQ(inch, outch, 0)
	return
}

func StackRedrawEvents(Redraw chan RedrawEvent) (out <-chan RedrawEvent) {
	outch := make(chan RedrawEvent)
	out = outch
	go func(Redraw, outch chan RedrawEvent) {
		var e RedrawEvent
		valid := false

		loop:
		for {
			if !valid {
				e, valid = <-Redraw
				if !valid {
					break
				}
			}
			select {
			case ne, ok := <-Redraw:
				if !ok {
					break loop
				}
				e.Bounds.Min.X = math.Min(e.Bounds.Min.X, ne.Bounds.Min.X)
				e.Bounds.Min.Y = math.Min(e.Bounds.Min.Y, ne.Bounds.Min.Y)
				e.Bounds.Max.X = math.Max(e.Bounds.Max.X, ne.Bounds.Max.X)
				e.Bounds.Max.Y = math.Max(e.Bounds.Max.Y, ne.Bounds.Max.Y)
			case outch<- e:
				valid = false
			}
		}
		close(outch)
	}(Redraw, outch)

	return
}

