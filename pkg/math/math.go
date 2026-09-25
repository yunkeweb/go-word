package math

// Math is an Office Math container (PHPOffice Math\Math).
type Math struct {
	Group
}

// New returns an empty math expression.
func New() *Math { return &Math{} }
