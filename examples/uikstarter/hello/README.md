# go.uik getting started
## Overview
The purpose of this document is to show you some examples of how to use go.uik. You should should have some basic knowledge of go (golang).

Before we start coding we have to understand some basics: go.uik works on Windows, Linux and OSX. This is possible because it uses different back ends. There are three files in every project to manage those back ends.

* wde\_darwin.go
* wde\_linux.go
* wde\_windows.go

Each of them imports a specific back end, e.g.

```g
import _ "github.com/skelterjohn/go.wde/cocoa"
```

For now we just have to know that we need those tree files in the directory of our project.

## Hello World
This example shows how to create a window with the title "Hello World".

The first part is easy, we define the package and import the needed packages.

```go
package main

import (
    "fmt"
    "github.com/skelterjohn/geom"
    "github.com/skelterjohn/go.uik"
    "github.com/skelterjohn/go.wde"
)
```

Now we create a function which we'll later start as a goroutine. To create a new window we use `uik.NewWindow` and check if there are any errors. `uik.NewWindow` returns a pointer to a `WindowFoundation`, a `WindowFoundation` wraps a `wde.Window`.

```go
func hello() {
    w, err := uik.NewWindow(nil, 320, 480)
    if err != nil {
        fmt.Println(err)
        return
    }   
}
```

Now we have a window, but no title, so let's set one by using `w.W.SetTitle("Hello World")`. `w.W` is our window (remember, `w` is the `WindowFoundation`. `SetTitle` is obviously the function to set the window title.

The next line `w.Show()` makes our `WindowFoundation` visible on the screen.

```go
	w.W.SetTitle("Hello World")

    w.Show()
```

This is the code of our Hello World program:


```go
package main

import (
    "fmt"
    "github.com/skelterjohn/geom"
    "github.com/skelterjohn/go.uik"
    "github.com/skelterjohn/go.wde"
)

func main() {
    go hello()
    wde.Run()
}

func hello() {

    wbounds := geom.Rect{
        Max: geom.Coord{480, 320},
    }   
    w, err := uik.NewWindow(nil, int(wbounds.Max.X), int(wbounds.Max.Y))
    if err != nil {
        fmt.Println(err)
        return
    }   
    w.W.SetTitle("Hello World")

    w.Show()
}
```

You can build it (go run doesn't work, because go run doesn't care about the three wde_ files).

	$ go build /path/to/your/helloworld

After running this, you'll get a `hello` file in you current directory (to start it `./hello`). But there's a problem, we can't close the window, this will be solved next. For now just press `Ctrl + c` on the console where you started the `hello` program.

### The Close Event
The program works, but we don't end it properly. We have to add a close event to the window to end the program when the window is being closed.

First we create a channel of type `interface{}` (no specific type), buffer size is one. `isDone` is a function variable, the function itself gets an event, it checks if this event is of type `uid.CloseEvent` (type assertation). If the event is of type `uik.CloseEvent`, `accept` will be `true` and we're done.

```go
   done := make(chan interface{}, 1)
    isDone := func(e interface{}) (accept, done bool) {
        _, accept = e.(uik.CloseEvent)
        done = accept
        return
    }

```

The next line 

```go
	w.Block.Subscribe <- uik.Subscription{isDone, done}

```
does the subscription.

Finally we have our channel `done` that blocks, until the close event happened. After this we close our window and call `wde.Stop()` to end our back end.

```go
<-done

w.W.Close()

wde.Stop()
```

Finally we start `hello()` as go routine in `main()` and start the back end with `wde.Run()`.

```go
func main() {
    go hello()
    wde.Run()
}
```

This is the complete example:
```go
package main

import (
    "fmt"
    "github.com/skelterjohn/go.uik"
    "github.com/skelterjohn/go.wde"
)

func main() {
    go hello()
    wde.Run()
}

func hello() {
    w, err := uik.NewWindow(nil, 480, 320)
    if err != nil {
        fmt.Println(err)
        return
    }   
    w.W.SetTitle("Hello World")

    w.Show()

    // Here we set up a subscription on the window's close events.
    done := make(chan interface{}, 1)
    isDone := func(e interface{}) (accept, done bool) {
        _, accept = e.(uik.CloseEvent)
        done = accept
        return
    }   
    w.Block.Subscribe <- uik.Subscription{isDone, done}

    // once a close event comes in on the subscription, end the program
    <-done

    w.W.Close()

    wde.Stop()
}

```
