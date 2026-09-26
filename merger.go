package word

import (
	"fmt"
	"strconv"

	"github.com/yunkeweb/go-word/element"
)

// MergeOptions controls how AppendDocument remaps colliding identifiers.
type MergeOptions struct {
	StylePrefix    string
	BookmarkPrefix string
	SectionBreak   string
}

// AppendDocument clones src onto d, remapping colliding style names, bookmark
// names, and media so the merged package writes unique rIds and part paths.
func (d *Document) AppendDocument(src *Document, opts MergeOptions) error {
	if src == nil {
		return fmt.Errorf("word: nil source document")
	}
	if opts.StylePrefix == "" {
		opts.StylePrefix = "src_"
	}
	if opts.BookmarkPrefix == "" {
		opts.BookmarkPrefix = "src_"
	}
	if opts.SectionBreak == "" {
		opts.SectionBreak = "nextPage"
	}

	styleMap := d.mergeStyles(src, opts.StylePrefix)
	usedBM := map[string]bool{}
	for _, b := range d.GetBookmarks() {
		usedBM[b.Name] = true
	}
	for _, t := range d.GetTitles() {
		if t.BookmarkName != "" {
			usedBM[t.BookmarkName] = true
		}
	}

	first := true
	for _, sec := range src.Sections() {
		cloned := element.CloneSection(sec)
		if cloned == nil {
			continue
		}
		if first {
			cloned.Style.BreakType = opts.SectionBreak
			first = false
		}
		d.sections = append(d.sections, cloned)
		cloned.SectionID = len(d.sections)
		for _, h := range cloned.Headers {
			h.SectionID = cloned.SectionID
		}
		for _, f := range cloned.Footers {
			f.SectionID = cloned.SectionID
		}
		remapClonedTree(cloned, styleMap, usedBM, opts.BookmarkPrefix)
	}
	return nil
}

func (d *Document) mergeStyles(src *Document, prefix string) map[string]string {
	out := map[string]string{}
	if src == nil {
		return out
	}
	have := map[string]bool{}
	for _, s := range d.styles {
		have[s.Name] = true
	}
	for _, s := range src.styles {
		name := s.Name
		if have[name] {
			name = prefix + s.Name
			n := 2
			for have[name] {
				name = prefix + s.Name + strconv.Itoa(n)
				n++
			}
		}
		ns := s
		ns.Name = name
		d.styles = append(d.styles, ns)
		have[name] = true
		if name != s.Name {
			out[s.Name] = name
		}
	}
	return out
}

func remapClonedTree(root element.Element, styleMap map[string]string, usedBM map[string]bool, bmPrefix string) {
	bmMap := map[string]string{}
	walkElement(root, func(el element.Element) {
		switch v := el.(type) {
		case *element.Bookmark:
			old := v.Name
			v.Name = uniqueName(v.Name, usedBM, bmPrefix)
			if old != "" && old != v.Name {
				bmMap[old] = v.Name
			}
		case *element.Title:
			if v.BookmarkName != "" {
				old := v.BookmarkName
				v.BookmarkName = uniqueName(v.BookmarkName, usedBM, bmPrefix)
				if old != v.BookmarkName {
					bmMap[old] = v.BookmarkName
				}
			}
		}
	})
	walkElement(root, func(el element.Element) {
		switch v := el.(type) {
		case *element.Text:
			v.FontStyle = remapStyleRef(v.FontStyle, styleMap)
			v.ParagraphStyle = remapStyleRef(v.ParagraphStyle, styleMap)
		case *element.TextRun:
			v.ParagraphStyle = remapStyleRef(v.ParagraphStyle, styleMap)
		case *element.Link:
			v.FontStyle = remapStyleRef(v.FontStyle, styleMap)
			v.ParagraphStyle = remapStyleRef(v.ParagraphStyle, styleMap)
			if v.Internal {
				if neu, ok := bmMap[v.Target]; ok {
					v.Target = neu
				}
			}
		case *element.ListItem:
			v.FontStyle = remapStyleRef(v.FontStyle, styleMap)
			v.ParagraphStyle = remapStyleRef(v.ParagraphStyle, styleMap)
		case *element.Table:
			if v.Style.StyleName != "" {
				if neu, ok := styleMap[v.Style.StyleName]; ok {
					v.Style.StyleName = neu
				}
			}
		case *element.Image:
			v.RelationID = 0
			if v.Media.Target != "" {
				v.Media.Target = bmPrefix + v.Media.Target
			}
		}
	})
}

func remapStyleRef(ref any, styleMap map[string]string) any {
	s, ok := ref.(string)
	if !ok || s == "" {
		return ref
	}
	if neu, hit := styleMap[s]; hit {
		return neu
	}
	return ref
}

func uniqueName(name string, used map[string]bool, prefix string) string {
	if name == "" {
		return name
	}
	if !used[name] {
		used[name] = true
		return name
	}
	cand := prefix + name
	n := 2
	for used[cand] {
		cand = prefix + name + strconv.Itoa(n)
		n++
	}
	used[cand] = true
	return cand
}
