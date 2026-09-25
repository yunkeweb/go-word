package word

// Package-level defaults (PHPWord Settings), stored per process.
// Document-specific settings live on Document.Settings().

const (
	UnitTwip  = "twip"
	UnitCM    = "cm"
	UnitMM    = "mm"
	UnitInch  = "inch"
	UnitPoint = "point"
	UnitPica  = "pica"
)

var (
	defaultFontName      = "Arial"
	defaultAsianFontName = "Times New Roman"
	defaultFontSize      = 10.0
	defaultFontColor     = "000000"
	defaultPaper         = "A4"
	outputEscaping       = true
	zipClass             = "ZipArchive"
	measurementUnit      = UnitTwip
	xmlCompatibility     = true
	defaultRtl           *bool
	tempDir              string
)

// DefaultFontName returns the process-wide default font.
func DefaultFontName() string { return defaultFontName }

// SetDefaultFontName sets the process-wide default font.
func SetDefaultFontName(name string) { defaultFontName = name }

// DefaultAsianFontName returns the process-wide default East-Asian font.
func DefaultAsianFontName() string { return defaultAsianFontName }

// SetDefaultAsianFontName sets the process-wide default East-Asian font.
func SetDefaultAsianFontName(name string) { defaultAsianFontName = name }

// DefaultFontSize returns the process-wide default size in points.
func DefaultFontSize() float64 { return defaultFontSize }

// SetDefaultFontSize sets the process-wide default size in points.
func SetDefaultFontSize(size float64) { defaultFontSize = size }

// DefaultFontColor returns the process-wide default color (hex, no #).
func DefaultFontColor() string { return defaultFontColor }

// SetDefaultFontColor sets the process-wide default color.
func SetDefaultFontColor(color string) { defaultFontColor = color }

// SetOutputEscapingEnabled toggles XML escaping of text (always recommended).
func SetOutputEscapingEnabled(v bool) { outputEscaping = v }

// IsOutputEscapingEnabled reports whether XML escaping is on.
func IsOutputEscapingEnabled() bool { return outputEscaping }

// DefaultPaper returns the process-wide default paper size name.
func DefaultPaper() string { return defaultPaper }

// SetDefaultPaper sets the process-wide default paper size name.
func SetDefaultPaper(name string) { defaultPaper = name }

// MeasurementUnit returns the process-wide unit (twip, cm, mm, inch, point, pica).
func MeasurementUnit() string { return measurementUnit }

// SetMeasurementUnit sets the process-wide unit.
func SetMeasurementUnit(unit string) { measurementUnit = unit }

// HasCompatibility reports XMLWriter compatibility mode (PHP Settings::hasCompatibility).
func HasCompatibility() bool { return xmlCompatibility }

// SetCompatibility toggles XMLWriter compatibility mode.
func SetCompatibility(v bool) { xmlCompatibility = v }

// SetDefaultRtl sets the process-wide default RTL flag.
func SetDefaultRtl(v *bool) { defaultRtl = v }

// IsDefaultRtl reports the process-wide default RTL flag.
func IsDefaultRtl() *bool { return defaultRtl }

// SetTempDir sets a user-defined temporary directory.
func SetTempDir(dir string) { tempDir = dir }

// TempDir returns the user-defined temporary directory.
func TempDir() string { return tempDir }

// ZipClass returns the zip backend name (always ZipArchive in Go).
func ZipClass() string { return zipClass }

// SetZipClass is accepted for PHP parity; Go always uses archive/zip.
func SetZipClass(name string) { zipClass = name }
