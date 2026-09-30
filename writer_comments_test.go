package word

import (
	"bytes"
	"github.com/yunkeweb/go-word/pkg/common"
	"strings"
	"testing"
)

func TestCommentsSurviveRepeatedWrites(t *testing.T) {
	for _, reuseWriter := range []bool{false, true} {
		d := New()
		d.CommentOn("one", "first comment", "author")
		writer := newWord2007Writer(d)
		for i := 0; i < 3; i++ {
			if !reuseWriter {
				writer = newWord2007Writer(d)
			}
			if i == 2 {
				d.CommentOn("two", "second comment", "author")
			}
			var out bytes.Buffer
			if _, err := writer.WriteTo(&out); err != nil {
				t.Fatal(err)
			}
			zr, err := common.OpenZipBytes(out.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			comments, err := zr.ReadFile("word/comments.xml")
			if err != nil {
				t.Fatalf("reuse=%v write=%d: %v", reuseWriter, i, err)
			}
			count := 1
			if i == 2 {
				count = 2
			}
			if got := strings.Count(string(comments), "<w:comment "); got != count {
				t.Fatalf("write %d: %d comments, want %d", i, got, count)
			}
			rels, _ := zr.ReadFile("word/_rels/document.xml.rels")
			if !bytes.Contains(rels, []byte("comments.xml")) {
				t.Fatal("missing comments relationship")
			}
			if !bytes.Contains(comments, []byte("first comment")) {
				t.Fatal("missing first comment text")
			}
			zr.Close()
		}
	}
}
