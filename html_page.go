package word

import (
	"strconv"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

// Paper preview preserves the section's text width; it is not a Word pagination
// engine. A long section grows vertically on screen. Print margins belong to
// @page, not the section, so they repeat without doubling the preview padding.
const pageHTMLCSS = `
body.goword-document{margin:0;padding:24px 0;background:#e7e7e7;color:#000}
.goword-page{box-sizing:border-box;display:flow-root;margin:0 auto 24px;background:#fff;box-shadow:0 0 1px #999;overflow:visible}
.goword-header,.goword-footer{break-inside:avoid}
@media print{
body.goword-document{padding:0;background:#fff}
.goword-page{display:block;width:auto!important;min-height:0!important;padding:0!important;margin:0!important;box-shadow:none}
.goword-page+.goword-page{break-before:page}
.goword-page[data-break-type="continuous"]{break-before:auto}
.goword-page[data-break-type="evenPage"]{break-before:left}
.goword-page[data-break-type="oddPage"]{break-before:right}
body.goword-document[data-repeat-header="true"] .goword-header[data-type="default"]{position:fixed;top:0;left:0;right:0;z-index:1}
body.goword-document[data-repeat-footer="true"] .goword-footer[data-type="default"]{position:fixed;right:0;bottom:0;left:0;z-index:1}
}
`

type htmlPageGeometry struct {
	width, height            int
	top, right, bottom, left int
}

func sectionPageGeometry(s style.Section) htmlPageGeometry {
	g := htmlPageGeometry{
		width: s.PageSizeW, height: s.PageSizeH,
		top: maxInt(s.MarginTop, 0), right: maxInt(s.MarginRight, 0),
		bottom: maxInt(s.MarginBottom, 0), left: maxInt(s.MarginLeft, 0),
	}
	if g.width <= 0 {
		g.width = style.DefaultPageWidth
	}
	if g.height <= 0 {
		g.height = style.DefaultPageHeight
	}
	if s.Orientation == style.OrientationLandscape && g.height > g.width {
		g.width, g.height = g.height, g.width
	}
	if s.RtlGutter {
		g.right += maxInt(s.Gutter, 0)
	} else {
		g.left += maxInt(s.Gutter, 0)
	}
	return g
}

func (g htmlPageGeometry) margins() string {
	return twipsPx(g.top) + " " + twipsPx(g.right) + " " + twipsPx(g.bottom) + " " + twipsPx(g.left)
}

func (g htmlPageGeometry) screenStyle() string {
	return "width:" + twipsPx(g.width) + ";min-height:" + twipsPx(g.height) + ";padding:" + g.margins()
}

func (r *htmlRenderer) writePageRules(sections []*element.Section) []string {
	names := make([]string, len(sections))
	known := make(map[htmlPageGeometry]string)
	for i, sec := range sections {
		if sec == nil {
			continue
		}
		geometry := sectionPageGeometry(sec.Style)
		name, ok := known[geometry]
		if !ok {
			name = "goword-section-" + strconv.Itoa(i+1)
			known[geometry] = name
			r.writeString("@page " + name + "{size:" + twipsPx(geometry.width) + " " + twipsPx(geometry.height) + ";margin:" + geometry.margins() + "}")
		}
		// Share a page name for equal geometry so a continuous section need
		// not force a printed page solely because its section index changed.
		names[i] = name
	}
	return names
}
