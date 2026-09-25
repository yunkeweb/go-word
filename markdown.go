package word

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

var mdLinkRe = regexp.MustCompile(`^\[([^\]]+)\]\(([^)]+)\)`)

// AddMarkdown parses Markdown and appends native GoWord elements to dst.
func AddMarkdown(dst htmlContainer, md string) error {
	md = strings.ReplaceAll(md, "\r\n", "\n")
	md = strings.ReplaceAll(md, "\r", "\n")
	lines := strings.Split(md, "\n")
	codeFont := style.Font{Name: "Courier New", Size: 10}

	var para []string
	inCode := false
	var code []string

	flushPara := func() {
		if len(para) == 0 {
			return
		}
		text := strings.Join(para, " ")
		para = nil
		addMarkdownParagraph(dst, text)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				if len(code) == 0 {
					dst.AddText("", codeFont)
				}
				for _, cl := range code {
					dst.AddText(cl, codeFont)
				}
				code = nil
				inCode = false
			} else {
				flushPara()
				inCode = true
			}
			continue
		}
		if inCode {
			code = append(code, line)
			continue
		}
		if trimmed == "" {
			flushPara()
			continue
		}
		if isMarkdownHR(trimmed) {
			flushPara()
			dst.AddLine(style.Line{Weight: 1, Color: "000000", Width: 4500})
			continue
		}
		if depth, title, ok := markdownHeading(trimmed); ok {
			flushPara()
			dst.AddTitle(title, depth)
			continue
		}
		if item, depth, ordered, ok := markdownList(line); ok {
			flushPara()
			lt := style.ListTypeBullet
			if ordered {
				lt = style.ListTypeNumber
			}
			dst.AddListItem(markdownPlain(item), depth, nil, nil, lt)
			continue
		}
		para = append(para, trimmed)
	}
	if inCode {
		for _, cl := range code {
			dst.AddText(cl, codeFont)
		}
	}
	flushPara()
	return nil
}

// AddMarkdown parses Markdown into the last (or a new) section.
func (d *Document) AddMarkdown(mdText string) error {
	return AddMarkdown(d.lastOrNewSection(), mdText)
}

func addMarkdownParagraph(dst htmlContainer, text string) {
	tr := dst.AddTextRun()
	addMarkdownInline(tr, text)
}

func addMarkdownInline(tr *element.TextRun, s string) {
	i := 0
	for i < len(s) {
		if n, inner, ok := takeDelim(s, i, "***"); ok {
			tr.AddText(inner, style.Font{Bold: true, Italic: true})
			i += n
			continue
		}
		if n, inner, ok := takeDelim(s, i, "**"); ok {
			tr.AddText(inner, style.Font{Bold: true})
			i += n
			continue
		}
		if n, inner, ok := takeDelim(s, i, "__"); ok {
			tr.AddText(inner, style.Font{Bold: true})
			i += n
			continue
		}
		if n, inner, ok := takeDelim(s, i, "++"); ok {
			tr.AddText(inner, style.Font{Underline: style.UnderlineSingle})
			i += n
			continue
		}
		if strings.HasPrefix(s[i:], "<u>") {
			if j := strings.Index(s[i+3:], "</u>"); j >= 0 {
				tr.AddText(s[i+3:i+3+j], style.Font{Underline: style.UnderlineSingle})
				i += 3 + j + 4
				continue
			}
		}
		if n, inner, ok := takeDelim(s, i, "*"); ok {
			tr.AddText(inner, style.Font{Italic: true})
			i += n
			continue
		}
		if n, inner, ok := takeDelim(s, i, "_"); ok {
			tr.AddText(inner, style.Font{Italic: true})
			i += n
			continue
		}
		if n, inner, ok := takeDelim(s, i, "`"); ok {
			tr.AddText(inner, style.Font{Name: "Courier New"})
			i += n
			continue
		}
		if s[i] == '[' {
			if m := mdLinkRe.FindStringSubmatch(s[i:]); len(m) == 3 {
				tr.AddLink(m[2], m[1])
				i += len(m[0])
				continue
			}
		}
		j := i + 1
		for j < len(s) {
			r := s[j]
			if r == '*' || r == '_' || r == '`' || r == '[' || r == '+' || r == '<' {
				break
			}
			j++
		}
		tr.AddText(s[i:j])
		i = j
	}
}

func takeDelim(s string, i int, delim string) (consumed int, inner string, ok bool) {
	if !strings.HasPrefix(s[i:], delim) {
		return 0, "", false
	}
	rest := s[i+len(delim):]
	j := strings.Index(rest, delim)
	if j < 0 {
		return 0, "", false
	}
	return len(delim)*2 + j, rest[:j], true
}

func markdownHeading(s string) (int, string, bool) {
	n := 0
	for n < len(s) && s[n] == '#' {
		n++
	}
	if n == 0 || n > 6 {
		return 0, "", false
	}
	if n == len(s) {
		return n, "", true
	}
	if s[n] != ' ' {
		return 0, "", false
	}
	return n, strings.TrimSpace(s[n+1:]), true
}

func markdownList(line string) (text string, depth int, ordered bool, ok bool) {
	i := 0
	spaces := 0
	for i < len(line) {
		switch line[i] {
		case ' ':
			spaces++
			i++
		case '\t':
			spaces += 2
			i++
		default:
			goto body
		}
	}
body:
	depth = spaces / 2
	rest := line[i:]
	if len(rest) >= 2 && (rest[0] == '-' || rest[0] == '*' || rest[0] == '+') && rest[1] == ' ' {
		return strings.TrimSpace(rest[2:]), depth, false, true
	}
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	if j > 0 && j+1 < len(rest) && rest[j] == '.' && rest[j+1] == ' ' {
		return strings.TrimSpace(rest[j+2:]), depth, true, true
	}
	return "", 0, false, false
}

func isMarkdownHR(s string) bool {
	if len(s) < 3 {
		return false
	}
	r := rune(s[0])
	if r != '-' && r != '*' && r != '_' {
		return false
	}
	n := 0
	for _, c := range s {
		if c == r {
			n++
			continue
		}
		if unicode.IsSpace(c) {
			continue
		}
		return false
	}
	return n >= 3
}

func markdownPlain(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "++", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "<u>", "")
	s = strings.ReplaceAll(s, "</u>", "")
	return s
}
