package style

// NumberingLevel is one level of a numbering definition.
type NumberingLevel struct {
	Start     int
	Format    string
	Text      string // e.g. "%1."
	Alignment string
	TabPos    int
	Left      int
	Hanging   int
	Font      string
	Hint      string
	Suffix    string // tab, space, nothing
}

// Numbering is a multi-level numbering style (PHPWord Style\Numbering).
type Numbering struct {
	Type   string // singleLevel, multilevel, hybridMultilevel
	Levels []NumberingLevel
}

// ListItem is the list style attached to a list paragraph.
type ListItem struct {
	ListType string
	NumId    int
	Depth    int
	Format   string
}

// List type names used by PHPWord.
const (
	ListTypeNone       = "none"
	ListTypeBullet     = "bullet"
	ListTypeNumber     = "number"
	ListTypeNumberNE   = "numbering"
	ListTypeMultilevel = "multilevel"
)
