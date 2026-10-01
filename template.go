package word

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
)

var (
	macroDefaultsMu sync.RWMutex
	macroOpen       = "${"
	macroClose      = "}"
)

// SetMacroOpeningChars sets the placeholder opening delimiter (PHPWord).
func SetMacroOpeningChars(s string) { macroDefaultsMu.Lock(); macroOpen = s; macroDefaultsMu.Unlock() }

// SetMacroClosingChars sets the placeholder closing delimiter.
func SetMacroClosingChars(s string) { macroDefaultsMu.Lock(); macroClose = s; macroDefaultsMu.Unlock() }

// SetMacroChars sets both placeholder delimiters.
func SetMacroChars(open, close string) {
	macroDefaultsMu.Lock()
	macroOpen = open
	macroClose = close
	macroDefaultsMu.Unlock()
}

func (t *TemplateProcessor) SetMacroOpeningChars(s string) {
	t.initMacros()
	t.SetMacroChars(s, t.macroClose)
}
func (t *TemplateProcessor) SetMacroClosingChars(s string) {
	t.initMacros()
	t.SetMacroChars(t.macroOpen, s)
}
func (t *TemplateProcessor) SetMacroChars(open, close string) {
	t.macroOpen = open
	t.macroClose = close
	t.macroPattern = regexp.MustCompile(regexp.QuoteMeta(open) + `(.+?)` + regexp.QuoteMeta(close))
}

// TemplateProcessor fills ${placeholders} in an existing .docx (PHPWord TemplateProcessor).
type TemplateProcessor struct {
	files        map[string][]byte
	order        []string
	values       map[string]string
	macroOpen    string
	macroClose   string
	macroPattern *regexp.Regexp
}

// NewTemplateProcessor opens a .docx template from disk.
func NewTemplateProcessor(filename string) (*TemplateProcessor, error) {
	return NewTemplateProcessorWithOptions(filename, ReadOptions{})
}

// NewTemplateProcessorWithOptions opens a disk template with ZIP read budgets.
func NewTemplateProcessorWithOptions(filename string, opts ReadOptions) (*TemplateProcessor, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	zr, err := common.OpenZipFileWithLimit(filename, opts.MaxArchiveSize)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return loadTemplate(zr, opts)
}

// NewTemplateProcessorBytes opens a .docx template from memory.
func NewTemplateProcessorBytes(data []byte) (*TemplateProcessor, error) {
	return NewTemplateProcessorBytesWithOptions(data, ReadOptions{})
}

// NewTemplateProcessorBytesWithOptions opens a template with ZIP read budgets.
func NewTemplateProcessorBytesWithOptions(data []byte, opts ReadOptions) (*TemplateProcessor, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if err := validateZipLimits(nil, int64(len(data)), opts); err != nil {
		return nil, err
	}
	zr, err := common.OpenZipBytes(data)
	if err != nil {
		return nil, err
	}
	return loadTemplate(zr, opts)
}

func loadTemplate(zr *common.ZipReader, opts ReadOptions) (*TemplateProcessor, error) {
	if err := validateZipLimits(zr.Files(), -1, opts); err != nil {
		return nil, err
	}
	tp := &TemplateProcessor{files: map[string][]byte{}, values: map[string]string{}}
	tp.initMacros()
	for _, f := range zr.Files() {
		rc, err := common.OpenZipEntry(f, opts.partLimit())
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		tp.files[f.Name] = b
		tp.order = append(tp.order, f.Name)
	}
	return tp, nil
}

func (t *TemplateProcessor) mainPart() string { return "word/document.xml" }

func (t *TemplateProcessor) xmlParts() []string {
	var out []string
	for _, name := range t.order {
		if !strings.HasSuffix(name, ".xml") {
			continue
		}
		if strings.HasPrefix(name, "word/") || name == t.mainPart() {
			out = append(out, name)
		}
	}
	return out
}

func (t *TemplateProcessor) macro(name string) string {
	t.initMacros()
	return t.macroOpen + name + t.macroClose
}

func (t *TemplateProcessor) unwrapMacro(s string) string {
	s = strings.TrimSpace(s)
	t.initMacros()
	s = strings.TrimPrefix(s, t.macroOpen)
	s = strings.TrimSuffix(s, t.macroClose)
	return decodeMacroXML(s)
}

