package element

import "sort"

// SDT control types (CT_SdtPr choice). PHPWord uses plainText|comboBox|dropDownList|date.
const (
	SDTTypePlainText = "plainText"
	SDTTypeComboBox  = "comboBox"
	SDTTypeDropDown  = "dropDownList"
	SDTTypeDate      = "date"
	SDTTypeCheckbox  = "checkbox"
	SDTTypeRichText  = "richText"
)

// Checkbox glyphs used in w:sdtContent (Word 2010 w14:checkbox).
const (
	SDTCheckboxUnchecked = "\u2610" // ☐
	SDTCheckboxChecked   = "\u2612" // ☒
	sdtCheckboxFont      = "MS Gothic"
)

// SDTListItem is one w:listItem on a drop-down or combo box.
type SDTListItem struct {
	Value       string
	DisplayText string
}

// SDT is a structured document tag (w:sdt).
//
// Child order in the package is w:sdt → w:sdtPr → w:sdtContent (CT_SdtBlock / CT_SdtRun).
type SDT struct {
	Container
	SDTType       string
	Alias         string
	Tag           string
	Value         string
	Placeholder   string
	DateFormat    string
	Locale        string
	Lock          string
	ID            int
	Checked       bool
	ShowingPlcHdr bool
	MultiLine     bool
	ListItems     []SDTListItem
}

func (s *SDT) Type() string { return "SDT" }

// SetAlias sets w:alias.
func (s *SDT) SetAlias(alias string) *SDT { s.Alias = alias; return s }

// SetTag sets w:tag.
func (s *SDT) SetTag(tag string) *SDT { s.Tag = tag; return s }

// SetValue sets the visible sdtContent text.
func (s *SDT) SetValue(value string) *SDT {
	s.Value = value
	s.ShowingPlcHdr = false
	return s
}

// SetListItems replaces drop-down / combo-box entries.
func (s *SDT) SetListItems(items []SDTListItem) *SDT {
	s.ListItems = append([]SDTListItem(nil), items...)
	return s
}

// SetChecked sets the w14:checkbox state and the content glyph.
func (s *SDT) SetChecked(checked bool) *SDT {
	s.Checked = checked
	if checked {
		s.Value = SDTCheckboxChecked
	} else {
		s.Value = SDTCheckboxUnchecked
	}
	return s
}

// SetDateFormat sets w:dateFormat (for example yyyy-MM-dd).
func (s *SDT) SetDateFormat(format string) *SDT { s.DateFormat = format; return s }

// AddSDT appends a structured document tag (PHPWord addSDT).
func (c *Container) AddSDT(typ string) *SDT {
	s := &SDT{SDTType: NormalizeSDTType(typ)}
	c.add(s)
	return s
}

// AddSDTText appends a plain-text content control (w:text).
func (c *Container) AddSDTText(alias, tag, placeholderText string) *SDT {
	s := &SDT{
		SDTType:       SDTTypePlainText,
		Alias:         alias,
		Tag:           tag,
		Placeholder:   placeholderText,
		Value:         placeholderText,
		ShowingPlcHdr: placeholderText != "",
	}
	c.add(s)
	return s
}

// AddSDTDropdown appends a drop-down list (w:dropDownList + w:listItem).
// options maps value → display text; keys are written in sorted order.
func (c *Container) AddSDTDropdown(alias, tag string, options map[string]string) *SDT {
	items := sdtItemsFromMap(options)
	s := &SDT{
		SDTType:   SDTTypeDropDown,
		Alias:     alias,
		Tag:       tag,
		ListItems: items,
	}
	if len(items) > 0 {
		s.Value = items[0].DisplayText
		if s.Value == "" {
			s.Value = items[0].Value
		}
	} else {
		s.Value = "Choose an item."
		s.ShowingPlcHdr = true
	}
	c.add(s)
	return s
}

// AddSDTDate appends a date picker (w:date). dateFormat is w:dateFormat, default yyyy-MM-dd.
func (c *Container) AddSDTDate(alias, tag, dateFormat string) *SDT {
	if dateFormat == "" {
		dateFormat = "yyyy-MM-dd"
	}
	s := &SDT{
		SDTType:       SDTTypeDate,
		Alias:         alias,
		Tag:           tag,
		DateFormat:    dateFormat,
		Locale:        "en-US",
		Placeholder:   "Click or tap to enter a date.",
		Value:         "Click or tap to enter a date.",
		ShowingPlcHdr: true,
	}
	c.add(s)
	return s
}

// AddSDTCheckbox appends a Word 2010 checkbox (w14:checkbox).
func (c *Container) AddSDTCheckbox(alias, tag string, checked bool) *SDT {
	s := &SDT{
		SDTType: SDTTypeCheckbox,
		Alias:   alias,
		Tag:     tag,
		Checked: checked,
	}
	if checked {
		s.Value = SDTCheckboxChecked
	} else {
		s.Value = SDTCheckboxUnchecked
	}
	c.add(s)
	return s
}

// NormalizeSDTType maps PHPWord / shorthand type names onto the OpenXML control kind.
func NormalizeSDTType(typ string) string {
	switch typ {
	case SDTTypePlainText, "text", "plain":
		return SDTTypePlainText
	case SDTTypeComboBox:
		return SDTTypeComboBox
	case SDTTypeDropDown, "dropDown":
		return SDTTypeDropDown
	case SDTTypeDate:
		return SDTTypeDate
	case SDTTypeCheckbox, "checkBox":
		return SDTTypeCheckbox
	case SDTTypeRichText, "rich":
		return SDTTypeRichText
	default:
		if typ == "" {
			return SDTTypeRichText
		}
		return typ
	}
}

func sdtItemsFromMap(options map[string]string) []SDTListItem {
	if len(options) == 0 {
		return nil
	}
	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	items := make([]SDTListItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, SDTListItem{Value: k, DisplayText: options[k]})
	}
	return items
}

// CheckboxFont is the East-Asian font Word expects for w14:checkbox glyphs.
func CheckboxFont() string { return sdtCheckboxFont }
