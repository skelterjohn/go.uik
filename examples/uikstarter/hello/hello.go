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