func (t *TemplateProcessor) macroRegexp() *regexp.Regexp {
	t.initMacros()
	return t.macroPattern
}

func (t *TemplateProcessor) initMacros() {
	if t.macroPattern != nil {
		return
	}
	macroDefaultsMu.RLock()
	open, close := macroOpen, macroClose
	macroDefaultsMu.RUnlock()
	t.SetMacroChars(open, close)
}

func (t *TemplateProcessor) replaceAll(old, new string) {
	for _, name := range t.xmlParts() {
		t.files[name] = bytes.ReplaceAll(t.files[name], []byte(old), []byte(new))
	}
}

// SetValue replaces ${search} with replace across document parts.
func (t *TemplateProcessor) SetValue(search, replace string) {
	t.SetValueLimit(search, replace, -1)
}

// SetValueLimit replaces at most limit occurrences (-1 = all).
func (t *TemplateProcessor) SetValueLimit(search, replace string, limit int) {
	search = t.unwrapMacro(search)
	if t.values == nil {
		t.values = map[string]string{}
	}
	t.values[search] = replace
	t.applyPipesFor(search)
	replace = t.ReplaceCarriageReturns(xmlEscape(replace))
	old := []byte(t.macro(search))
	neu := []byte(replace)
	for _, name := range t.xmlParts() {
		if limit == 0 {
			return
		}
		src := t.files[name]
		if limit < 0 {
			t.files[name] = replaceAcrossRuns(src, old, neu, -1)
			continue
		}
		var n int
		t.files[name], n = replaceAcrossRunsCount(src, old, neu, limit)
		limit -= n
	}
}

var templateParagraphRe = regexp.MustCompile(`(?s)<w:p\b[^>]*>.*?</w:p>`)
var templateTextRe = regexp.MustCompile(`(?s)(<w:t\b[^>]*>)(.*?)(</w:t>)`)

func replaceAcrossRuns(src, old, neu []byte, limit int) []byte {
	out, _ := replaceAcrossRunsCount(src, old, neu, limit)
	return out
}

func replaceAcrossRunsCount(src, old, neu []byte, limit int) ([]byte, int) {
	if len(old) == 0 {
		return src, 0
	}
	count := 0
	out := templateParagraphRe.ReplaceAllFunc(src, func(par []byte) []byte {
		if limit >= 0 && count >= limit {
			return par
		}
		matches := templateTextRe.FindAllSubmatchIndex(par, -1)
		if len(matches) == 0 {
			return par
		}
		contents := make([][]byte, len(matches))
		for i, m := range matches {
			contents[i] = append([]byte(nil), par[m[4]:m[5]]...)
		}
		for {
			if limit >= 0 && count >= limit {
				break
			}
			joined := bytes.Join(contents, nil)
			idx := bytes.Index(joined, old)
			if idx < 0 {
				break
			}
			end := idx + len(old)
			startNode, endNode := -1, -1
			startOff, endOff := 0, 0
			off := 0
			for i, c := range contents {
				next := off + len(c)
				if startNode < 0 && idx < next {
					startNode, startOff = i, idx-off
				}
				if end <= next {
					endNode, endOff = i, end-off
					break
				}
				off = next
			}
			if startNode < 0 || endNode < 0 {
				break
			}
			if startNode == endNode {
				contents[startNode] = append(append(append([]byte{}, contents[startNode][:startOff]...), neu...), contents[startNode][endOff:]...)
			} else {
				contents[startNode] = append(append(append([]byte{}, contents[startNode][:startOff]...), neu...), nil...)
				for i := startNode + 1; i < endNode; i++ {
					contents[i] = nil
				}
				contents[endNode] = append([]byte{}, contents[endNode][endOff:]...)
			}
			count++
		}
		var b bytes.Buffer
		prev := 0
		for i, m := range matches {
			b.Write(par[prev:m[4]])
			b.Write(contents[i])
			prev = m[5]
		}
		b.Write(par[prev:])
		return b.Bytes()
	})
	return out, count
}

// SetValues replaces many placeholders.
func (t *TemplateProcessor) SetValues(values map[string]string) {
	for k, v := range values {
		t.SetValue(k, v)
	}
}

