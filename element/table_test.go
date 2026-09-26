package element

import (
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestTableAdvancedSetters(t *testing.T) {
	tbl := NewTable(style.Table{Width: 4000})
	row := tbl.AddRow(200)
	if tbl.SetHeaderRow(row) != tbl {
		t.Fatal("SetHeaderRow chain")
	}
	if !row.IsHeader() || !row.Style.TblHeader || !row.Style.Header {
		t.Fatal("header flags")
	}
	if row.SetCantSplit(true) != row || !row.IsCantSplit() {
		t.Fatal("cantSplit")
	}
	row.SetHeader(false).SetCantSplit(false)
	if row.IsHeader() || row.IsCantSplit() || row.Style.TblHeader {
		t.Fatal("cleared flags")
	}

	cell := row.AddCell(2000)
	if cell.SetVAlign("CENTER") != cell || cell.Style.VAlign != style.VAlignCenter {
		t.Fatal("SetVAlign")
	}
	cell.SetVAlign("middle")
	if cell.Style.VAlign != style.VAlignCenter {
		t.Fatal("middle")
	}
	cell.SetVerticalAlignment("bottom")
	if cell.Style.VAlign != style.VAlignBottom {
		t.Fatal("SetVerticalAlignment")
	}
	if cell.SetTextDirection("TbRl") != cell || cell.Style.TextDir != style.TextDirectionTbRl {
		t.Fatal("SetTextDirection")
	}
	cell.SetTextDirection("btLr")
	if cell.Style.TextDir != style.TextDirectionBtLr {
		t.Fatal("btLr")
	}
}

func TestTableSetHeaderRowNil(t *testing.T) {
	tbl := NewTable(nil)
	if tbl.SetHeaderRow(nil) != tbl {
		t.Fatal("nil row")
	}
}
