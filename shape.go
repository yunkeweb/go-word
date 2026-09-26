package word

import (
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

// DrawingML preset geometry names.
type ShapeType string

const (
	ShapeRect      ShapeType = "rect"
	ShapeRoundRect ShapeType = "roundRect"
	ShapeArrow     ShapeType = "rightArrow"
	ShapeTextBox   ShapeType = "textBox"
)

// ShapeOptions controls DrawingML fill, outline, size (EMU) and text.
type ShapeOptions struct {
	Width     int
	Height    int
	FillColor string
	LineColor string
	LineWidth int
	Text      string
	Font      style.Font
}

func shapePrst(t ShapeType) string {
	switch t {
	case ShapeRoundRect:
		return "roundRect"
	case ShapeArrow:
		return "rightArrow"
	case ShapeRect, ShapeTextBox, "":
		return "rect"
	default:
		return string(t)
	}
}

// AddShape appends a DrawingML shape (rectangle, rounded rectangle, arrow, or text box)
// to the last (or a new) section.
func (d *Document) AddShape(shapeType ShapeType, opts ShapeOptions) *element.DMLShape {
	prst := shapePrst(shapeType)
	sh := d.lastOrNewSection().AddDMLShape(prst, opts.Width, opts.Height, opts.FillColor, opts.LineColor, opts.LineWidth)
	text := opts.Text
	if text == "" && shapeType == ShapeTextBox {
		text = " "
	}
	if text != "" {
		if opts.Font.Size != 0 || opts.Font.Name != "" || opts.Font.Color != "" || opts.Font.Bold {
			sh.AddText(text, opts.Font)
		} else {
			sh.AddText(text)
		}
	}
	return sh
}
