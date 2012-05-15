Creating a widget with go.uik

As some of you may know, I've been working lately on a pure go GUI toolkit called go.uik (go UI kit). The project's repository is available on github (http://github.com/skelterjohn/go.uik).

One of my core goals with go.uik is to ensure that it doesn't get in its own way. Specifically, components cannot block each other. If component 1 decides to go off and do a web search in response to an event, this won't slow down component 2 on the other side of the window.

The way this is accomplished with go.uik is the restriction that all inter-component communication be done via non-blocking channel communication, and that each component runs its own goroutine. In many languages, this would be difficult to do, complicated to understand, and inefficient to execute.

Fortunately for us, go makes it trivial.

In this post, I will add a radio button/group to the (completely insufficient) widget toolkit available in the uik/widgets package. As I create it, I'll document what I need to do this this post in hopes that someone will read it and make other widgets for me, so I can finish writing my dissertation without stalling the world of go+GUI.

The first step is vision.

I don't like the classical circles-means-radio-buttons scheme. I'm going to make something a bit different, but still recognizable as a set of options from which you may choose only one.

Here's a sketch: https://www.dropbox.com/sh/89guvpsbtlcexf1/t4VN0RZ2oq/rb/vision.pdf

It will be a series of vertically stacked regular buttons, connected by a dark background, and the one that is currently selected will be shaded darker than the others.

Every component in a go.uik app that gets to draw to a certain part of the screen and receive user input needs to associate itself with a uik.Block. The correct way to do this is to simply embed the uik.Block in your widget type.

A uik.Block has all the channels that are needed to function in the program. You, as a widget designer, have to make sure that your widget is actually receiving from the correct channels. Fortunately, this is easy to do.

There is an important specialization of uik.Block, and that is uik.Foundation. The uik.Foundation type uses uik.Block just like any other widget should:

type Foundation struct {
  Block
  /* other stuff */
}

As a result, if you embed a uik.Foundation, you are also embedding a uik.Block.

The purpose of uik.Foundation is to support one or more uik.Blocks by positioning them correctly, forwarding user input, and compositing the visuals correctly.

Back to the radio group widget. First, I'll create the source file widgets/radio.go, and declare the type. Since our radio group widget will be making use of several buttons, we'll embed uik.Foundation.

    package widgets
    import (
        "github.com/skelterjohn/go.uik"
    )
    type Radio struct {
        uik.Foundation
    }

The next step is to decide how the interface will work. I may change my mind later, but for now I'm going to have the programmer pass the Radio a []string containing the options to display and the order in which they're displayed. I'll also let the programmer specify which one is currently selected, and provide a mechanism to notify the programmer when a new option is clicked.

Since the Radio, like all go.uik components, will be running its own goroutine, it's important to make sure data goes to and from it in a threadsafe manner. For go.uik, that means via a channel.


...go
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
...

This might look a bit heavy, but the channels that are spelled the same with different capitalizations are the same channel. So, each bit of configurable data needs a channel to set it and to get it.

In general, a widget designer might want to collect all configuration data into a single type MyWidgetConfig. In this case I think it's important to be able to set the current selection without also setting the available options, so I've split it into two pieces.

The next step is to create the widget's initialization functions.

//<code>
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

}
//</code>

Two initialization functions, here. One for when someone wants just a radio, and one for any time someone wants to use the radio type, including embedding it somewhere else. It's important that all initialization necessary go in the .Initialize() method, since it's the only function likely to be called by a widget that might want to embed the radio.

You'll also notice that the NewRadio() function, along with calling r.Initialize(), starts a goroutine running the (currently empty) .HandleEvents() method.

The .HandleEvents() method is the widgets core logic goroutine. Every widget must have one, and the correct place to invoke it is in the NewXYZ() function. The reason the NewXYZ() function doesn't contain the code in the .Initialize() method is because you only want to create the goroutine if the widget is not embedded in something else. If it *is* embedded somewhere else, and only .Initialize() is called, the embedding widget will create its own goroutine. You definitely don't want two goroutines fighting over the same channels, here.

The next obvious bit to fill out here is the .HandleEvents() method. We will start it off like the following.

//<code>
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
//</code>

The .HandleEvents() method just waits for input. Right now it listens only to the .UserEvents channel (which comes from uik.Block), though later we'll add a couple more. When it gets something from this channel, it does a type switch to see what it's looking at. We may later decide to do something with certain kinds of events. For now we'll let uik.Foundation's .HandleEvent() method deal with them. uik.Foundation's .HandleEvent() takes care of funneling events to the correct children.