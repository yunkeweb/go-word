package word

// DocumentOptions configures defaults for one document instance.
// Zero values inherit the package defaults used by New.
type DocumentOptions struct {
	DefaultFontName      string
	DefaultAsianFontName string
	DefaultFontSize      float64
	DefaultFontColor     string
}

// NewWithOptions creates a document with instance-local defaults.
func NewWithOptions(opts DocumentOptions) *Document {
	d := New()
	if opts.DefaultFontName != "" {
		d.defaultFontName = opts.DefaultFontName
	}
	if opts.DefaultAsianFontName != "" {
		d.defaultAsianFont = opts.DefaultAsianFontName
	}
	if opts.DefaultFontSize != 0 {
		d.defaultFontSize = opts.DefaultFontSize
	}
	if opts.DefaultFontColor != "" {
		d.defaultFontColor = opts.DefaultFontColor
	}
	return d
}
