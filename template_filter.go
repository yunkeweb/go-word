package word

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// TemplateFilter transforms a placeholder value. args are the : separated
// arguments from ${var | filter:arg}.
type TemplateFilter func(in any, args ...string) string

var (
	templateFilterMu sync.RWMutex
	templateFilters  = map[string]TemplateFilter{}
)

func init() {
	RegisterTemplateFilter("upper", filterUpper)
	RegisterTemplateFilter("lower", filterLower)
	RegisterTemplateFilter("trim", filterTrim)
	RegisterTemplateFilter("truncate", filterTruncate)
	RegisterTemplateFilter("default", filterDefault)
	RegisterTemplateFilter("formatDate", filterFormatDate)
	RegisterTemplateFilter("formatCurrency", filterFormatCurrency)
}

// RegisterTemplateFilter installs or replaces a named pipe filter.
func RegisterTemplateFilter(name string, fn func(in any, args ...string) string) {
	name = strings.TrimSpace(name)
	if name == "" || fn == nil {
		return
	}
	templateFilterMu.Lock()
	templateFilters[name] = fn
	templateFilterMu.Unlock()
}

func lookupTemplateFilter(name string) TemplateFilter {
	templateFilterMu.RLock()
	fn := templateFilters[name]
	templateFilterMu.RUnlock()
	return fn
}

