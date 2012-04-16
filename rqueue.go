package uik

type RB struct {
	length, start int
	data []interface{}
}

func (rb *RB) Enqueue(x interface{}) {
	if rb.length >= len(rb.data) {
		rb.sizeUp(2*rb.length)
	}
	index := rb.start + rb.length
	index = index % len(rb.data)
	rb.data[index] = x
	rb.length++
}

func (rb *RB) Dequeue() (x interface{}) {
	x = rb.data[rb.start]
	rb.start = (rb.start + 1) % len(rb.data)
	rb.length--
	
	if rb.start == 0 {
		rb.data = rb.data[:rb.length]
	}
	
	return
}

func (rb *RB) sizeUp(newSize int) {
	newData := make([]interface{}, newSize)
	//copy the stuff at the end
	endLength := len(rb.data) - rb.start
	if endLength > rb.length {
		endLength = rb.length
	}
	copy(newData[:endLength], rb.data[rb.start:rb.start+endLength])
	
	// did it wrap?
	if endLength < rb.length {
		//copy the stuff at the beginning
		endIndex := (rb.start + rb.length) % len(rb.data)
		copy(newData[rb.start+endLength:rb.start+rb.length], rb.data[:endIndex])
	}
	rb.start = 0
	rb.data = newData
}

func QueuePipe() (in chan<- interface{}, out <-chan interface{}) {
	inch := make(chan interface{})
	in = inch
	outch := make(chan interface{})
	out = outch
	go func(inch <-chan interface{}, outch chan<- interface{}) {
		var rb RB
		var cur interface{}
		for {
			if rb.length == 0 {
				var ok bool
				cur, ok = <-inch
				if !ok {
					close(outch)
					return
				}
			}
			select {
			case outch<- cur:
				if rb.length != 0 {
					cur = rb.Dequeue()
				}
			case v, ok := <-inch:
				if !ok {
					outch<- cur
					for rb.length != 0 {
						outch<- rb.Dequeue()
					}
					close(outch)
					return
				}
				rb.Enqueue(v)
			}
		}
	}(inch, outch)
	return
}