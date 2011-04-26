package transformbench

import (
	. "html/transform"
	"os"
	"testing"
)

func benchParseReader(b *testing.B, f *os.File) {
	for i := 0; i < b.N; i++ {
		NewDocFromReader(f)
	}
	b.StopTimer()
	fi, _ := f.Stat()
	b.SetBytes(fi.Size*int64(b.N))
}

func benchParse(b *testing.B, docStr string) {
	for i := 0; i < b.N; i++ {
		NewDoc(docStr)
	}
	b.SetBytes(int64(len(docStr)*b.N))
}

func BenchmarkNewDocSingleElement(b *testing.B) {
	benchParse(b, "<a>foo</a>")
}

func BenchmarkNewDocMultiElement(b *testing.B) {
	benchParse(b, "<a>foo</a><div>bar</div>")
}

func BenchmarkNewDocComplex(b *testing.B) {
	benchParse(b, "<html><head>\n	<meta http-equiv=\"something\" />\n	<link href=\"thing.css\" >\n</head>\n<body>\n	<a>foo</a>\n	<div id=\"foobar\">bar</div>\n</body></html>")
}

func BenchmarkNewDocRealPage(b *testing.B) {
	f, _ := os.OpenFile("walljm.com.html", os.O_RDONLY, 0666)
	benchParseReader(b, f)	
}
