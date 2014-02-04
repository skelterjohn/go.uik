# go.uik getting started - Labels
## Overview
In this section we will learn how to use labels and how to place them on the screen using a layout engine.

## Labels
Labels are very basic elements to place text within our window. Creating a new label is very easy. To get a new label, we use the widgets package:

```g
import (
	"github.com/skelterjohn/go.uik/widgets"
)
```

Then call `widgets.NewLabel` with the following parameters:

- `geom.Coord{100, 50}` this is the size of the label
- `widgets.LabelConfig{"text", 14, color.Black}` contains the labels configuration information: Text to show, font size and font color

When we put everything together we get:
```g
l1 := widgets.NewLabel(geom.Coord{100, 50}, widgets.LabelConfig{"text 0", 14, color.Black})

```

All configuration information can be changed at any time. To change a labels configuration, use `SetConfig`

```g
l0.SetConfig(widgets.LabelConfig{"new text", 20, color.RGBA{0, 0, 255, 255}})
```

## Layout Engine
To place the label (or any other widgets), we use `layouts.Grid` (there will be other layout engines, but for now we use the this one). The `layouts.Grid` is a layout manager in the style of java swing's GridBagLayout. To understand how it places components, consider a grid with an infinite number of rows and columns. You then place your child components in cells by specifying their X and Y positions, and also how many rows and columns they occupy (with the top left being at the listed X and Y). Then every row and column in the grid is expanded to fit the components that occupy it. Since we're doing this on a computer with finite memory and processing power, the infinite number of rows and columns that have no occupants, and take up no space, will not be represented in the data structure.

First we need a `GridEngine`, this is where we'll place the widgets. Then we pass the `GridEngine` to `layouts.NewLayouter`. `layouts.NewLayouter` takes a `layouts.LayoutEngine` and returns a `layouts.Layouter`-Pointer which can be placed like a widget.


```g
ge := layouts.NewGridEngine(layouts.GridConfig{})
g := layouts.NewLayouter(ge)
```

## Placing the label
Now we add the widget on the grid:

```g
ge.Add(&l0.Block, layouts.GridComponent{
	GridX: 0, GridY: 0,
})
```
This places the label on the first column (GridX: 0) and the first row (GridY: 0). If you'd like to span over multiple columns/rows, use `ExtraX`/`ExtraY`:

```g
ge.Add(&l2.Block, layouts.GridComponent{
	GridX: 0, GridY: 0,
	ExtraX: 1,
})
```

To align a text use `AnchorRight`, `AnchorLeft`, `AnchorTop`, `AnchorBottom`

```g
ge.Add(&l2.Block, layouts.GridComponent{
	GridX: 0, GridY: 0,
	ExtraX: 1,
	AnchorRight: true,
})
```

There are three more attributes that you can (but don't have to) pass to a `GridComponent`: `MinSize`, `PreferredSize`, `MaxSize`. Anything larger than the min and smaller than the max will look nice, with the preferred as a default.

```g
ge.Add(&l2.Block, layouts.GridComponent{
	GridX: 0, GridY: 1,
    ExtraX: 1,
    AnchorRight: true,
    AnchorTop: true,
    MinSize: geom.Coord{100, 100},
    MaxSize: geom.Coord{480, 250},
    PreferredSize: geom.Coord{480, 220},
})

```
## Finish
Finally we just need to set the layout to the window pane
```g
w.SetPane(&g.Block)
```
The code can be found under examples/uikstarter/labels.
