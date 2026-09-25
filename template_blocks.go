package word

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BlockData is one clone of a named template block, including nested
// blocks and ${if} conditions scoped to that instance.
type BlockData struct {
	Values map[string]string
	Blocks map[string][]BlockData
	If     map[string]bool
}

type macroTok struct {
	kind, name string
	pos, end   int
}

type ifPair struct {
	name                 string
	ifStart, ifEnd       int
	endifStart, endifEnd int
}

var (
	vMergeRestartRe  = regexp.MustCompile(`<w:vMerge\b[^>]*\bw:val="restart"`)
	vMergeContinueRe = regexp.MustCompile(`<w:vMerge(?:\s*/>|\s*></w:vMerge>|\b[^>]*w:val="continue")`)
	wtTextRe         = regexp.MustCompile(`<w:t\b[^>]*>([^<]*)</w:t>`)
)

func parseControlMacro(key string) (kind, name string) {
	key = strings.TrimSpace(key)
	if key == "endif" || strings.HasPrefix(key, "endif#") {
		return "endif", key
	}
	if strings.HasPrefix(key, "/") {
		return "endblock", strings.TrimPrefix(key, "/")
	}
	if strings.HasPrefix(key, "if ") || strings.HasPrefix(key, "if\t") {
		return "if", strings.TrimSpace(key[2:])
	}
	return "var", key
}

func isControlMacro(key string) bool {
	kind, _ := parseControlMacro(key)
	return kind != "var"
}

func indexMacros(s string, n int) string {
	idx := "#" + strconv.Itoa(n)
	return macroRegexp().ReplaceAllStringFunc(s, func(m string) string {
		raw := unwrapMacro(m)
		kind, ident := parseControlMacro(raw)
		switch kind {
		case "endif":
			return macro(raw + idx)
		case "if":
			return macro("if " + ident + idx)
		case "endblock":
			return macro("/" + ident + idx)
		default:
			return macro(ident + idx)
		}
	})
}

func findBlockBounds(xml, blockName string) (openStart, openEnd, closeStart, closeEnd int, ok bool) {
	open := macro(blockName)
	close := macro("/" + blockName)
	openStart = strings.Index(xml, open)
	if openStart < 0 {
		return
	}
	openEnd = openStart + len(open)
	depth := 1
	pos := openEnd
	for depth > 0 {
		iOpen := indexFrom(xml, open, pos)
		iClose := indexFrom(xml, close, pos)
		if iClose < 0 {
			return 0, 0, 0, 0, false
		}
		if iOpen >= 0 && iOpen < iClose {
			depth++
			pos = iOpen + len(open)
			continue
		}
		depth--
		if depth == 0 {
			return openStart, openEnd, iClose, iClose + len(close), true
		}
		pos = iClose + len(close)
	}
	return
}

func indexFrom(s, sub string, from int) int {
	if from >= len(s) {
		return -1
	}
	i := strings.Index(s[from:], sub)
	if i < 0 {
		return -1
	}
	return from + i
}

func indexedSuffix(blockName string) string {
	i := strings.Index(blockName, "#")
	if i < 0 {
		return ""
	}
	return blockName[i:]
}

// CloneNestedBlock clones ${name}...${/name} once per item, then clones
// nested blocks and applies values / conditions with #n suffixes.
func (t *TemplateProcessor) CloneNestedBlock(blockName string, items []BlockData) error {
	if len(items) == 0 {
		return t.DeleteBlock(blockName)
	}
	if err := t.CloneBlock(blockName, len(items)); err != nil {
		return err
	}
	inherited := indexedSuffix(blockName)
	for i, item := range items {
		suf := inherited + "#" + strconv.Itoa(i+1)
		for cond, keep := range item.If {
			if err := t.SetCondition(cond+suf, keep); err != nil {
				return err
			}
		}
		for nested, nestedItems := range item.Blocks {
			if err := t.CloneNestedBlock(nested+suf, nestedItems); err != nil {
				return err
			}
		}
		for k, v := range item.Values {
			t.SetValue(k+suf, v)
		}
	}
	return nil
}

