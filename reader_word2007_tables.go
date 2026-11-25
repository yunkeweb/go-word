package word

import (
	"encoding/xml"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

type wordTableMeasure struct {
	Width int    `xml:"w,attr"`
	Unit  string `xml:"type,attr"`
}

type wordTableBorder struct {
	Style      string `xml:"val,attr"`
	Size       int    `xml:"sz,attr"`
	Color      string `xml:"color,attr"`
	ThemeColor string `xml:"themeColor,attr"`
	ThemeTint  string `xml:"themeTint,attr"`
	ThemeShade string `xml:"themeShade,attr"`
	Space      int    `xml:"space,attr"`
}

type wordTableBorders struct {
	Top     *wordTableBorder `xml:"top"`
	Left    *wordTableBorder `xml:"left"`
	Right   *wordTableBorder `xml:"right"`
	Bottom  *wordTableBorder `xml:"bottom"`
	InsideH *wordTableBorder `xml:"insideH"`
	InsideV *wordTableBorder `xml:"insideV"`
}

type wordTableMargins struct {
	Top    *wordTableMeasure `xml:"top"`
	Left   *wordTableMeasure `xml:"left"`
	Right  *wordTableMeasure `xml:"right"`
	Bottom *wordTableMeasure `xml:"bottom"`
}

type wordTableProperties struct {
	Style  *wordValue        `xml:"tblStyle"`
	Width  *wordTableMeasure `xml:"tblW"`
	Layout *struct {
		Type string `xml:"type,attr"`
	} `xml:"tblLayout"`
	Indent      *wordTableMeasure `xml:"tblInd"`
	CellSpacing *wordTableMeasure `xml:"tblCellSpacing"`
	Borders     *wordTableBorders `xml:"tblBorders"`
	Shading     *wordShading      `xml:"shd"`
	CellMargins *wordTableMargins `xml:"tblCellMar"`
	BidiVisual  *wordValue        `xml:"bidiVisual"`
}

type wordTableRowProperties struct {
	Header    *wordValue `xml:"tblHeader"`
	CantSplit *wordValue `xml:"cantSplit"`
	Height    *struct {
		Value int    `xml:"val,attr"`
		Rule  string `xml:"hRule,attr"`
	} `xml:"trHeight"`
}

type wordTableCellProperties struct {
	Width         *wordTableMeasure `xml:"tcW"`
	GridSpan      *wordValue        `xml:"gridSpan"`
	VMerge        *wordValue        `xml:"vMerge"`
	VAlign        *wordValue        `xml:"vAlign"`
	TextDirection *wordValue        `xml:"textDirection"`
	NoWrap        *wordValue        `xml:"noWrap"`
	Shading       *wordShading      `xml:"shd"`
	Borders       *wordTableBorders `xml:"tcBorders"`
	Margins       *wordTableMargins `xml:"tcMar"`
}

func (p wordTableProperties) apply(t *element.Table, theme *wordTheme) {
	if t == nil {
		return
	}
	if p.Style != nil {
		t.Style.StyleName = p.Style.Val
	}
	if p.Width != nil {
		t.Style.Width, t.Width, t.Style.Unit = p.Width.Width, p.Width.Width, p.Width.Unit
	}
	if p.Layout != nil {
		t.Style.Layout = p.Layout.Type
	}
	if p.Indent != nil {
		t.Style.Indent = p.Indent.Width
	}
	if p.CellSpacing != nil {
		t.Style.CellSpacingVal = p.CellSpacing.Width
		if p.CellSpacing.Unit == "dxa" {
			t.Style.CellSpacing = p.CellSpacing.Width
		}
	}
	if p.Borders != nil {
		t.Style.Borders = p.Borders.styleBorders(theme)
	}
	if p.Shading != nil {
		t.Style.Shading = resolveWordShading(*p.Shading, theme)
	}
	if p.CellMargins != nil {
		t.Style.CellMarginTop = tableMargin(p.CellMargins.Top)
		t.Style.CellMarginLeft = tableMargin(p.CellMargins.Left)
		t.Style.CellMarginRight = tableMargin(p.CellMargins.Right)
		t.Style.CellMarginBottom = tableMargin(p.CellMargins.Bottom)
	}
	if p.BidiVisual != nil {
		t.Style.BidiVisual = wordOn(p.BidiVisual.Val)
	}
}

func (p wordTableRowProperties) apply(row *element.Row) {
	if row == nil {
		return
	}
	if p.Header != nil {
		row.SetHeader(wordOn(p.Header.Val))
	}
	if p.CantSplit != nil {
		row.Style.CantSplit = wordOn(p.CantSplit.Val)
	}
	if p.Height != nil {
		row.Style.Height = p.Height.Value
		row.Style.Rule = p.Height.Rule
	}
}

func (p wordTableCellProperties) apply(cell *element.Cell, theme *wordTheme) {
	if cell == nil {
		return
	}
	if p.Width != nil {
		cell.Width, cell.Style.Width, cell.Style.Unit = p.Width.Width, p.Width.Width, p.Width.Unit
	}
	if p.GridSpan != nil {
		if span := atoi(p.GridSpan.Val); span > 1 {
			cell.Style.GridSpan = span
		}
	}
	if p.VMerge != nil {
		cell.Style.VMerge = p.VMerge.Val
		if cell.Style.VMerge == "" {
			cell.Style.VMerge = "continue"
		}
	}
	if p.VAlign != nil {
		cell.Style.VAlign = p.VAlign.Val
	}
	if p.TextDirection != nil {
		cell.Style.TextDir = p.TextDirection.Val
	}
	if p.NoWrap != nil {
		cell.Style.NoWrap = wordOn(p.NoWrap.Val)
	}
	if p.Shading != nil {
		cell.Style.Shading = resolveWordShading(*p.Shading, theme)
	}
	if p.Borders != nil {
		cell.Style.Borders = p.Borders.styleBorders(theme)
	}
	if p.Margins != nil {
		cell.Style.PaddingTop = tableMargin(p.Margins.Top)
		cell.Style.PaddingLeft = tableMargin(p.Margins.Left)
		cell.Style.PaddingRight = tableMargin(p.Margins.Right)
		cell.Style.PaddingBottom = tableMargin(p.Margins.Bottom)
	}
}

func (p *wordTableBorders) styleBorders(theme *wordTheme) style.Borders {
	if p == nil {
		return style.Borders{}
	}
	return style.Borders{
		Top: tableBorder(p.Top, theme), Left: tableBorder(p.Left, theme), Right: tableBorder(p.Right, theme),
		Bottom: tableBorder(p.Bottom, theme), InsideH: tableBorder(p.InsideH, theme), InsideV: tableBorder(p.InsideV, theme),
	}
}

func tableBorder(b *wordTableBorder, theme *wordTheme) style.Border {
	if b == nil {
		return style.Border{}
	}
	return style.Border{Style: b.Style, Size: b.Size, Color: resolveWordColor(theme, b.ThemeColor, b.ThemeTint, b.ThemeShade, b.Color), Space: b.Space}
}

func tableMargin(m *wordTableMeasure) int {
	if m == nil || m.Unit != "dxa" {
		return 0
	}
	return m.Width
}

func decodeWordTableProperties(dec *xml.Decoder, start xml.StartElement, dst any) error {
	return dec.DecodeElement(dst, &start)
}
