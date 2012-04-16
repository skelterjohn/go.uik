package uik

type Coord struct {
	X, Y float64
}

type Bounds struct {
	Min, Max Coord
}

func (b Bounds) Contains(c Coord) bool {
	return c.X >= b.Min.X && c.Y >= b.Min.Y && c.X <= b.Max.X && c.Y <= b.Max.Y
}
