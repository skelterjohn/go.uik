package uik

// this type stolen from github.com/kylelemons/iq
type Ring struct {
	cap    int
	cnt, i int
	data   []interface{}
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
		rb.grow(2*rb.cnt + 1)
	}
	rb.data[(rb.i+rb.cnt)%len(rb.data)] = x
	rb.cnt++
}

func (rb *Ring) Dequeue() {
	rb.cnt, rb.i = rb.cnt-1, (rb.i+1)%len(rb.data)
}

func (rb *Ring) grow(newSize int) {
	newData := make([]interface{}, newSize)

	n := copy(newData, rb.data[rb.i:])
	copy(newData[n:], rb.data[:rb.cnt-n])

	rb.i = 0
	rb.data = newData
}

// this function stolen from github.com/kylelemons/iq
func RingI2Q(in <-chan interface{}, next chan<- interface{}, cap int) {
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

func QueuePipe2() (in chan<- interface{}, out <-chan interface{}) {
	inch := make(chan interface{})
	in = inch
	outch := make(chan interface{})
	out = outch
	go RingI2Q(inch, outch, 0)
	return
}

type CompositeRequestChan chan CompositeRequest

func (ch CompositeRequestChan) Stack(cr CompositeRequest) {
	if ch == nil {
		return
	}
	for {
		select {
		case ch <- cr:
			return
		case <-ch:
		}
	}
}

type SizeHintChan chan SizeHint

func (ch SizeHintChan) Stack(sh SizeHint) {
	if ch == nil {
		return
	}
	for {
		select {
		case ch <- sh:
			return
		case <-ch:
		}
	}
}

type RedrawEventChan chan RedrawEvent

func (ch RedrawEventChan) Stack(e RedrawEvent) {
	if ch == nil {
		return
	}
	for {
		select {
		case ch <- e:
			return
		case ne := <-ch:
			e.Bounds.ExpandToContainRect(ne.Bounds)
		}
	}
}

type placementNotificationChan chan placementNotification

func (ch placementNotificationChan) Stack(e placementNotification) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		case ch <- e:
			return
		}
	}
}

type Filter func(e interface{}) (accept, done bool)

type Subscription struct {
	Filter Filter
	Ch     chan<- interface{}
}

type DropChan chan<- interface{}

func (ch DropChan) SendOrDrop(e interface{}) {
	select {
	case ch <- e:
	default:
	}
}

func SubscriptionQueue(cap int) (in chan<- interface{}, out <-chan interface{}, sub chan<- Subscription) {
	inch := make(chan interface{}, cap)
	mch := inch
	// mch := make(chan interface{})
	// go RingIQ(inch, mch, cap)

	subch := make(chan Subscription, 1)
	outch := make(chan interface{}, 1)

	go func(mch <-chan interface{}, outch chan<- interface{}, subch <-chan Subscription) {
		subscriptions := map[*Filter]chan<- interface{}{}
		for {
			select {
			case e := <-mch:
				ok := true
				if !ok {
					close(outch)
					return
				}
				outch <- e
				for foo, ch := range subscriptions {
					accept, done := (*foo)(e)
					if accept {
						select {
						case ch <- e:
						default:
						}
					}
					if done {
						delete(subscriptions, foo)
					}
				}
			case sub := <-subch:
				subscriptions[&sub.Filter] = sub.Ch
			}

		}
	}(mch, outch, subch)

	in = inch
	out = outch
	sub = subch

	return
}

type KeyFocusChan chan *Block

func (ch KeyFocusChan) Stack(b *Block) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		case ch <- b:
			return
		}
	}
}
