package metadata

import "github.com/yunkeweb/go-word/style"

// Settings is document-level Word settings (PHPWord Metadata\Settings).
type Settings struct {
	ThemeFontLang          string
	ThemeFont              *style.Language
	HideGrammaticalErrors  bool
	HideSpellingErrors     bool
	TrackRevisions         bool
	DoNotTrackMoves        bool
	DoNotTrackFormatting   bool
	RevisionView           *TrackChangesView
	Zoom                   int
	ZoomPreset             string
	UpdateFields           bool
	MirrorMargins          bool
	EvenAndOddHeaders      bool
	AutoHyphenation        bool
	ConsecutiveHyphenLimit int
	HyphenationZone        int
	DoNotHyphenateCaps     bool
	BookFoldPrinting       bool
	DecimalSymbol          string
	ListSeparator          string
	ProofState             ProofState
	DocumentProtection     *Protection
}

// ProofState is spelling/grammar proofing state (PHPWord ComplexType\ProofState).
type ProofState struct {
	Spelling string // dirty, clean
	Grammar  string
}

const (
	ProofDirty = "dirty"
	ProofClean = "clean"
)

// TrackChangesView is revision-view flags (PHPWord ComplexType\TrackChangesView).
type TrackChangesView struct {
	Markup         *bool
	Comments       *bool
	InsDel         *bool
	Formatting     *bool
	InkAnnotations *bool
}

// Protection is document write-protection.
type Protection struct {
	Editing   string // readOnly, comments, trackedChanges, forms
	Password  string
	Algorithm string
	Salt      []byte
	SpinCount int
	Hash      string
}

// NewSettings returns default document settings.
func NewSettings() *Settings {
	return &Settings{
		DecimalSymbol: ".",
		ListSeparator: ",",
		Zoom:          100,
	}
}
