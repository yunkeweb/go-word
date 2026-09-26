package math

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ParseLaTeX converts a basic TeX math string into an Office Math tree.
// Supported constructs: identifiers, numbers, operators, {groups},
// \frac{a}{b}, \sqrt{x}, x^{n}, x_{n}, and a small set of named symbols.
func ParseLaTeX(src string) (*Math, error) {
	p := &latexParser{s: strings.TrimSpace(src)}
	m := New()
	for _, el := range p.parseExpr(false) {
		m.Add(el)
	}
	if len(m.Elements) == 0 && src != "" {
		m.Add(NewIdentifier(src))
	}
	return m, nil
}

type latexParser struct {
	s string
	i int
}

func (p *latexParser) peek() rune {
	if p.i >= len(p.s) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.s[p.i:])
	return r
}

func (p *latexParser) next() rune {
	if p.i >= len(p.s) {
		return 0
	}
	r, n := utf8.DecodeRuneInString(p.s[p.i:])
	p.i += n
	return r
}

func (p *latexParser) skipSpace() {
	for p.i < len(p.s) {
		r, n := utf8.DecodeRuneInString(p.s[p.i:])
		if !unicode.IsSpace(r) {
			return
		}
		p.i += n
	}
}

func (p *latexParser) parseExpr(stopAtBrace bool) []Element {
	var out []Element
	for {
		p.skipSpace()
		r := p.peek()
		if r == 0 {
			break
		}
		if stopAtBrace && r == '}' {
			break
		}
		if r == ')' || r == ']' {
			break
		}
		if strings.HasPrefix(p.s[p.i:], `\right`) {
			break
		}
		el := p.parseAtom()
		if el == nil {
			break
		}
		out = append(out, el)
	}
	return out
}

func (p *latexParser) parseAtom() Element {
	p.skipSpace()
	r := p.peek()
	if r == 0 {
		return nil
	}
	base := p.parsePrimary()
	if base == nil {
		return nil
	}
	for {
		p.skipSpace()
		switch p.peek() {
		case '^':
			p.next()
			sup := p.parseScript()
			base = NewSuperscript(base, sup)
		case '_':
			p.next()
			sub := p.parseScript()
			base = NewSubscript(base, sub)
		default:
			return base
		}
	}
}

func (p *latexParser) parsePrimary() Element {
	p.skipSpace()
	r := p.peek()
	switch {
	case r == 0:
		return nil
	case r == '{':
		p.next()
		els := p.parseExpr(true)
		if p.peek() == '}' {
			p.next()
		}
		return wrapRow(els)
	case r == '(':
		p.next()
		els := p.parseExpr(false)
		if p.peek() == ')' {
			p.next()
		}
		return NewDelimiter("(", ")", wrapRow(els))
	case r == '\\':
		return p.parseCommand()
	case unicode.IsDigit(r) || r == '.':
		return NewNumeric(p.scanWhile(func(rr rune) bool {
			return unicode.IsDigit(rr) || rr == '.'
		}))
	case unicode.IsLetter(r):
		return NewIdentifier(string(p.next()))
	default:
		p.next()
		return NewOperator(string(r))
	}
}

func (p *latexParser) parseScript() Element {
	p.skipSpace()
	if p.peek() == '{' {
		return p.parsePrimary()
	}
	r := p.peek()
	if r == 0 {
		return NewIdentifier("")
	}
	if r == '\\' {
		return p.parseCommand()
	}
	p.next()
	if unicode.IsDigit(r) {
		return NewNumeric(string(r))
	}
	return NewIdentifier(string(r))
}

func (p *latexParser) parseCommand() Element {
	p.next() // backslash
	name := p.scanWhile(func(r rune) bool {
		return unicode.IsLetter(r)
	})
	if name == "" {
		r := p.next()
		if r == 0 {
			return NewOperator(`\`)
		}
		return NewOperator(string(r))
	}
	switch name {
	case "frac":
		num := p.parseScript()
		den := p.parseScript()
		return NewFraction(num, den)
	case "sqrt":
		return NewRadical(p.parseScript())
	case "left":
		open := p.takeDelim()
		els := p.parseExpr(false)
		p.skipSpace()
		if strings.HasPrefix(p.s[p.i:], `\right`) {
			p.i += len(`\right`)
		}
		close := p.takeDelim()
		return NewDelimiter(open, close, wrapRow(els))
	case "right":
		return NewOperator(p.takeDelim())
	}
	if sym, ok := latexSymbols[name]; ok {
		return NewIdentifier(sym)
	}
	return NewIdentifier(name)
}

func (p *latexParser) takeDelim() string {
	p.skipSpace()
	r := p.peek()
	if r == 0 {
		return ""
	}
	if r == '\\' {
		p.next()
		name := p.scanWhile(unicode.IsLetter)
		if name == "" {
			return string(p.next())
		}
		if name == "lvert" || name == "vert" {
			return "|"
		}
		return name
	}
	p.next()
	if r == '.' {
		return ""
	}
	return string(r)
}

func (p *latexParser) scanWhile(ok func(rune) bool) string {
	start := p.i
	for p.i < len(p.s) {
		r, n := utf8.DecodeRuneInString(p.s[p.i:])
		if !ok(r) {
			break
		}
		p.i += n
	}
	return p.s[start:p.i]
}

func wrapRow(els []Element) Element {
	switch len(els) {
	case 0:
		return NewIdentifier("")
	case 1:
		return els[0]
	default:
		row := NewRow()
		for _, el := range els {
			row.Add(el)
		}
		return row
	}
}

var latexSymbols = map[string]string{
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ",
	"pi": "π", "theta": "θ", "lambda": "λ", "sigma": "σ", "omega": "ω",
	"times": "×", "cdot": "·", "leq": "≤", "geq": "≥", "neq": "≠",
	"pm": "±", "infty": "∞", "sum": "∑", "int": "∫",
	"cdotp": "·", "div": "÷",
}
