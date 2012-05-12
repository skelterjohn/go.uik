/*
   Copyright 2012 the go.wde authors

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

package uik

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

type InvalidationChan chan Invalidation

func (ch InvalidationChan) Stack(e Invalidation) {
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
