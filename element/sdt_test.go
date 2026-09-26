package element

import "testing"

func TestSDTConstructors(t *testing.T) {
	sec := NewSection(1, nil)
	tx := sec.AddSDTText("Name", "name", "Type here")
	if tx.Type() != "SDT" || tx.SDTType != SDTTypePlainText || !tx.ShowingPlcHdr {
		t.Fatal("text")
	}
	dd := sec.AddSDTDropdown("Dept", "dept", map[string]string{"b": "Beta", "a": "Alpha"})
	if dd.SDTType != SDTTypeDropDown || len(dd.ListItems) != 2 {
		t.Fatal("dropdown")
	}
	if dd.ListItems[0].Value != "a" || dd.ListItems[1].Value != "b" {
		t.Fatalf("sorted keys %+v", dd.ListItems)
	}
	dt := sec.AddSDTDate("When", "when", "")
	if dt.DateFormat != "yyyy-MM-dd" || dt.Locale != "en-US" {
		t.Fatal("date defaults")
	}
	cb := sec.AddSDTCheckbox("Ok", "ok", true)
	if !cb.Checked || cb.Value != SDTCheckboxChecked {
		t.Fatal("checkbox")
	}
	cb.SetChecked(false)
	if cb.Checked || cb.Value != SDTCheckboxUnchecked {
		t.Fatal("uncheck")
	}
}

func TestNormalizeSDTType(t *testing.T) {
	cases := map[string]string{
		"":             SDTTypeRichText,
		"rich":         SDTTypeRichText,
		"plain":        SDTTypePlainText,
		"text":         SDTTypePlainText,
		"plainText":    SDTTypePlainText,
		"dropDown":     SDTTypeDropDown,
		"dropDownList": SDTTypeDropDown,
		"comboBox":     SDTTypeComboBox,
		"date":         SDTTypeDate,
		"checkbox":     SDTTypeCheckbox,
		"checkBox":     SDTTypeCheckbox,
	}
	for in, want := range cases {
		if got := NormalizeSDTType(in); got != want {
			t.Fatalf("%q -> %q want %q", in, got, want)
		}
	}
}

func TestSDTCloneZerosID(t *testing.T) {
	s := &SDT{SDTType: SDTTypePlainText, ID: 42, ListItems: []SDTListItem{{Value: "1", DisplayText: "One"}}}
	c, ok := CloneElement(s).(*SDT)
	if !ok || c.ID != 0 {
		t.Fatal("clone id")
	}
	if len(c.ListItems) != 1 || c.ListItems[0].Value != "1" {
		t.Fatal("clone items")
	}
	c.ListItems[0].Value = "x"
	if s.ListItems[0].Value != "1" {
		t.Fatal("alias")
	}
}
