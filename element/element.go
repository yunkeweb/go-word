package element

// Element is a document-model node.
type Element interface {
	Type() string
	Parent() Element
	setParent(Element)
}

// Base is embedded by every concrete element.
type Base struct {
	parent       Element
	RelationID   int
	ElementID    string
	DocPart      string
	SectionID    int
	TrackChange  *TrackChange
	CommentStart *Comment
	CommentEnd   *Comment
	EditGroup    string // w:permStart w:edGrp (everyone, editors, …)
	EditUser     string // w:permStart w:ed
}

func (b *Base) SetTrackChange(tc *TrackChange) { b.TrackChange = tc }
func (b *Base) GetTrackChange() *TrackChange   { return b.TrackChange }
func (b *Base) SetChangeInfo(typ, author, date string) {
	b.TrackChange = &TrackChange{ChangeType: typ, Author: author, Date: date}
}
func (b *Base) SetCommentRangeStart(c *Comment) { b.CommentStart = c }
func (b *Base) GetCommentRangeStart() *Comment  { return b.CommentStart }
func (b *Base) SetCommentRangeEnd(c *Comment)   { b.CommentEnd = c }
func (b *Base) GetCommentRangeEnd() *Comment    { return b.CommentEnd }

func (b *Base) Parent() Element     { return b.parent }
func (b *Base) setParent(p Element) { b.parent = p }

func setParent(el Element, p Element) {
	if el != nil {
		el.setParent(p)
	}
}

// Media holds binary media attached to an image or OLE object.
type Media struct {
	MediaType string // image, object
	Target    string
	Source    string
	Data      []byte
	Ext       string
	IsURL     bool
}
