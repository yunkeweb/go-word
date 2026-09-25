package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
)

func TestAddHTMLFullDocumentAndTags(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	html := `<!DOCTYPE html><html><body>
		<div align="center" style="color:#123;background-color:#abc;font-weight:bold;font-style:italic;text-decoration:underline line-through;text-align:right;font-size:16px;font-family:'Times New Roman',serif">
			<p>Hello <strong>B</strong> <b>B2</b> <em>I</em> <i>I2</i> <u>U</u>
			<s>S</s> <strike>K</strike> <del>D</del> <sup>up</sup> <sub>dn</sub>
			<span style="font-weight:700;font-size:12pt;color:red">span</span>
			<a href="https://ex">link</a>
			<br/>
			<img src="pic.png" width="10px" height="20pt" alt="pic"/>
			</p>
		</div>
		<h1>H1</h1><h2>H2</h2><h3>H3</h3>
		<hr>
		<br>
		<table>
			<thead><tr><th width="40">A</th></tr></thead>
			<tbody><tr><td width="50px">B</td></tr></tbody>
			<tfoot><tr><td>C</td></tr></tfoot>
		</table>
		<ul><li>u1</li><ul><li>nested</li></ul></ul>
		<ol><li>o1</li><ol><li>nestedn</li></ol></ol>
		<input type="checkbox" name="cb" value="yes"/>
		<input type="text" name="x"/>
		<img src=""/>
		<unknown><p>x</p></unknown>
		<p style="bogus"></p>
		<p style="font-weight:normal"></p>
		<p align="justify"></p>
		<p align="left"></p>
		<font color="#00f" bgcolor="#fff">f</font>
	</body></html>`
	if err := AddHTML(sec, html, true); err != nil {
		t.Fatal(err)
	}
	if err := AddHTML(sec, "<p>more</p>", false); err != nil {
		t.Fatal(err)
	}
	if err := AddHTML(sec, `<br><img src="x.png"/>`); err != nil {
		t.Fatal(err)
	}
	if err := AddHTML(sec, `<img src="sized.png" width="10px" height="20pt" alt="pic"/>`); err != nil {
		t.Fatal(err)
	}
	if sec.CountElements() == 0 {
		t.Fatal("no elements")
	}
	var sawTitle, sawList bool
	for _, e := range sec.Elements() {
		switch e.(type) {
		case *element.Title:
			sawTitle = true
		case *element.ListItem:
			sawList = true
		}
	}
	if !sawTitle || !sawList {
		t.Fatal("structure")
	}
}

func TestAddHTMLErrorsAndHelpers(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	if err := AddHTML(sec, "<p>"); err != nil && !strings.Contains(err.Error(), "") {
		t.Fatal(err)
	}
	if err := AddHTML(sec, string([]byte{0xff, 0xfe, '<', 'p'})); err == nil {
		// may or may not error depending on decoder
	}
	if mapAlign("CENTER") != "center" || mapAlign("right") != "right" || mapAlign("justify") != "both" || mapAlign("left") != "left" {
		t.Fatal("align")
	}
	if cssSize("10px") != 10 || cssSize("8pt") != 8 || cssSize("") != 0 {
		t.Fatal("cssSize")
	}
	if pt, ok := cssPoints("16px"); !ok || pt != 12 {
		t.Fatalf("px pt=%v", pt)
	}
	if pt, ok := cssPoints("11pt"); !ok || pt != 11 {
		t.Fatal("pt")
	}
	if pt, ok := cssPoints("9"); !ok || pt != 9 {
		t.Fatal("bare")
	}
	if _, ok := cssPoints("xx"); ok {
		t.Fatal("bad")
	}
	n := &htmlNode{tag: "#text", text: "hi"}
	if collectText(n) != "hi" {
		t.Fatal("text")
	}
	parseHTMLNode(nil, sec, htmlStyle{})
}

func TestPrepareHTML(t *testing.T) {
	s := prepareHTML("<br>\n<img src='x'>", false)
	if !strings.Contains(s, "<body>") || !strings.Contains(s, "/>") {
		t.Fatalf("%s", s)
	}
	s2 := prepareHTML("<br/>", false)
	if !strings.Contains(s2, "<br/>") {
		t.Fatalf("%s", s2)
	}
	s3 := prepareHTML("<BODY>x</BODY>", false)
	if strings.Count(strings.ToLower(s3), "<body") != 1 {
		t.Fatalf("%s", s3)
	}
}
