package word

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
)

var (
	macroOpen  = "${"
	macroClose = "}"
)

// SetMacroOpeningChars sets the placeholder opening delimiter (PHPWord).
func SetMacroOpeningChars(s string) { macroOpen = s }

// SetMacroClosingChars sets the placeholder closing delimiter.
func SetMacroClosingChars(s string) { macroClose = s }

// SetMacroChars sets both placeholder delimiters.
func SetMacroChars(open, close string) {
	macroOpen = open
	macroClose = close
}

func (t *TemplateProcessor) SetMacroOpeningChars(s string) { SetMacroOpeningChars(s) }
func (t *TemplateProcessor) SetMacroClosingChars(s string) { SetMacroClosingChars(s) }
func (t *TemplateProcessor) SetMacroChars(open, close string) {
	SetMacroChars(open, close)
}

// TemplateProcessor fills ${placeholders} in an existing .docx (PHPWord TemplateProcessor).
type TemplateProcessor struct {
	files map[string][]byte
	order []string
}

// NewTemplateProcessor opens a .docx template from disk.
func NewTemplateProcessor(filename string) (*TemplateProcessor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewTemplateProcessorBytes(data)
}

// NewTemplateProcessorBytes opens a .docx template from memory.
func NewTemplateProcessorBytes(data []byte) (*TemplateProcessor, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	tp := &TemplateProcessor{files: map[string][]byte{}}
	for _, f := range zr.File {
		rc, err := f.Open()
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

func macro(name string) string { return macroOpen + name + macroClose }

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
	search = unwrapMacro(search)
	replace = t.ReplaceCarriageReturns(xmlEscape(replace))
	old := []byte(macro(search))
	neu := []byte(replace)
	for _, name := range t.xmlParts() {
		if limit == 0 {
			return
		}
		src := t.files[name]
		if limit < 0 {
			t.files[name] = bytes.ReplaceAll(src, old, neu)
			continue
		}
		n := bytes.Count(src, old)
		if n > limit {
			n = limit
		}
		t.files[name] = bytes.Replace(src, old, neu, n)
		limit -= n
	}
}

// SetValues replaces many placeholders.
func (t *TemplateProcessor) SetValues(values map[string]string) {
	for k, v := range values {
		t.SetValue(k, v)
	}
}

// CloneRow clones the table row that contains ${search} count times,
// renaming macros to ${search#1}, ${search#2}, ...
func (t *TemplateProcessor) CloneRow(search string, count int) error {
	search = unwrapMacro(search)
	needle := macro(search)
	name := t.mainPart()
	xml := string(t.files[name])
	row := extractRowContaining(xml, needle)
	if row == "" {
		return fmt.Errorf("word: row with %s not found", needle)
	}
	var b strings.Builder
	for i := 1; i <= count; i++ {
		cloned := replaceRowMacros(row, i)
		b.WriteString(cloned)
	}
	xml = strings.Replace(xml, row, b.String(), 1)
	t.files[name] = []byte(xml)
	return nil
}

// CloneBlock clones the region between ${block} and ${/block} count times.
func (t *TemplateProcessor) CloneBlock(blockName string, count int) error {
	name := t.mainPart()
	xml := string(t.files[name])
	start := macro(blockName)
	end := macro("/" + blockName)
	i := strings.Index(xml, start)
	j := strings.Index(xml, end)
	if i < 0 || j < 0 || j < i {
		return fmt.Errorf("word: block %s not found", blockName)
	}
	innerStart := i + len(start)
	block := xml[innerStart:j]
	var b strings.Builder
	for n := 1; n <= count; n++ {
		b.WriteString(replaceRowMacros(block, n))
	}
	xml = xml[:i] + b.String() + xml[j+len(end):]
	t.files[name] = []byte(xml)
	return nil
}

// ReplaceBlock replaces the region between ${block} and ${/block}.
func (t *TemplateProcessor) ReplaceBlock(blockName, replacement string) error {
	name := t.mainPart()
	xml := string(t.files[name])
	start := macro(blockName)
	end := macro("/" + blockName)
	i := strings.Index(xml, start)
	j := strings.Index(xml, end)
	if i < 0 || j < 0 || j < i {
		return fmt.Errorf("word: block %s not found", blockName)
	}
	xml = xml[:i] + xmlEscape(replacement) + xml[j+len(end):]
	t.files[name] = []byte(xml)
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
	search = unwrapMacro(search)
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

	rid := t.addImageRel(media, ext)
	drawing := fmt.Sprintf(
		`<w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">`+
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
		idx, xmlEscape(filename), xmlEscape(filename), rid)
	t.replaceAll(macro(search), drawing)
	return nil
}

func (t *TemplateProcessor) addImageRel(mediaPath, ext string) string {
	relsName := "word/_rels/document.xml.rels"
	rels := string(t.files[relsName])
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

var macroInRow = regexp.MustCompile(`\$\{([^}]+)\}`)

func replaceRowMacros(row string, n int) string {
	return macroInRow.ReplaceAllStringFunc(row, func(m string) string {
		name := m[2 : len(m)-1]
		if strings.HasPrefix(name, "/") {
			return m
		}
		return macro(name + "#" + strconv.Itoa(n))
	})
}

func unwrapMacro(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, macroOpen)
	s = strings.TrimSuffix(s, macroClose)
	return s
}

func xmlEscape(s string) string {
	s = common.ControlCharEncode(s)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
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
	var buf bytes.Buffer
	zw := common.NewZipWriter(&buf)
	for _, name := range t.order {
		if err := zw.AddFile(name, t.files[name]); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveAs is an alias for Save (PHPWord).
func (t *TemplateProcessor) SaveAs(filename string) error { return t.Save(filename) }

func macroRegexp() *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(macroOpen) + `(.+?)` + regexp.QuoteMeta(macroClose))
}

// GetVariableCount returns placeholder name → occurrence count (PHPWord getVariableCount).
func (t *TemplateProcessor) GetVariableCount() map[string]int {
	re := macroRegexp()
	out := map[string]int{}
	for _, name := range t.xmlParts() {
		for _, m := range re.FindAllSubmatch(t.files[name], -1) {
			key := string(m[1])
			if strings.HasPrefix(key, "/") {
				continue
			}
			out[key]++
		}
	}
	return out
}

// GetVariables returns unique placeholder names in first-seen order.
func (t *TemplateProcessor) GetVariables() []string {
	re := macroRegexp()
	seen := map[string]struct{}{}
	var out []string
	for _, name := range t.xmlParts() {
		for _, m := range re.FindAllSubmatch(t.files[name], -1) {
			key := string(m[1])
			if strings.HasPrefix(key, "/") {
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
	search = unwrapMacro(search)
	needle := macro(search)
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
	tableXML := xml[tableStart:tableEnd]
	if strings.Count(tableXML, "<w:tr") == 1 {
		t.files[name] = []byte(xml[:tableStart] + xml[tableEnd:])
		return nil
	}
	row := extractRowContaining(xml, needle)
	if row == "" {
		return fmt.Errorf("word: row with %s not found", needle)
	}
	t.files[name] = []byte(strings.Replace(xml, row, "", 1))
	return nil
}

// SetCheckbox toggles a content-control checkbox at ${search}.
func (t *TemplateProcessor) SetCheckbox(search string, checked bool) {
	search = unwrapMacro(search)
	needle := macro(search)
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
	macroName = unwrapMacro(macroName)
	needle := macro(macroName)
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
	xw := common.NewXMLWriter()
	w.writeElement(xw, el, inline)
	return xw.String()
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
