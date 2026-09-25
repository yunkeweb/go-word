package word

import "github.com/yunkeweb/go-word/element"

func walkElement(el element.Element, fn func(element.Element)) {
	if el == nil {
		return
	}
	fn(el)
	for _, c := range childElements(el) {
		walkElement(c, fn)
	}
}

func childElements(el element.Element) []element.Element {
	switch v := el.(type) {
	case *element.Section:
		out := append([]element.Element{}, v.Elements()...)
		for _, h := range v.Headers {
			out = append(out, h)
		}
		for _, f := range v.Footers {
			out = append(out, f)
		}
		return out
	case *element.Header:
		return v.Elements()
	case *element.Footer:
		return v.Elements()
	case *element.Cell:
		return v.Elements()
	case *element.TextRun:
		return v.Elements()
	case *element.ListItemRun:
		return v.Elements()
	case *element.TextBox:
		return v.Elements()
	case *element.Footnote:
		return v.Elements()
	case *element.Endnote:
		return v.Elements()
	case *element.SDT:
		return v.Elements()
	case *element.Comment:
		return v.Elements()
	case *element.Table:
		var out []element.Element
		for _, r := range v.Rows {
			out = append(out, r)
		}
		return out
	case *element.Row:
		var out []element.Element
		for _, c := range v.Cells {
			out = append(out, c)
		}
		return out
	default:
		return nil
	}
}

func walkDocument(d *Document, fn func(element.Element)) {
	for _, sec := range d.sections {
		walkElement(sec, fn)
	}
}

func elementText(el element.Element) string {
	switch v := el.(type) {
	case *element.Text:
		return v.Content
	case *element.Link:
		return v.Text
	case *element.Title:
		return v.Text
	case *element.ListItem:
		return v.Text
	case *element.PreserveText:
		return v.Content
	case *element.CheckBox:
		return v.Content
	default:
		return ""
	}
}
