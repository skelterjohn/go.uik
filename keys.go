package uik

const (
	KeyArrowUp      = 111
	KeyArrowDown    = 116
	KeyArrowLeft    = 113
	KeyArrowRight   = 114
	KeyBackspace    = 22
	KeyDelete       = 119
	KeyLeftShift    = 50
	KeyRightShift   = 62
	KeyLeftControl  = 37
	KeyRightControl = 105
	KeyEscape       = 9
	KeyLeftAlt      = 64
	KeyRightAlt     = 108
	KeyRightSuper   = 134
	KeyReturn       = 36
)

var notglyphs = map[int]bool{
	KeyArrowUp:      true,
	KeyArrowDown:    true,
	KeyArrowLeft:    true,
	KeyArrowRight:   true,
	KeyBackspace:    true,
	KeyDelete:       true,
	KeyLeftShift:    true,
	KeyRightShift:   true,
	KeyLeftControl:  true,
	KeyRightControl: true,
	KeyEscape:       true,
	KeyLeftAlt:      true,
	KeyRightAlt:     true,
	KeyRightSuper:   true,
	KeyReturn:       true,
}

func IsGlyph(code int) (isglyph bool) {
	return !notglyphs[code]
}