// CloneRow clones the table row that contains ${search} count times,
// renaming macros to ${search#1}, ${search#2}, ... Vertically merged
// (w:vMerge restart/continue) rows are cloned as a single group.
func (t *TemplateProcessor) CloneRow(search string, count int) error {
	search = t.unwrapMacro(search)
	needle := t.macro(search)
	name := t.mainPart()
	xml := string(t.files[name])
	row, start, end := extractRowGroupContaining(xml, needle)
	if row == "" {
		return fmt.Errorf("word: row with %s not found", needle)
	}
	if count <= 0 {
		t.files[name] = []byte(xml[:start] + xml[end:])
		return nil
	}
	var b strings.Builder
	for i := 1; i <= count; i++ {
		b.WriteString(t.indexMacros(row, i))
	}
	t.files[name] = []byte(xml[:start] + b.String() + xml[end:])
	return nil
}

// CloneBlock clones the region between ${block} and ${/block} count times.
// Nested ${inner}...${/inner} markers are indexed (${inner#1}) so they can
// be cloned per instance. Matching uses nesting depth for same-named blocks.
func (t *TemplateProcessor) CloneBlock(blockName string, count int) error {
	name := t.mainPart()
	xml := string(t.files[name])
	openStart, openEnd, closeStart, closeEnd, ok := t.findBlockBounds(xml, blockName)
	if !ok {
		return fmt.Errorf("word: block %s not found", blockName)
	}
	if count <= 0 {
		t.files[name] = []byte(xml[:openStart] + xml[closeEnd:])
		return nil
	}
	block := xml[openEnd:closeStart]
	var b strings.Builder
	for n := 1; n <= count; n++ {
		b.WriteString(t.indexMacros(block, n))
	}
	t.files[name] = []byte(xml[:openStart] + b.String() + xml[closeEnd:])
	return nil
}

// ReplaceBlock replaces the region between ${block} and ${/block}.
func (t *TemplateProcessor) ReplaceBlock(blockName, replacement string) error {
	name := t.mainPart()
	xml := string(t.files[name])
	openStart, _, _, closeEnd, ok := t.findBlockBounds(xml, blockName)
	if !ok {
		return fmt.Errorf("word: block %s not found", blockName)
	}
	t.files[name] = []byte(xml[:openStart] + xmlEscape(replacement) + xml[closeEnd:])
	return nil
}

// DeleteBlock removes the region between ${block} and ${/block}.
func (t *TemplateProcessor) DeleteBlock(blockName string) error {
	return t.ReplaceBlock(blockName, "")
}

// SetImageValue replaces a placeholder with a drawing that references a new image part.
func (t *TemplateProcessor) SetImageValue(search, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return t.SetImageValueBytes(search, path, data)
}

// SetImageValueBytes embeds image bytes at ${search}.
func (t *TemplateProcessor) SetImageValueBytes(search, filename string, data []byte) error {
	search = t.unwrapMacro(search)
	ext := "png"
	if i := strings.LastIndexByte(filename, '.'); i >= 0 {
		ext = strings.ToLower(filename[i+1:])
		if ext == "jpg" {
			ext = "jpeg"
		}
	}
	idx := 1
	for k := range t.files {
		if strings.HasPrefix(k, "word/media/image") {
			idx++
		}
	}
	media := fmt.Sprintf("word/media/imageT%d.%s", idx, ext)
	t.files[media] = data
	t.order = append(t.order, media)
	needle := t.macro(search)
	for _, part := range t.xmlParts() {
		if !bytes.Contains(t.files[part], []byte(needle)) {
			continue
		}
		rid := t.addImageRelForPart(part, media, ext)
		drawingID := 1
		dec := xml.NewDecoder(bytes.NewReader(t.files[part]))
		for {
			tok, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "docPr" {
				if n := atoi(attr(se, "id")); n >= drawingID {
					drawingID = n + 1
				}
			}
		}
		makeDrawing := func() string {
			drawing := fmt.Sprintf(
				`<w:drawing xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><wp:inline distT="0" distB="0" distL="0" distR="0">`+
					`<wp:extent cx="990600" cy="792480"/>`+
					`<wp:docPr id="%d" name="%s"/>`+
					`<a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`+
					`<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
					`<pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
					`<pic:nvPicPr><pic:cNvPr id="0" name="%s"/><pic:cNvPicPr/></pic:nvPicPr>`+
					`<pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`+
					`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="990600" cy="792480"/></a:xfrm>`+
					`<a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`+
					`</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing>`,
				drawingID, xmlEscape(filename), xmlEscape(filename), rid)
			drawingID++
			return drawing
		}
		t.files[part] = wtTextRe.ReplaceAllFunc(t.files[part], func(tag []byte) []byte {
			if !bytes.Contains(tag, []byte(needle)) {
				return tag
			}
			start := bytes.IndexByte(tag, '>') + 1
			pieces := strings.Split(string(tag[start:len(tag)-len("</w:t>")]), needle)
			var b strings.Builder
			for i, text := range pieces {
				if i > 0 {
					b.WriteString(makeDrawing())
				}
				b.WriteString(`<w:t xml:space="preserve">`)
				b.WriteString(text)
				b.WriteString(`</w:t>`)
			}
			return []byte(b.String())
		})
	}
	return nil
}

