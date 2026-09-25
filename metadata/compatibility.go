package metadata

// Compatibility is OOXML compatibility mode (PHPWord Metadata\Compatibility).
type Compatibility struct {
	OOXMLVersion int // 12 = Word 2007, 14 = 2010, 15 = 2013
}

// NewCompatibility returns Word 2007 compatibility.
func NewCompatibility() *Compatibility {
	return &Compatibility{OOXMLVersion: 12}
}

func (c *Compatibility) SetOoxmlVersion(v int) { c.OOXMLVersion = v }
func (c *Compatibility) GetOoxmlVersion() int  { return c.OOXMLVersion }
