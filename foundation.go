package uik

import (
	"image"
	"github.com/skelterjohn/go.wde"
)

// The foundation type is for channeling events to children, and passing along
// draw calls.
type Foundation struct {
	Block
	Children    []*Block
	DrawBuffers chan image.Image

	// this block currently has keyboard priority
	KeyboardBlock    *Block
}

func (f *Foundation) MakeChannels() {
	f.Block.MakeChannels()
	f.DrawBuffers = make(chan image.Image)
}

func (f *Foundation) AddBlock(b *Block) {
	// TODO: place the block somewhere clever
	// TODO: resize the block in a clever way
	f.Children = append(f.Children, b)
	b.Parent = f
	b.ParentDrawBuffer = f.DrawBuffers
}

func (f *Foundation) BlockForCoord(p Coord) (b *Block) {
	// quad-tree one day?
	for _, bl := range f.Children {
		if bl.BoundsInParent().Contains(p) {
			b = bl
			return
		}
	}
	return
}

func (f *Foundation) handleRedraw() {
	for dirtyBounds := range f.Redraw {
		dirtyBounds.Min.X -= f.Min.X
		dirtyBounds.Min.Y -= f.Min.Y
		f.Parent.Redraw <- dirtyBounds
	}
}

// dispense events to children, as appropriate
func (f *Foundation) handleEvents() {
	f.ListenedChannels[f.CloseEvents] = true
	f.ListenedChannels[f.MouseDownEvents] = true
	f.ListenedChannels[f.MouseUpEvents] = true

	var dragOriginBlocks = map[wde.Button]*Block{}
	// drag and up events for the same button get sent to the origin as well

	for {
		select {
		case e := <-f.CloseEvents:
			for _, b := range f.Children {
				b.allEventsIn <- e
			}
		case e := <-f.MouseDownEvents:
			b := f.BlockForCoord(e.Loc)
			if b == nil {
				break
			}
			dragOriginBlocks[e.Which] = b
			e.Loc.X -= b.Min.X
			e.Loc.Y -= b.Min.Y
			b.allEventsIn <- e
		case e := <-f.MouseUpEvents:
			b := f.BlockForCoord(e.Loc)
			if b != nil {
				be := e
				be.Loc.X -= b.Min.X
				be.Loc.Y -= b.Min.Y
				b.allEventsIn <- be
			}
			if origin, ok := dragOriginBlocks[e.Which]; ok && origin != b {
				oe := e
				oe.Loc.X -= origin.Min.X
				oe.Loc.Y -= origin.Min.Y
				origin.allEventsIn <- oe
			}

		case dr := <-f.Draw:
			bgc := f.PrepareBuffer()
			if f.Paint != nil {
				f.Paint(bgc)
			}
			for _, child := range f.Children {
				translatedDirty := dr.Dirty
				translatedDirty.Min.X -= child.Min.X
				translatedDirty.Min.Y -= child.Min.Y

				cdr := DrawRequest{
					Dirty: translatedDirty,
				}
				child.Draw <- cdr

			}
		case buffer := <-f.DrawBuffers:
			bgc := f.PrepareBuffer()
			bgc.DrawImage(buffer)
			if f.ParentDrawBuffer != nil {
				f.ParentDrawBuffer <- f.Buffer
			}

		}
	}
}