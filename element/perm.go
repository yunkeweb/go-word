package element

import "strings"

// ST_EdGrp values used on w:permStart.
const (
	EdGrpNone            = "none"
	EdGrpEveryone        = "everyone"
	EdGrpAdministrators  = "administrators"
	EdGrpContributors    = "contributors"
	EdGrpEditors         = "editors"
	EdGrpOwners          = "owners"
	EdGrpCurrent         = "current"
)

// ParseAllowEdit maps a user-facing group or user name onto w:edGrp / w:ed.
func ParseAllowEdit(groupOrUser string) (edGrp, ed string) {
	t := strings.TrimSpace(groupOrUser)
	if t == "" {
		return EdGrpEveryone, ""
	}
	switch strings.ToLower(t) {
	case "everyone", "everybody", "all":
		return EdGrpEveryone, ""
	case EdGrpNone, EdGrpAdministrators, EdGrpContributors, EdGrpEditors, EdGrpOwners, EdGrpCurrent:
		return strings.ToLower(t), ""
	default:
		return "", t
	}
}

func (b *Base) setAllowEdit(groupOrUser string) {
	b.EditGroup, b.EditUser = ParseAllowEdit(groupOrUser)
}

// HasAllowEdit reports whether a permission range should be written.
func (b *Base) HasAllowEdit() bool {
	return b.EditGroup != "" || b.EditUser != ""
}

// AllowEdit marks this paragraph as an exception range in a protected document.
func (t *Text) AllowEdit(groupOrUser string) *Text {
	t.setAllowEdit(groupOrUser)
	return t
}

// AllowEdit marks this rich paragraph as an exception range in a protected document.
func (t *TextRun) AllowEdit(groupOrUser string) *TextRun {
	t.setAllowEdit(groupOrUser)
	return t
}

// AllowEdit marks this cell's contents as an exception range in a protected document.
func (c *Cell) AllowEdit(groupOrUser string) *Cell {
	c.setAllowEdit(groupOrUser)
	return c
}

// AllowEdit marks this table as an exception range in a protected document.
func (t *Table) AllowEdit(groupOrUser string) *Table {
	t.setAllowEdit(groupOrUser)
	return t
}
