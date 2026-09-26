package element

// CloneElement returns a deep copy of el (or nil).
func CloneElement(el Element) Element {
	if el == nil {
		return nil
	}
	switch v := el.(type) {
	case *Text:
		c := *v
		return &c
	case *TextBreak:
		return &TextBreak{}
	case *PageBreak:
		return &PageBreak{}
	case *Link:
		c := *v
		return &c
	case *Bookmark:
		c := *v
		return &c
	case *Title:
		c := *v
		if v.Run != nil {
			if r, ok := CloneElement(v.Run).(*TextRun); ok {
				c.Run = r
			}
		}
		return &c
	case *ListItem:
		c := *v
		return &c
	case *ListItemRun:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *TextRun:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *PreserveText:
		c := *v
		return &c
	case *CheckBox:
		c := *v
		return &c
	case *Field:
		c := *v
		if v.Properties != nil {
			c.Properties = map[string]string{}
			for k, val := range v.Properties {
				c.Properties[k] = val
			}
		}
		if v.Options != nil {
			c.Options = append([]string(nil), v.Options...)
		}
		return &c
	case *Line:
		c := *v
		return &c
	case *Shape:
		c := *v
		return &c
	case *TextBox:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *DMLShape:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *Formula:
		c := *v
		return &c
	case *TOC:
		c := *v
		return &c
	case *TextWatermark:
		c := *v
		return &c
	case *SDT:
		c := *v
		c.Container = cloneContainer(v.Container)
		c.ListItems = append([]SDTListItem(nil), v.ListItems...)
		c.ID = 0
		return &c
	case *Image:
		return cloneImage(v)
	case *Chart:
		return cloneChart(v)
	case *Table:
		return cloneTable(v)
	case *Footnote:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *Endnote:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *Comment:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *OLEObject:
		c := *v
		c.Media.Data = append([]byte(nil), v.Media.Data...)
		c.RelationID = 0
		return &c
	case *Header:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	case *Footer:
		c := *v
		c.Container = cloneContainer(v.Container)
		return &c
	default:
		return nil
	}
}

// CloneSection returns a deep copy of a section, including headers and footers.
func CloneSection(s *Section) *Section {
	if s == nil {
		return nil
	}
	out := &Section{Style: s.Style, FootnoteProperties: s.FootnoteProperties}
	out.Kind = "Section"
	out.Container = cloneContainer(s.Container)
	for _, h := range s.Headers {
		if c, ok := CloneElement(h).(*Header); ok {
			out.Headers = append(out.Headers, c)
			setParent(c, out)
		}
	}
	for _, f := range s.Footers {
		if c, ok := CloneElement(f).(*Footer); ok {
			out.Footers = append(out.Footers, c)
			setParent(c, out)
		}
	}
	return out
}

func cloneContainer(src Container) Container {
	dst := Container{Kind: src.Kind, Base: src.Base}
	for _, el := range src.Elements() {
		if c := CloneElement(el); c != nil {
			dst.add(c)
		}
	}
	return dst
}

func cloneImage(img *Image) *Image {
	c := *img
	c.Data = append([]byte(nil), img.Data...)
	c.Media.Data = c.Data
	c.RelationID = 0
	return &c
}

func cloneChart(ch *Chart) *Chart {
	c := *ch
	c.Categories = append([]string(nil), ch.Categories...)
	c.Values = append([]float64(nil), ch.Values...)
	c.Series = append([]ChartSeries(nil), ch.Series...)
	for i := range c.Series {
		c.Series[i].Categories = append([]string(nil), ch.Series[i].Categories...)
		c.Series[i].Values = append([]float64(nil), ch.Series[i].Values...)
	}
	c.Style.Colors = append([]string(nil), ch.Style.Colors...)
	return &c
}

func cloneTable(t *Table) *Table {
	out := &Table{Style: t.Style, Width: t.Width}
	out.Base = t.Base
	for _, r := range t.Rows {
		nr := &Row{Style: r.Style}
		setParent(nr, out)
		for _, cell := range r.Cells {
			nc := &Cell{Style: cell.Style, Width: cell.Width}
			nc.Kind = "Cell"
			nc.Container = cloneContainer(cell.Container)
			setParent(nc, nr)
			nr.Cells = append(nr.Cells, nc)
		}
		out.Rows = append(out.Rows, nr)
	}
	return out
}