func (t *TemplateProcessor) addImageRel(mediaPath, ext string) string {
	return t.addImageRelForPart("word/document.xml", mediaPath, ext)
}

func relationshipPartName(part string) string {
	part = strings.ReplaceAll(part, "\\", "/")
	i := strings.LastIndexByte(part, '/')
	if i < 0 {
		return "_rels/" + part + ".rels"
	}
	dir, base := part[:i], part[i+1:]
	return dir + "/_rels/" + base + ".rels"
}

func (t *TemplateProcessor) addImageRelForPart(part, mediaPath, ext string) string {
	relsName := relationshipPartName(part)
	rels := string(t.files[relsName])
	if rels == "" {
		rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`
		t.files[relsName] = []byte(rels)
		t.order = append(t.order, relsName)
	}
	id := nextRelID(rels)
	target := strings.TrimPrefix(mediaPath, "word/")
	entry := fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"/>`,
		id, "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image", target)
	rels = strings.Replace(rels, "</Relationships>", entry+"</Relationships>", 1)
	t.files[relsName] = []byte(rels)

	ctName := "[Content_Types].xml"
	ct := string(t.files[ctName])
	def := fmt.Sprintf(`<Default Extension="%s" ContentType="%s"/>`, ext, mimeForExt(ext))
	if !strings.Contains(ct, `Extension="`+ext+`"`) {
		ct = strings.Replace(ct, "</Types>", def+"</Types>", 1)
		t.files[ctName] = []byte(ct)
	}
	return id
}

func nextRelID(rels string) string {
	re := regexp.MustCompile(`Id="rId(\d+)"`)
	max := 0
	for _, m := range re.FindAllStringSubmatch(rels, -1) {
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
	}
	return "rId" + strconv.Itoa(max+1)
}

func extractRowContaining(xml, needle string) string {
	idx := strings.Index(xml, needle)
	if idx < 0 {
		return ""
	}
	start := lastTagStart(xml[:idx], "w:tr")
	if start < 0 {
		return ""
	}
	end := strings.Index(xml[idx:], "</w:tr>")
	if end < 0 {
		return ""
	}
	return xml[start : idx+end+len("</w:tr>")]
}

// lastTagStart finds the last start tag named local (e.g. w:tbl) so that
// w:tblPr / w:tblGrid / w:trPr are not mistaken for the parent element.
func lastTagStart(xml, local string) int {
	s1 := strings.LastIndex(xml, "<"+local+" ")
	s2 := strings.LastIndex(xml, "<"+local+">")
	if s1 > s2 {
		return s1
	}
	return s2
}

func replaceRowMacros(row string, n int) string { return new(TemplateProcessor).indexMacros(row, n) }

func xmlEscape(s string) string {
	return common.EscapeXMLText(common.ControlCharEncode(s))
}

// Save writes the filled template to filename.
func (t *TemplateProcessor) Save(filename string) error {
	b, err := t.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0o644)
}

