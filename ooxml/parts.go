package ooxml

import "encoding/xml"

// ContentTypes is [Content_Types].xml.
type ContentTypes struct {
	XMLName  xml.Name   `xml:"Types"`
	Xmlns    string     `xml:"xmlns,attr"`
	Defaults []Default  `xml:"Default"`
	Override []Override `xml:"Override"`
}

// Default is a content-type default by extension.
type Default struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// Override is a content-type override by part name.
type Override struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// NewContentTypes returns the required defaults for a .docx package.
func NewContentTypes() *ContentTypes {
	return &ContentTypes{
		Xmlns: NSContentTypes,
		Defaults: []Default{
			{Extension: "rels", ContentType: CTRels},
			{Extension: "xml", ContentType: "application/xml"},
		},
	}
}

func (c *ContentTypes) AddOverride(part, ct string) {
	c.Override = append(c.Override, Override{PartName: part, ContentType: ct})
}

func (c *ContentTypes) AddDefault(ext, ct string) {
	for _, d := range c.Defaults {
		if d.Extension == ext {
			return
		}
	}
	c.Defaults = append(c.Defaults, Default{Extension: ext, ContentType: ct})
}

// Relationships is a .rels part.
type Relationships struct {
	XMLName xml.Name       `xml:"Relationships"`
	Xmlns   string         `xml:"xmlns,attr"`
	Rel     []Relationship `xml:"Relationship"`
}

// Relationship is one package relationship.
type Relationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr,omitempty"`
}

func NewRelationships() *Relationships {
	return &Relationships{Xmlns: NSRelationships}
}

func (r *Relationships) Add(id, typ, target, mode string) {
	r.Rel = append(r.Rel, Relationship{ID: id, Type: typ, Target: target, TargetMode: mode})
}

// CoreProperties is docProps/core.xml.
type CoreProperties struct {
	XMLName        xml.Name `xml:"cp:coreProperties"`
	XmlnsCP        string   `xml:"xmlns:cp,attr"`
	XmlnsDC        string   `xml:"xmlns:dc,attr"`
	XmlnsDCTERMS   string   `xml:"xmlns:dcterms,attr"`
	XmlnsDCMIType  string   `xml:"xmlns:dcmitype,attr"`
	XmlnsXSI       string   `xml:"xmlns:xsi,attr"`
	Title          string   `xml:"dc:title,omitempty"`
	Subject        string   `xml:"dc:subject,omitempty"`
	Creator        string   `xml:"dc:creator,omitempty"`
	Keywords       string   `xml:"cp:keywords,omitempty"`
	Description    string   `xml:"dc:description,omitempty"`
	LastModifiedBy string   `xml:"cp:lastModifiedBy,omitempty"`
	Revision       string   `xml:"cp:revision,omitempty"`
	Created        *W3Time  `xml:"dcterms:created,omitempty"`
	Modified       *W3Time  `xml:"dcterms:modified,omitempty"`
	Category       string   `xml:"cp:category,omitempty"`
}

// W3Time is a dcterms timestamp with xsi:type.
type W3Time struct {
	XSIType string `xml:"xsi:type,attr"`
	Value   string `xml:",chardata"`
}

// AppProperties is docProps/app.xml.
type AppProperties struct {
	XMLName     xml.Name `xml:"Properties"`
	Xmlns       string   `xml:"xmlns,attr"`
	XmlnsVT     string   `xml:"xmlns:vt,attr"`
	Application string   `xml:"Application,omitempty"`
	DocSecurity int      `xml:"DocSecurity"`
	Pages       int      `xml:"Pages,omitempty"`
	Words       int      `xml:"Words,omitempty"`
	Company     string   `xml:"Company,omitempty"`
	Manager     string   `xml:"Manager,omitempty"`
	AppVersion  string   `xml:"AppVersion,omitempty"`
}

// CustomProperties is docProps/custom.xml.
type CustomProperties struct {
	XMLName xml.Name         `xml:"Properties"`
	Xmlns   string           `xml:"xmlns,attr"`
	XmlnsVT string           `xml:"xmlns:vt,attr"`
	Props   []CustomProperty `xml:"property"`
}

// CustomProperty is one custom property.
type CustomProperty struct {
	FmtID  string `xml:"fmtid,attr"`
	PID    int    `xml:"pid,attr"`
	Name   string `xml:"name,attr"`
	Lpwstr string `xml:"vt:lpwstr,omitempty"`
	I4     string `xml:"vt:i4,omitempty"`
	R8     string `xml:"vt:r8,omitempty"`
	Bool   string `xml:"vt:bool,omitempty"`
}
