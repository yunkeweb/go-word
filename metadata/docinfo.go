package metadata

import "time"

// DocInfo is core and extended document properties (PHPWord Metadata\DocInfo).
type DocInfo struct {
	Creator        string
	LastModifiedBy string
	Created        time.Time
	Modified       time.Time
	Title          string
	Description    string
	Subject        string
	Keywords       string
	Category       string
	Company        string
	Manager        string
	Custom         []CustomProperty
}

// CustomProperty is a docProps/custom.xml entry.
type CustomProperty struct {
	Name  string
	Type  string // string, int, float, date, bool
	Value string
}

// NewDocInfo returns properties with timestamps set to now.
func NewDocInfo() *DocInfo {
	now := time.Now().UTC()
	return &DocInfo{
		Creator:        "GoWord",
		LastModifiedBy: "GoWord",
		Created:        now,
		Modified:       now,
	}
}

func (d *DocInfo) SetCreator(v string)        { d.Creator = v }
func (d *DocInfo) SetLastModifiedBy(v string) { d.LastModifiedBy = v }
func (d *DocInfo) SetTitle(v string)          { d.Title = v }
func (d *DocInfo) SetDescription(v string)    { d.Description = v }
func (d *DocInfo) SetSubject(v string)        { d.Subject = v }
func (d *DocInfo) SetKeywords(v string)       { d.Keywords = v }
func (d *DocInfo) SetCategory(v string)       { d.Category = v }
func (d *DocInfo) SetCompany(v string)        { d.Company = v }
func (d *DocInfo) SetManager(v string)        { d.Manager = v }

func (d *DocInfo) SetCustomProperty(name, typ, value string) {
	for i := range d.Custom {
		if d.Custom[i].Name == name {
			d.Custom[i].Type = typ
			d.Custom[i].Value = value
			return
		}
	}
	d.Custom = append(d.Custom, CustomProperty{Name: name, Type: typ, Value: value})
}
