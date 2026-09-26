package element

import "testing"

func TestParseAllowEdit(t *testing.T) {
	grp, user := ParseAllowEdit("")
	if grp != EdGrpEveryone || user != "" {
		t.Fatal("empty")
	}
	grp, user = ParseAllowEdit("Everyone")
	if grp != EdGrpEveryone {
		t.Fatal("Everyone")
	}
	grp, user = ParseAllowEdit("editors")
	if grp != EdGrpEditors {
		t.Fatal("editors")
	}
	grp, user = ParseAllowEdit("alice")
	if grp != "" || user != "alice" {
		t.Fatal("user")
	}
}

func TestAllowEditSetters(t *testing.T) {
	tx := NewText("x", nil, nil).AllowEdit("Everyone")
	if !tx.HasAllowEdit() || tx.EditGroup != EdGrpEveryone {
		t.Fatal("text")
	}
	tr := NewTextRun(nil).AllowEdit("owners")
	if tr.EditGroup != EdGrpOwners {
		t.Fatal("textrun")
	}
	cell := NewCell(1000).AllowEdit("bob")
	if cell.EditUser != "bob" {
		t.Fatal("cell")
	}
	tbl := NewTable(nil).AllowEdit("Everyone")
	if !tbl.HasAllowEdit() {
		t.Fatal("table")
	}
}
