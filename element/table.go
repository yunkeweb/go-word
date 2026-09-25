package element

import "github.com/yunkeweb/go-word/style"

// Table is a table element (PHPWord Element\Table).
type Table struct {
	Base
	Style style.Table
	Rows  []*Row
	Width int
}

func (t *Table) Type() string { return "Table" }

// NewTable constructs a table.
func NewTable(st any) *Table {
	tbl := &Table{}
	switch v := st.(type) {
	case string:
		tbl.Style.StyleName = v
	case style.Table:
		tbl.Style = v
	case *style.Table:
		if v != nil {
			tbl.Style = *v
		}
	}
	return tbl
}

// AddRow appends a row with optional height in twips.
func (t *Table) AddRow(height ...int) *Row {
	r := &Row{}
	if len(height) > 0 {
		r.Style.Height = height[0]
	}
	if len(height) > 1 {
		// unused; PHPWord second arg is style
	}
	setParent(r, t)
	t.Rows = append(t.Rows, r)
	return r
}

// SetWidth sets the table width in twips (PHPWord Table::setWidth).
func (t *Table) SetWidth(w int) { t.Width = w; t.Style.Width = w }

// GetWidth returns the table width in twips.
func (t *Table) GetWidth() int { return t.Width }

// GetRows returns table rows.
func (t *Table) GetRows() []*Row { return t.Rows }

// CountColumns returns the maximum number of columns, counting grid spans.
func (t *Table) CountColumns() int {
	max := 0
	for _, r := range t.Rows {
		n := 0
		for _, c := range r.Cells {
			span := c.Style.GridSpan
			if span < 1 {
				span = 1
			}
			n += span
		}
		if n > max {
			max = n
		}
	}
	return max
}

// FindFirstDefinedCellWidths returns the first row's cell widths.
func (t *Table) FindFirstDefinedCellWidths() []int {
	for _, r := range t.Rows {
		if len(r.Cells) == 0 {
			continue
		}
		out := make([]int, len(r.Cells))
		for i, c := range r.Cells {
			out[i] = c.Width
		}
		return out
	}
	return nil
}

// AddCell is a PHPWord convenience that adds a cell to the last row,
// creating a row if needed.
func (t *Table) AddCell(width int, st ...any) *Cell {
	if len(t.Rows) == 0 {
		t.AddRow()
	}
	return t.Rows[len(t.Rows)-1].AddCell(width, st...)
}

// Row is a table row.
type Row struct {
	Base
	Style style.Row
	Cells []*Cell
}

func (r *Row) Type() string { return "Row" }

// GetCells returns the row's cells.
func (r *Row) GetCells() []*Cell { return r.Cells }

// GetHeight returns the row height in twips.
func (r *Row) GetHeight() int { return r.Style.Height }

// AddCell appends a cell with width in twips.
func (r *Row) AddCell(width int, st ...any) *Cell {
	c := NewCell(width, st...)
	setParent(c, r)
	r.Cells = append(r.Cells, c)
	return c
}

// Cell is a table cell and a container.
type Cell struct {
	Container
	Style style.Cell
	Width int
}

func (c *Cell) Type() string { return "Cell" }

// NewCell constructs a cell.
func NewCell(width int, st ...any) *Cell {
	c := &Cell{Width: width}
	c.Kind = "Cell"
	c.Style.Width = width
	if len(st) > 0 {
		switch v := st[0].(type) {
		case style.Cell:
			c.Style = v
			if c.Style.Width == 0 {
				c.Style.Width = width
			}
		case *style.Cell:
			if v != nil {
				c.Style = *v
			}
		}
	}
	return c
}