// SetCondition keeps or clips ${if name}...${endif}.
func (t *TemplateProcessor) SetCondition(name string, keep bool) error {
	name = strings.TrimSpace(unwrapMacro(name))
	if strings.HasPrefix(name, "if ") || strings.HasPrefix(name, "if\t") {
		name = strings.TrimSpace(name[2:])
	}
	found := false
	for _, part := range t.xmlParts() {
		for {
			xml := string(t.files[part])
			p, ok := findIfPair(xml, name)
			if !ok {
				break
			}
			found = true
			t.files[part] = []byte(applyIf(xml, p, keep))
		}
	}
	if !found {
		return fmt.Errorf("word: if block %s not found", name)
	}
	return nil
}

// SetConditions applies many named ${if} blocks.
func (t *TemplateProcessor) SetConditions(conds map[string]bool) error {
	for name, keep := range conds {
		if err := t.SetCondition(name, keep); err != nil {
			return err
		}
	}
	return nil
}

// ApplyConditionsFromValues treats empty/"0"/"false"/"no"/"off" as false.
func (t *TemplateProcessor) ApplyConditionsFromValues(values map[string]string) error {
	for name, v := range values {
		if err := t.SetCondition(name, conditionTrue(v)); err != nil {
			return err
		}
	}
	return nil
}

