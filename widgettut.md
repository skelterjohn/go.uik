#Creating a widget with go.uik

As some of you may know, I've been working lately on a pure go GUI toolkit called go.uik (go UI kit). The project's repository is available on github (http://github.com/skelterjohn/go.uik).

One of my core goals with go.uik is to ensure that it doesn't get in its own way. Specifically, components cannot block each other. If component 1 decides to go off and do a web search in response to an event, this won't slow down component 2 on the other side of the window.

The way this is accomplished with go.uik is the restriction that all inter-component communication be done via non-blocking channel communication, and that each component runs its own goroutine. In many languages, this would be difficult to do, complicated to understand, and inefficient to execute.

Fortunately for us, go makes it all very easy.

In this post, I will add a radio button/group to the (completely insufficient) widget toolkit available in the uik/widgets package. As I create it, I'll document what I need to do this this post in hopes that someone will read it and make other widgets for me, so I can finish writing my dissertation without stalling the world of go+GUI.

###The first step is vision

I don't like the classical circles-means-radio-buttons scheme. I'm going to make something a bit different, but still recognizable as a set of options from which you may choose only one.

Here's a sketch: https://www.dropbox.com/sh/89guvpsbtlcexf1/t4VN0RZ2oq/rb/vision.pdf

It will be a series of vertically stacked regular buttons, connected by a dark background, and the one that is currently selected will be shaded darker than the others.

###Working with the go.uik types

Every component in a go.uik app that gets to draw to a certain part of the screen and receive user input needs to associate itself with a ```uik.Block```. The correct way to do this is to simply embed the ```uik.Block``` in your widget type.

A ```uik.Block``` has all the channels that are needed to function in the program. You, as a widget designer, have to make sure that your widget is actually receiving from the correct channels. Fortunately, this is easy to do.

There is an important specialization of ```uik.Block```, and that is ```uik.Foundation```. The ```uik.Foundation``` type uses ```uik.Block``` just like any other widget should:

```go
type Foundation struct {
	Block
	/* ... */
}
```

As a result, if you embed a ```uik.Foundation```, you are also embedding a uik.Block.

The purpose of ```uik.Foundation``` is to support one or more ```uik.Block```s by positioning them correctly, forwarding user input, and compositing the visuals.

###Back to the radio group widget

First, I'll create the source file ```go.uik/widgets/radio.go``` and declare the ```Radio``` type. Since our radio group widget will be making use of several buttons, we'll embed ```uik.Foundation```.

```go
package widgets

import (
	"github.com/skelterjohn/go.uik"
)

type Radio struct {
	uik.Foundation
}
```

The next step is to decide how the interface will work. I may change my mind later, but for now I'm going to have the programmer pass the ```Radio``` a ```[]string``` containing the options to display and the order in which they're displayed. I'll also let the programmer specify which one is currently selected, and provide a mechanism to notify the programmer when a new option is clicked.

Since the ```Radio```, like all go.uik components, will be running its own goroutine, it's important to make sure data goes to and from it in a threadsafe manner. For go.uik, that means via a channel.


```go
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
```

This might look a bit heavy, but the channels that are spelled the same with different capitalizations are the same channel. So, each bit of configurable data needs a channel to set it and a channel to get it.

In general, a widget designer might want to collect all configuration data into a single type ```MyWidgetConfig```. In this case I think it's important to be able to set the current selection without also setting the available options, so I've split it into two pieces.

The next step is to create the widget's initialization functions.

```go
func NewRadio(options []string) (r *Radio) {
	r = new(Radio)

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

}
```

Two initialization functions, here. One for when someone wants just a ```Radio```, and one for any time someone wants to use the ```Radio``` type, including embedding it somewhere else. It's important that all initialization necessary go in the ```.Initialize()``` method, since it's the only function likely to be called by a widget that might want to embed the ```Radio```.

You'll also notice that the ```NewRadio()``` function, along with calling ```r.Initialize()```, starts a goroutine running the (currently empty) ```.HandleEvents()``` method.

The ```.HandleEvents()``` method is the widgets core logic goroutine. Every widget must have one, and the correct place to invoke it is in the ```NewXYZ()``` function. The reason the ```NewXYZ()``` function doesn't contain the code in the ```.Initialize()``` method is because you only want to create the goroutine if the widget is not embedded in something else. If it *is* embedded somewhere else, and only ```.Initialize()``` is called, the embedding widget will create its own goroutine. You definitely don't want two goroutines fighting over the same channels, here.

The next obvious bit to fill out here is the ```.HandleEvents()``` method. We will start it off like the following.

```go
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
```

The ```.HandleEvents()``` method just waits for input. Right now it listens only to the ```.UserEvents``` channel (which comes from uik.Block), though later we'll add a couple more. When it gets something from this channel, it does a type switch to see what it's looking at. We may later decide to do something with certain kinds of events. For now we'll let uik.Foundation's ```.HandleEvent()``` method deal with them. ```uik.Foundation```'s ```.HandleEvent()``` takes care of funneling events to the correct children.

The next thing to do is to create the buttons for each option. I'll add a slice to the ```Radio``` type and create a ```.makeButtons()``` method.

```go
type Radio struct {
	/* ... */

	buttons []*Button
}
```

```go
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

	r.buttons = make([]*Button, len(r.options))
	for i, option := range r.options {
		ob := NewButton(option)
		r.buttons[i] = ob
		r.AddBlock(&ob.Block)
	}
}
```

The next step is to call ```.makeButtons()``` from somewhere. We'll add a case in the ```.HandleEvents()``` ```select{}``` statement to listen to configuration setting.

```go
func (r *Radio) HandleEvents() {
	for {
		select {
		case e := <-r.UserEvents:
			r.HandleEvent(e)
		case options := <-r.setOptions:
			r.makeButtons(options)
		}
	}
}
```

As a result of dedicating a single goroutine to this widget, we know that we can safeuly manipulate the ```Radio```'s internals in ```.makeButtons()``` and forward, among other things, the mouse events from the ```.UserEvents``` channel to the right button without worrying about thread safety. It's threadsafe because only one goroutine touches this data.

Let's modify ```NewRadio()``` to set the options provided.

```go
func NewRadio(options []string) (r *Radio) {
	r = new(Radio)
	r.Initialize()

	go r.HandleEvents()

	r.SetOptions <- options

	return
}
```

###Placing the buttons

To place the buttons, we'll use a ```layouts.Grid```. The ```uik.Foundation``` embedded in the ```Radio``` will have one child which is the ```layouts.Grid```.

The ```layouts.Grid``` is a layout manager in the style of java swing's ```GridBagLayout```. To understand how it places components, consider a grid with an infinite number of rows and columns. You then place your child components in cells by specifying their X and Y positions, and also how many rows and columns they occupy (with the top left being at the listed X and Y). Then every row and column in the grid is expanded to fit the components that occupy it. Since we're doing this on a computer with finite memory and processing power, the infinite number of rows and columns that have no occupants, and take up no space, will not be represented in the data structure.

To nicely use this ```layouts.Grid```, we'll have to resize it when the ```Radio``` is resized and use the ```uik.SizeHint```s that come in from the ```layouts.Grid``` to create the ```uik.SizeHint```s that will be sent to the ```Radio```'s parent.

First, let's create the ```layouts.Grid``` and add it to the ```Radio```'s children.

```go
type Radio struct {
	/* ... */
	radioGrid *layouts.Grid
}
```

```go
func (r *Radio) Initialize() {
	/* ... */
	r.radioGrid = layouts.NewGrid(layouts.GridConfig{})
	r.AddBlock(&r.radioGrid.Block)
}

```

Notice that we aren't passing the ```layouts.Grid``` itself, but rather the ```uik.Block``` that it embeds. The ```uik.Block``` defines all the important communication channels that define its interface. Rather than using blocking method calls that come with using an ```interface``` type, go.uik uses non-blocking channel communication to keep things moving.

Where a classical UI toolkit might use inheritance and polymorphism to define how a component differs from its ancestor in how it responds to user input, in go.uik the channels are read by a different goroutine, which can call the embedded helper ```.HandleEvent()``` methods in default cases.

One event we'll want to handle differently is the ```uik.ResizeEvent```. This isn't exactly user input, so it might one day no longer come in on the ```.UserEvents``` channel, but for now that's where it goes. When the ```Radio``` is resized, we'll want to place ```.radioGrid``` in such a way that it takes up the entirety of the ```Radio```.

To do this, we'll add a case in ```.HandleEvent()```.

```go

func (r *Radio) HandleEvent(e interface{}) {
	switch e := e.(type) {
	case uik.ResizeEvent:
		r.Foundation.HandleEvent(e)
		r.PlaceBlock(&r.radioGrid.Block, geom.Rect{Max: e.Size})
	default:
		r.Foundation.HandleEvent(e)
	}
}
```

Notice that we're still forwarding the ```uik.ResizeEvent``` to the ```uik.Foundation```'s ```.HandleEvent()``` method. ```uik.Foundation``` might do something important with that event, I don't remember. Or one day it might start doing something with it, and I don't want ```Radio``` to break. Either way, we forward it.

That takes care of resizing, but we also need to monitor the ```uik.SizeHint``` reported by ```.radioGrid```.

```go
func (r *Radio) HandleEvents() {
	for {
		select {
		case e := <-r.UserEvents:
			r.HandleEvent(e)
		case options := <-r.setOptions:
			r.makeButtons(options)
		case bsh := <-r.BlockSizeHints:
			r.ChildrenHints[bsh.Block] = bsh.SizeHint
			if bsh.Block != &r.radioGrid.Block {
				// who is this?
				break
			}
			sh := bsh.SizeHint
			if r.Size.X <= sh.MaxSize.X && r.Size.X >= sh.MinSize.X {
				sh.PreferredSize.X = r.Size.X
			}
			if r.Size.Y <= sh.MaxSize.Y && r.Size.Y >= sh.MinSize.Y {
				sh.PreferredSize.Y = r.Size.Y
			}
			r.SetSizeHint(sh)
		}
	}
}
```

####Size hints

It's worth taking a moment here to discuss what ```uik.SizeHint```s are and how they travel between parents and children.

```go
type SizeHint struct {
	MinSize, PreferredSize, MaxSize geom.Coord
}
```

A ```uik.SizeHint``` tries to suggest to a component's parent how it ought to be sized. Anything larger than the min and smaller than the max will look nice, with the preferred as a default. The component's parent can then set the size to whatever it likes - these are just hints after all.

When a ```uik.Block``` is added to a ```uik.Foundation```, the ```uik.Foundation``` creates a channel that combines the ```uik.SizeHint```s coming in from the ```uik.Block``` with the ```uik.Block``` itself, so the parent only has to read from one channel to get all the size hints and know where they came from.

If that was confusing, the actual code that does this might make it clear.

```go
func (f *Foundation) AddBlock(b *Block) {
	/* ... */

	sizeHints := make(SizeHintChan, 1)
	go func(b *Block, sizeHints chan SizeHint) {
		for sh := range sizeHints {
			f.BlockSizeHints <- BlockSizeHint{
				SizeHint: sh,
				Block:    b,
			}
		}
	}(b, sizeHints)

	b.placementNotifications.Stack(placementNotification{
		Foundation: f,
		SizeHints:  sizeHints,
	})
}
```

Once the ```uik.Block``` has been placed in the ```uik.Foundation```, it gets a notification of the placement, complete with a channel just itching to take some ```uik.SizeHint```s. When the ```Radio``` calls ```.SetSizeHint(sh)```, it sets ```sh``` to be perpetually sent along that channel, until a new ```uik.SizeHint``` is set.

###Adding to the grid

Back to placing the buttons, we're ready to place them within the ```layouts.Grid```.

```go
func (r *Radio) makeButtons(options []string) {
	/* ... */

	r.buttons = make([]*Button, len(r.options))
	for i, option := range r.options {
		ob := NewButton(option)
		r.buttons[i] = ob

		r.radioGrid.Add <- layouts.BlockData{
			Block: &ob.Block,
			GridX: 0, GridY: i,
		}
	}
}
```

This will add each button in the first column, and in successive rows according to the order they appeared in ```.options```.

There's only one thing more we need to do to get buttons drawing and clicking.

```go
func (r *Radio) HandleEvents() {
	for {
		select {
		/* ... */
		case inv := <-r.BlockInvalidations:
			r.Invalidate(inv.Bounds...)
		}
	}
}
```

We need to pass on invalidations. An invalidation is a set of rectangles indicating which areas of the screen need to be redrawn. When a ```uik.Button``` is clicked, it invalidates its whole area, which sends a message to its parent. This message needs to trickle up all the way to the ```uik.WindowFoundation``` at the top. The ```uik.WindowFoundation``` collects all invalidations and, every so often, redraws components that lie within them.

###Mutual exclusion

The whole point of a radio group is that only one option can be selected at a time. We will now set up the code to monitor when a button is clicked, and to set it to the "selected" element.

The first thing to do is to get a notification when a button is clicked. And, to clean things up, something to tell the notification goroutine to shut down.

```go
type Radio struct {
	/* ... */
	
	buttons     []*Button
	buttonsDone []chan bool

	/* ... */
}
```

```go
func (r *Radio) makeButtons(options []string) {
	/* ... */

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

		r.radioGrid.Add <- layouts.BlockData{
			Block: &ob.Block,
			GridX: 0, GridY: i,
		}

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
}
```

Of course, we aren't actually monitoring ```.setSelection```, so let's set that up too.

```go
func (r *Radio) HandleEvents() {
	for {
		select {
		/* ... */
		case r.selection = <-r.setSelection:
			r.updateButtons()
		case r.getSelection <- r.selection:
		/* ... */
		}
	}
}
```

```go
func (r *Radio) makeButtons(options []string) {
	/* ... */

	r.updateButtons()
}
```

```go
func (r *Radio) updateButtons() {
	for i, b := range r.buttons {
		if i == r.selection {
			b.SetConfig <- ButtonConfig{
				Color: color.RGBA{110, 110, 110, 255},
			}
		} else {
			b.SetConfig <- ButtonConfig{}
		}
	}
	r.Invalidate()
}
```

###Setting up a notifier

Now that the ```Radio``` is maintaining its exclusion, we can set up a way to subscribe to be notified when a new selection is made, similar to the click notifications a ```uik.Button``` sends out.

First we can define a type that will contain all the necessary information and a chan type for it, and add a set of them to the ```Radio```. Then we'll add a way for things to subscribe.

```go
type RadioSelection struct {
	Index int
	Option string
}

type SelectionListener chan RadioSelection
```

```go
type Radio struct {
	/* ... */
	selectionListeners map[SelectionListener]bool
	addSelectionListener    chan SelectionListener
	AddSelectionListener    chan<- SelectionListener
	removeSelectionListener chan SelectionListener
	RemoveSelectionListener <-chan SelectionListener
}
```
```go
func (r *Radio) Initialize() {
	/* ... */
	r.selectionListeners = map[SelectionListener]bool{}
	r.addSelectionListener = make(chan SelectionListener, 1)
	r.AddSelectionListener = r.addSelectionListener
	r.removeSelectionListener = make(chan SelectionListener, 1)
	r.RemoveSelectionListener = r.removeSelectionListener
}
```
```go
func (r *Radio) HandleEvents() {
	for {
		select {
		/* ... */
		case selLis := <-r.addSelectionListener:
			r.selectionListeners[selLis] = true
		case selLis := <-r.removeSelectionListener:
			if r.selectionListeners[selLis] {
				delete(r.selectionListeners, selLis)
			}
		}
	}
}
```

With this infrastructure set up, we can send out the notification when a selection is made.

```go
func (r *Radio) HandleEvents() {
	for {
		select {
		/* ... */
		case r.selection = <-r.setSelection:
			r.updateButtons()
			for selLis := range r.selectionListeners {
				selLis <- RadioSelection{
					Index:  r.selection,
					Option: r.options[r.selection],
				}
			}
		/* ... */
		}
	}
}
```

And that's it! The result isn't quite the vision outlined above. To do that, I'd have to implement some sort of margins for components in a grid, and that is a (simple) task for another day.

The completed code is in go.uik/widgets/radio.go.

The following program demonstrates the new ```Radio``` widget.

```go
package main

import (
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.uik/layouts"
	"github.com/skelterjohn/go.uik/widgets"
	"github.com/skelterjohn/go.wde"
	"image/color"
)

func main() {
	go uikplay()
	wde.Run()
}

func uikplay() {

	w, err := uik.NewWindow(nil, 480, 320)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.W.SetTitle("go.uik")

	g := layouts.NewGrid(layouts.GridConfig{})

	rg := widgets.NewRadio([]string{"bread", "cake", "beheadings"})
	g.Add <- layouts.BlockData{
		Block:   &rg.Block,
		GridX:   0,
		GridY:   0,
		AnchorY: layouts.AnchorMin,
	}

	l := widgets.NewLabel(geom.Coord{100, 30}, widgets.LabelData{"text", 14, color.Black})
	g.Add <- layouts.BlockData{
		Block:   &l.Block,
		GridX:   1,
		GridY:   0,
		AnchorY: layouts.AnchorMin,
	}

	selLis := make(widgets.SelectionListener, 1)
	go func() {
		for sel := range selLis {
			l.SetConfig <- widgets.LabelData{
				Text:     fmt.Sprintf("Clicked option %d, %q", sel.Index, sel.Option),
				FontSize: 14,
				Color:    color.Black,
			}
		}
	}()
	rg.AddSelectionListener <- selLis

	w.Pane <- &g.Block

	w.Show()

	done := make(chan interface{}, 1)
	isDone := func(e interface{}) (accept, done bool) {
		_, accept = e.(uik.CloseEvent)
		done = accept
		return
	}
	w.Block.Subscribe <- uik.Subscription{isDone, done}

	<-done

	w.W.Close()

	wde.Stop()
}
```
