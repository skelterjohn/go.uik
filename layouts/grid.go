package layouts

type Align uint16

const (
	AlignLeft Align = 1 << iota
	AlignRight
	AlignTop
	AlignBottom
	AlignHorizCenter
	AlignVertCenter
	AlignCenter = AlignHorizCenter | AlignVertCenter
)

type GridConstraints struct {
	Align        Align
	SpanX, SpanY int
}

type Grid struct {
}
