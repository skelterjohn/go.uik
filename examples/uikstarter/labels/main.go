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
    go hello()
    wde.Run()
}

func hello() {
    w, err := uik.NewWindow(nil, 480, 320)
    if err != nil {
        fmt.Println(err)
        return
    }
    w.W.SetTitle("Labels")

	ge := layouts.NewGridEngine(layouts.GridConfig{})
	g := layouts.NewLayouter(ge)

	l0 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"text 0", 14, color.Black})
	l1 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"text 1", 14, color.Black})
	l2 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"text 2", 14, color.Black})

	// label l0 is placed on the top left corner
	ge.Add(&l0.Block, layouts.GridComponent{
		GridX:0, GridY:0,
	})

	// label l1 is placed just right of l0
	ge.Add(&l1.Block, layouts.GridComponent{
		GridX:1, GridY:0,
	})

	// label l2 is placed on the second row and spans over two cells, it has a size of 480x220 per default and is aligned right and top
    ge.Add(&l2.Block, layouts.GridComponent{
    GridX: 0, GridY: 1,
    ExtraX: 1,
    AnchorRight: true,
    AnchorTop: true,
    MinSize: geom.Coord{100, 100},
    MaxSize: geom.Coord{480, 250},
    PreferredSize: geom.Coord{480, 220},
	})

	// change config of l0 to show a blue text of size 20
	l0.SetConfig(widgets.LabelConfig{"new text", 20, color.RGBA{0, 0, 255, 255}})

	w.SetPane(&g.Block)

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