func stringifyFilterIn(in any) string {
	if in == nil {
		return ""
	}
	switch v := in.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func filterUpper(in any, args ...string) string {
	_ = args
	return strings.ToUpper(stringifyFilterIn(in))
}

func filterLower(in any, args ...string) string {
	_ = args
	return strings.ToLower(stringifyFilterIn(in))
}

func filterTrim(in any, args ...string) string {
	_ = args
	return strings.TrimSpace(stringifyFilterIn(in))
}

func filterTruncate(in any, args ...string) string {
	s := stringifyFilterIn(in)
	if len(args) == 0 || args[0] == "" {
		return s
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func filterDefault(in any, args ...string) string {
	s := stringifyFilterIn(in)
	if strings.TrimSpace(s) != "" {
		return s
	}
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return "N/A"
}

func filterFormatDate(in any, args ...string) string {
	s := stringifyFilterIn(in)
	layout := "2006-01-02"
	if len(args) > 0 && args[0] != "" {
		layout = args[0]
	}
	if s == "" {
		return ""
	}
	if t, ok := in.(time.Time); ok {
		return t.Format(layout)
	}
	for _, l := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"2006-01-02T15:04:05",
		time.RFC1123,
		time.RFC1123Z,
	} {
		t, err := time.Parse(l, s)
		if err == nil {
			return t.Format(layout)
		}
	}
	return s
}

func filterFormatCurrency(in any, args ...string) string {
	s := stringifyFilterIn(in)
	symbol := "¥"
	if len(args) > 0 && args[0] != "" {
		symbol = args[0]
	}
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		if s == "" {
			return symbol + "0.00"
		}
		return symbol + s
	}
	return symbol + strconv.FormatFloat(f, 'f', 2, 64)
}

type pipeCall struct {
	name string
	args []string
}

func decodeMacroXML(s string) string {
	return html.UnescapeString(s)
}

func splitTopLevel(s string, sep rune) []string {
	var parts []string
	var b strings.Builder
	quote := rune(0)
	for _, r := range s {
		if quote != 0 {
			b.WriteRune(r)
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '"' || r == '\'' {
			quote = r
			b.WriteRune(r)
			continue
		}
		if r == sep {
			parts = append(parts, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	parts = append(parts, b.String())
	return parts
}

func splitPipes(inner string) (field string, pipes []pipeCall) {
	inner = strings.TrimSpace(decodeMacroXML(inner))
	parts := splitTopLevel(inner, '|')
	if len(parts) == 0 {
		return "", nil
	}
	field = strings.TrimSpace(parts[0])
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pipes = append(pipes, parsePipeCall(p))
	}
	return field, pipes
}

func parsePipeCall(s string) pipeCall {
	name := s
	arg := ""
	if i := strings.IndexByte(s, ':'); i >= 0 {
		name = strings.TrimSpace(s[:i])
		arg = strings.TrimSpace(s[i+1:])
	}
	var args []string
	if arg != "" {
		args = []string{unquoteArg(arg)}
	}
	return pipeCall{name: name, args: args}
}

func unquoteArg(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func applyFilters(in string, pipes []pipeCall) string {
	cur := any(in)
	for _, p := range pipes {
		fn := lookupTemplateFilter(p.name)
		if fn == nil {
			cur = stringifyFilterIn(cur)
			continue
		}
		cur = fn(cur, p.args...)
	}
	return stringifyFilterIn(cur)
}

func pipeRest(inner string) string {
	inner = strings.TrimSpace(decodeMacroXML(inner))
	i := strings.IndexRune(inner, '|')
	if i < 0 {
		return ""
	}
	return strings.TrimSpace(inner[i+1:])
}

func parseIfExpr(s string) (field, op, value string) {
	s = strings.TrimSpace(decodeMacroXML(s))
	ops := []string{"==", "!=", ">=", "<=", ">", "<"}
	for _, o := range ops {
		if i := indexBareOp(s, o); i >= 0 {
			field = strings.TrimSpace(s[:i])
			value = unquoteArg(strings.TrimSpace(s[i+len(o):]))
			return field, o, value
		}
	}
	return s, "", ""
}

func indexBareOp(s, op string) int {
	quote := byte(0)
	for i := 0; i+len(op) <= len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			continue
		}
		if s[i:i+len(op)] == op {
			if i > 0 && !isExprBoundary(rune(s[i-1])) {
				continue
			}
			if i+len(op) < len(s) && !isExprBoundary(rune(s[i+len(op)])) {
				continue
			}
			return i
		}
	}
	return -1
}

func isExprBoundary(r rune) bool {
	return unicode.IsSpace(r) || r == '"' || r == '\'' || unicode.IsDigit(r) || unicode.IsLetter(r)
}

func compareValues(left, op, right string) bool {
	lf, lErr := strconv.ParseFloat(strings.TrimSpace(left), 64)
	rf, rErr := strconv.ParseFloat(strings.TrimSpace(right), 64)
	if lErr == nil && rErr == nil {
		switch op {
		case "==":
			return lf == rf
		case "!=":
			return lf != rf
		case ">":
			return lf > rf
		case ">=":
			return lf >= rf
		case "<":
			return lf < rf
		case "<=":
			return lf <= rf
		}
		return false
	}
	switch op {
	case "==":
		return left == right
	case "!=":
		return left != right
	case ">":
		return left > right
	case ">=":
		return left >= right
	case "<":
		return left < right
	case "<=":
		return left <= right
	default:
		return false
	}
}

func (t *TemplateProcessor) evalIfName(name string) bool {
	field, op, value := parseIfExpr(name)
	left := ""
	if t.values != nil {
		left = t.values[field]
	}
	if op == "" {
		if t.values == nil {
			return false
		}
		if _, ok := t.values[field]; !ok {
			return false
		}
		return conditionTrue(left)
	}
	return compareValues(left, op, value)
}

func (t *TemplateProcessor) applyPipesFor(field string) {
	if field == "" {
		return
	}
	re := macroRegexp()
	for _, name := range t.xmlParts() {
		t.files[name] = re.ReplaceAllFunc(t.files[name], func(m []byte) []byte {
			inner := unwrapMacro(string(m))
			kind, _ := parseControlMacro(inner)
			if kind != "var" {
				return m
			}
			fname, pipes := splitPipes(inner)
			if fname != field || len(pipes) == 0 {
				return m
			}
			val := ""
			if t.values != nil {
				val = t.values[field]
			}
			out := applyFilters(val, pipes)
			return []byte(t.ReplaceCarriageReturns(xmlEscape(out)))
		})
	}
}

func (t *TemplateProcessor) applyRemainingPipes() {
	re := macroRegexp()
	for _, name := range t.xmlParts() {
		t.files[name] = re.ReplaceAllFunc(t.files[name], func(m []byte) []byte {
			inner := unwrapMacro(string(m))
			kind, _ := parseControlMacro(inner)
			if kind != "var" {
				return m
			}
			field, pipes := splitPipes(inner)
			if len(pipes) == 0 {
				return m
			}
			val := ""
			if t.values != nil {
				val = t.values[field]
			}
			out := applyFilters(val, pipes)
			return []byte(t.ReplaceCarriageReturns(xmlEscape(out)))
		})
	}
}