// Bytes returns the filled template as a .docx package.
func (t *TemplateProcessor) Bytes() ([]byte, error) {
	t.applyRemainingPipes()
	t.applyRemainingIfBlocks()
	buf := common.GetBuffer()
	defer common.PutBuffer(buf)
	zw := common.NewZipWriter(buf)
	for _, name := range t.order {
		if err := zw.AddFile(name, t.files[name]); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return common.CloneBytes(buf.Bytes()), nil
}

// SaveAs is an alias for Save (PHPWord).
func (t *TemplateProcessor) SaveAs(filename string) error { return t.Save(filename) }

// GetVariableCount returns placeholder name → occurrence count (PHPWord getVariableCount).
func (t *TemplateProcessor) GetVariableCount() map[string]int {
	re := t.macroRegexp()
	out := map[string]int{}
	for _, name := range t.xmlParts() {
		for _, m := range re.FindAllSubmatch(t.files[name], -1) {
			key := string(m[1])
			if isControlMacro(key) {
				continue
			}
			out[key]++
		}
	}
	return out
}

// GetVariables returns unique placeholder names in first-seen order.
func (t *TemplateProcessor) GetVariables() []string {
	re := t.macroRegexp()
	seen := map[string]struct{}{}
	var out []string
	for _, name := range t.xmlParts() {
		for _, m := range re.FindAllSubmatch(t.files[name], -1) {
			key := string(m[1])
			if isControlMacro(key) {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out
}

// CloneRowAndSetValues clones a row and fills ${name#n} from values.
func (t *TemplateProcessor) CloneRowAndSetValues(search string, values []map[string]string) error {
	if err := t.CloneRow(search, len(values)); err != nil {
		return err
	}
	for i, row := range values {
		n := i + 1
		for k, v := range row {
			t.SetValue(k+"#"+strconv.Itoa(n), v)
		}
	}
	return nil
}

// DeleteRow removes the table row that contains ${search}.
func (t *TemplateProcessor) DeleteRow(search string) error {
	search = t.unwrapMacro(search)
	needle := t.macro(search)
	name := t.mainPart()
	xml := string(t.files[name])
	idx := strings.Index(xml, needle)
	if idx < 0 {
		return fmt.Errorf("word: row with %s not found", needle)
	}
	tableStart := lastTagStart(xml[:idx], "w:tbl")
	tableEndRel := strings.Index(xml[idx:], "</w:tbl>")
	if tableStart < 0 || tableEndRel < 0 {
		return fmt.Errorf("word: table for %s not found", needle)
	}
	tableEnd := idx + tableEndRel + len("</w:tbl>")
	row, rowStart, rowEnd := extractRowGroupContaining(xml, needle)
	if row == "" {
		return fmt.Errorf("word: row with %s not found", needle)
	}
	remaining := xml[tableStart:rowStart] + xml[rowEnd:tableEnd]
	if countOpenTags(remaining, "w:tr") == 0 {
		t.files[name] = []byte(xml[:tableStart] + xml[tableEnd:])
		return nil
	}
	t.files[name] = []byte(xml[:rowStart] + xml[rowEnd:])
	return nil
}

// SetCheckbox toggles a content-control checkbox at ${search}.
func (t *TemplateProcessor) SetCheckbox(search string, checked bool) {
	search = t.unwrapMacro(search)
	needle := t.macro(search)
	name := t.mainPart()
	xml := string(t.files[name])
	start, end := findXMLBlock(xml, needle, "w:sdt")
	if start < 0 {
		return
	}
	block := xml[start:end]
	val := "0"
	text := "☐"
	if checked {
		val = "1"
		text = "☒"
	}
	reVal := regexp.MustCompile(`(<w14:checked w14:val=").*?("/>)`)
	block = reVal.ReplaceAllString(block, `${1}`+val+`${2}`)
	reText := regexp.MustCompile(`(<w:t>).*?(</w:t>)`)
	block = reText.ReplaceAllString(block, `${1}`+text+`${2}`)
	t.files[name] = []byte(xml[:start] + block + xml[end:])
}

// SetUpdateFields sets w:updateFields in word/settings.xml.
func (t *TemplateProcessor) SetUpdateFields(update bool) {
	name := "word/settings.xml"
	xml := string(t.files[name])
	if xml == "" {
		return
	}
	val := "false"
	if update {
		val = "true"
	}
	re := regexp.MustCompile(`<w:updateFields w:val="[^"]*"/>`)
	tag := `<w:updateFields w:val="` + val + `"/>`
	if re.MatchString(xml) {
		t.files[name] = []byte(re.ReplaceAllString(xml, tag))
		return
	}
	t.files[name] = []byte(strings.Replace(xml, "</w:settings>", tag+"</w:settings>", 1))
}

// ReplaceCarriageReturns turns newlines into Word break runs (PHPWord).
func (t *TemplateProcessor) ReplaceCarriageReturns(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", "</w:t><w:br/><w:t>")
}

// ReplaceXmlBlock replaces the XML element of blockType that contains ${macro}.
func (t *TemplateProcessor) ReplaceXmlBlock(macroName, block, blockType string) error {
	macroName = t.unwrapMacro(macroName)
	needle := t.macro(macroName)
	name := t.mainPart()
	xml := string(t.files[name])
	start, end := findXMLBlock(xml, needle, blockType)
	if start < 0 {
		return fmt.Errorf("word: block %s for %s not found", blockType, needle)
	}
	t.files[name] = []byte(xml[:start] + block + xml[end:])
	return nil
}

// SetComplexValue replaces the w:r containing ${search} with the rendered element.
func (t *TemplateProcessor) SetComplexValue(search string, el element.Element) error {
	xml := renderElementXML(el, true)
	return t.ReplaceXmlBlock(search, xml, "w:r")
}

// SetComplexBlock replaces the w:p containing ${search} with the rendered element.
func (t *TemplateProcessor) SetComplexBlock(search string, el element.Element) error {
	xml := renderElementXML(el, false)
	return t.ReplaceXmlBlock(search, xml, "w:p")
}

// CloneBlockAndSetValues clones ${block}...${/block} once per values row (PHPWord).
func (t *TemplateProcessor) CloneBlockAndSetValues(blockName string, values []map[string]string) error {
	if err := t.CloneBlock(blockName, len(values)); err != nil {
		return err
	}
	for i, row := range values {
		n := i + 1
		mapped := map[string]string{}
		for k, v := range row {
			mapped[k+"#"+strconv.Itoa(n)] = v
			mapped[k] = v
		}
		t.SetValues(mapped)
	}
	return nil
}

// SetChart replaces ${search} with a chart drawing and chart XML part (PHPWord).
func (t *TemplateProcessor) SetChart(search string, ch *element.Chart) error {
	if ch == nil {
		return fmt.Errorf("word: nil chart")
	}
	relsName := "word/_rels/document.xml.rels"
	rels := string(t.files[relsName])
	rid := nextRelID(rels)
	idx := 1
	for k := range t.files {
		if strings.HasPrefix(k, "word/charts/chart") {
			idx++
		}
	}
	part := fmt.Sprintf("word/charts/chart%d.xml", idx)
	t.files[part] = chartPartXML(ch)
	t.order = append(t.order, part)
	entry := fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"/>`,
		rid, ooxml.NSOfficeRelChart, strings.TrimPrefix(part, "word/"))
	rels = strings.Replace(rels, "</Relationships>", entry+"</Relationships>", 1)
	t.files[relsName] = []byte(rels)

	ctName := "[Content_Types].xml"
	ct := string(t.files[ctName])
	override := fmt.Sprintf(`<Override PartName="/%s" ContentType="%s"/>`, part, ooxml.CTChart)
	if !strings.Contains(ct, part) {
		ct = strings.Replace(ct, "</Types>", override+"</Types>", 1)
		t.files[ctName] = []byte(ct)
	}
	return t.ReplaceXmlBlock(search, chartDrawingXML(rid, ch, true), "w:p")
}

func renderElementXML(el element.Element, inline bool) string {
	w := newWord2007Writer(New())
	xw := common.GetXMLWriter()
	w.writeElement(xw, el, inline)
	s := xw.String()
	common.PutXMLWriter(xw)
	return s
}

func findXMLBlock(xml, needle, blockType string) (start, end int) {
	idx := strings.Index(xml, needle)
	if idx < 0 {
		return -1, -1
	}
	close := "</" + blockType + ">"
	start = lastTagStart(xml[:idx], blockType)
	if start < 0 {
		return -1, -1
	}
	rel := strings.Index(xml[idx:], close)
	if rel < 0 {
		return -1, -1
	}
	return start, idx + rel + len(close)
}