func conditionTrue(v string) bool {
	s := strings.TrimSpace(strings.ToLower(v))
	switch s {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func (t *TemplateProcessor) applyRemainingIfBlocks() {
	for _, part := range t.xmlParts() {
		for {
			xml := string(t.files[part])
			p, ok := innermostIfPair(xml)
			if !ok {
				break
			}
			t.files[part] = []byte(applyIf(xml, p, false))
		}
	}
}

func scanMacros(xml string) []macroTok {
	locs := macroRegexp().FindAllStringSubmatchIndex(xml, -1)
	out := make([]macroTok, 0, len(locs))
	for _, loc := range locs {
		key := unwrapMacro(xml[loc[0]:loc[1]])
		kind, name := parseControlMacro(key)
		if kind == "var" {
			name = key
		}
		out = append(out, macroTok{kind: kind, name: name, pos: loc[0], end: loc[1]})
	}
	return out
}

func findIfPair(xml, condName string) (ifPair, bool) {
	toks := scanMacros(xml)
	for i, tok := range toks {
		if tok.kind != "if" || tok.name != condName {
			continue
		}
		depth := 1
		for _, t2 := range toks[i+1:] {
			switch t2.kind {
			case "if":
				depth++
			case "endif":
				depth--
				if depth == 0 {
					return ifPair{
						name:       tok.name,
						ifStart:    tok.pos,
						ifEnd:      tok.end,
						endifStart: t2.pos,
						endifEnd:   t2.end,
					}, true
				}
			}
		}
	}
	return ifPair{}, false
}

func innermostIfPair(xml string) (ifPair, bool) {
	toks := scanMacros(xml)
	type frame struct {
		name string
		s, e int
	}
	var stack []frame
	var best ifPair
	bestDepth := -1
	ok := false
	for _, tok := range toks {
		switch tok.kind {
		case "if":
			stack = append(stack, frame{tok.name, tok.pos, tok.end})
		case "endif":
			if len(stack) == 0 {
				continue
			}
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			depth := len(stack) + 1
			if depth >= bestDepth {
				bestDepth = depth
				best = ifPair{name: f.name, ifStart: f.s, ifEnd: f.e, endifStart: tok.pos, endifEnd: tok.end}
				ok = true
			}
		}
	}
	return best, ok
}

func applyIf(xml string, p ifPair, keep bool) string {
	if keep {
		xml = removeMacroAndEmptyNode(xml, p.endifStart, p.endifEnd)
		xml = removeMacroAndEmptyNode(xml, p.ifStart, p.ifEnd)
		return xml
	}
	from, to := clipDeleteRange(xml, p.ifStart, p.endifEnd)
	return xml[:from] + xml[to:]
}

func removeMacroAndEmptyNode(xml string, ms, me int) string {
	xml2 := xml[:ms] + xml[me:]
	ps, pe := tagBounds(xml, ms, "w:p")
	if ps < 0 {
		return xml2
	}
	pe2 := pe - (me - ms)
	if pe2 <= ps || pe2 > len(xml2) {
		return xml2
	}
	if !hasVisibleText(xml2[ps:pe2]) {
		return xml2[:ps] + xml2[pe2:]
	}
	return xml2
}

func hasVisibleText(paraXML string) bool {
	for _, m := range wtTextRe.FindAllStringSubmatch(paraXML, -1) {
		if strings.TrimSpace(m[1]) != "" {
			return true
		}
	}
	return false
}

func clipDeleteRange(xml string, from, to int) (int, int) {
	if to > len(xml) {
		to = len(xml)
	}
	p0s, p0e := tagBounds(xml, from, "w:p")
	_, p1e := tagBounds(xml, to-1, "w:p")
	r0s, _ := tagBounds(xml, from, "w:tr")
	r1s, r1e := tagBounds(xml, to-1, "w:tr")
	if r0s >= 0 && r1s >= 0 && r0s != r1s {
		return r0s, r1e
	}
	if p0s >= 0 && p1e >= 0 {
		p1s, _ := tagBounds(xml, to-1, "w:p")
		if p0s == p1s {
			return from, to
		}
		return p0s, p1e
	}
	_ = p0e
	return from, to
}

func tagBounds(xml string, pos int, local string) (int, int) {
	if pos < 0 || pos >= len(xml) {
		return -1, -1
	}
	start := lastTagStart(xml[:pos], local)
	if start < 0 {
		return -1, -1
	}
	close := "</" + local + ">"
	rel := strings.Index(xml[pos:], close)
	if rel < 0 {
		return -1, -1
	}
	return start, pos + rel + len(close)
}

func extractRowGroupContaining(xml, needle string) (group string, start, end int) {
	idx := strings.Index(xml, needle)
	if idx < 0 {
		return "", -1, -1
	}
	start = lastTagStart(xml[:idx], "w:tr")
	if start < 0 {
		return "", -1, -1
	}
	rel := strings.Index(xml[idx:], "</w:tr>")
	if rel < 0 {
		return "", -1, -1
	}
	end = idx + rel + len("</w:tr>")
	row := xml[start:end]
	if !vMergeRestartRe.MatchString(row) {
		return row, start, end
	}
	tblRel := strings.Index(xml[end:], "</w:tbl>")
	limit := len(xml)
	if tblRel >= 0 {
		limit = end + tblRel
	}
	pos := end
	for {
		ns, ne := nextRowBounds(xml, pos, limit)
		if ns < 0 {
			break
		}
		next := xml[ns:ne]
		if !vMergeContinueRe.MatchString(next) {
			break
		}
		end = ne
		pos = ne
	}
	return xml[start:end], start, end
}

func nextRowBounds(xml string, from, limit int) (int, int) {
	if from >= limit {
		return -1, -1
	}
	start := indexOpenTag(xml[from:limit], "w:tr")
	if start < 0 {
		return -1, -1
	}
	start += from
	rel := strings.Index(xml[start:], "</w:tr>")
	if rel < 0 {
		return -1, -1
	}
	end := start + rel + len("</w:tr>")
	if end > limit {
		return -1, -1
	}
	return start, end
}

func indexOpenTag(xml, local string) int {
	a := strings.Index(xml, "<"+local+" ")
	b := strings.Index(xml, "<"+local+">")
	switch {
	case a < 0:
		return b
	case b < 0:
		return a
	case a < b:
		return a
	default:
		return b
	}
}

func countOpenTags(xml, local string) int {
	n := 0
	rest := xml
	for {
		i := indexOpenTag(rest, local)
		if i < 0 {
			return n
		}
		n++
		rest = rest[i+len(local)+1:]
	}
}
