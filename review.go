package word

import "github.com/yunkeweb/go-word/element"

// EnableTrackChanges turns on document-level revision tracking (w:trackRevisions).
func (d *Document) EnableTrackChanges(on bool) {
	d.settings.TrackRevisions = on
}

// TrackRevisions reports whether revision tracking is on.
func (d *Document) TrackRevisions() bool {
	return d.settings.TrackRevisions
}

// CommentOn annotates text in the last (or a new) section.
// extras is optional initials then date.
func (d *Document) CommentOn(text, body, author string, extras ...string) *element.Text {
	return d.lastOrNewSection().CommentOn(text, body, author, extras...)
}

// AddInsertion appends tracked-insert text (w:ins).
func (d *Document) AddInsertion(text, author, date string, styles ...any) *element.Text {
	return d.lastOrNewSection().AddInsertion(text, author, date, styles...)
}

// AddDeletion appends tracked-delete text (w:del).
func (d *Document) AddDeletion(text, author, date string, styles ...any) *element.Text {
	return d.lastOrNewSection().AddDeletion(text, author, date, styles...)
}
