package h5

import (
	"fmt"
	"testing"
	//	"os"
	"reflect"
)

type tester interface {
	Errorf(msg string, args ...interface{})
}

func assertEqual(t tester, val, expected interface{}) {
	if !reflect.DeepEqual(val, expected) {
		t.Errorf("val: %v not equal to %v", val, expected)
	}
}

func assertTrue(t tester, ok bool, msg string, args ...interface{}) {
	if !ok {
		t.Errorf(msg, args...)
	}
}

func TestPushNode(t *testing.T) {
	p := new(Parser)
	assertTrue(t, p.Top == nil, "Top is not nil")
	assertTrue(t, p.curr == nil, "curr is not nil")
	top := pushNode(p)
	top.data = append(top.data, []rune("foo")...)
	assertTrue(t, p.Top != nil, "Top is still nil")
	assertTrue(t, p.curr != nil, "curr is stil nil")
	assertEqual(t, p.Top, top)
	assertEqual(t, p.curr, top)
	next := pushNode(p)
	next.data = append(next.data, []rune("bar")...)
	assertEqual(t, len(top.Children), 1)
	assertEqual(t, p.Top, top)
	assertEqual(t, p.curr, next)
	assertEqual(t, p.curr.Parent, p.Top)
}

func TestPopNode(t *testing.T) {
	p := new(Parser)
	top := pushNode(p)
	top.data = append(top.data, []rune("foo")...)
	next := pushNode(p)
	next.data = append(next.data, []rune("bar")...)
	popped := popNode(p)
	assertEqual(t, popped, top)
	assertEqual(t, p.Top, p.curr)
}

func TestAddSibling(t *testing.T) {
	p := new(Parser)
	top := pushNode(p)
	top.data = append(top.data, []rune("foo")...)
	next := pushNode(p)
	next.data = append(next.data, []rune("bar")...)
	sib := addSibling(p)
	sib.data = append(sib.data, []rune("baz")...)
	assertEqual(t, len(top.Children), 2)
	assertEqual(t, top.Children[0], next)
	assertEqual(t, top.Children[1], sib)
}

func TestBogusCommentHandlerNoEOF(t *testing.T) {
	p := NewParserFromString("foo comment >")
	top := pushNode(p)
	pushNode(p)
	st, err := bogusCommentHandler(p)
	assertEqual(t, len(top.Children), 2)
	assertEqual(t, string(top.Children[1].data), "foo comment ")
	assertTrue(t, st != nil, "next state handler is nil")
	assertTrue(t, err == nil, "err is not nil")
}

// TODO error cases
func TestBogusCommentHandlerEOF(t *testing.T) {
	p := NewParserFromString("foo comment")
	top := pushNode(p)
	pushNode(p)
	st, err := bogusCommentHandler(p)
	assertEqual(t, len(top.Children), 2)
	assertEqual(t, string(top.Children[1].data), "foo comment")
	assertTrue(t, st == nil, "next state handler is not nil")
	assertTrue(t, err != nil, "err is nil")
}

func TestEndTagOpenHandlerOk(t *testing.T) {
	p := NewParserFromString("FoO>")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st != nil, "next state handler is nil")
	assertEqual(t, err, nil)
	assertTrue(t, err == nil, "err is not nil")
	//assertTrue(t, p.curr == nil, "did not pop node")
}

func TestEndTagOpenHandlerTrunc(t *testing.T) {
	p := NewParserFromString("fo>")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st == nil, "next state handler is not nil")
	assertTrue(t, err != nil, "err is nil")
	assertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerLong(t *testing.T) {
	p := NewParserFromString("fooo>")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st == nil, "next state handler is not nil")
	assertTrue(t, err != nil, "err is nil")
	assertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerWrong(t *testing.T) {
	p := NewParserFromString("bar>")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st == nil, "next state handler is not nil")
	assertTrue(t, err != nil, "err is nil")
	assertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerBogusComment(t *testing.T) {
	p := NewParserFromString("f\no>")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st != nil, "next state handler is not nil")
	assertTrue(t, err != nil, "err is nil")
	assertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerEOF(t *testing.T) {
	p := NewParserFromString("foo")
	curr := pushNode(p)
	curr.data = []rune("foo")
	assertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	assertTrue(t, st == nil, "next state handler is nil")
	assertTrue(t, err != nil, "err is nil")
	assertEqual(t, p.curr, curr)
}

func TestTagNameHandler(t *testing.T) {
	p := NewParserFromString("f> ")
	curr := pushNode(p)
	st, err := handleChar(tagNameHandler)(p)
	assertTrue(t, st != nil, "next state handler is nil")
	assertTrue(t, err == nil, "err is not nil")
	assertEqual(t, curr.data[0], 'f')
	st, err = handleChar(tagNameHandler)(p)
	assertTrue(t, st != nil, "next state handler is nil")
	assertTrue(t, err == nil, "err is not nil")
	assertEqual(t, curr.data[0], 'f')
	p = NewParserFromString("F")
	curr = pushNode(p)
	st, err = handleChar(tagNameHandler)(p)
	assertTrue(t, st != nil, "next state handler is nil")
	assertTrue(t, err == nil, "err is not nil")
	assertEqual(t, curr.data[0], 'f')
}

func TestTagOpenHandler(t *testing.T) {
	p := NewParserFromString("")
	st := tagOpenHandler(p, 'f')
	assertTrue(t, st != nil, "next state handler is nil")
	//assertEqual(t, st, handleChar(tagNameHandler))
	assertEqual(t, p.curr.data[0], 'f')
	assertEqual(t, p.curr.Type, ElementNode)
}

func TestTagOpenHandlerEndTag(t *testing.T) {
	p := NewParserFromString("")
	st := tagOpenHandler(p, '/')
	assertTrue(t, st != nil, "next state handler is nil")
	//assertEqual(t, st, endTagOpenHandler)
}

func TestDataStateHandler(t *testing.T) {
	p := NewParserFromString("")
	st := dataStateHandler(p, '<')
	assertTrue(t, st != nil, "next state handler is nil")
	//assertEqual(t, st, handleChar(tagOpenHandler))
	assertTrue(t, p.curr == nil, "curr is currently nil")
	assertTrue(t, p.Top == nil, "Top is currently nil")
	p = NewParserFromString("oo<")
	st = dataStateHandler(p, 'f')
	assertTrue(t, st != nil, "next state handler is nil")
	assertTrue(t, p.curr != nil, "curr is currently nil")
	assertTrue(t, p.Top != nil, "Top is currently nil")
	assertEqual(t, p.curr.data, []rune("foo"))
}

func TestSimpledoc(t *testing.T) {
	p := NewParserFromString("<html><body>foo</body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	//fmt.Printf("XXX doc: %s\n", p.Top)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, len(p.Top.Children), 1)
	assertEqual(t, len(p.Top.Children[0].Children), 1)
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "foo")
}

func TestScriptDoc(t *testing.T) {
	p := NewParserFromString(
		"<html><body><script> if (foo < 10) { }</script></body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	//fmt.Printf("XXX doc: %s\n", p.Top)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, len(p.Top.Children), 1)
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, len(p.Top.Children[0].Children), 1)
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "script")
	assertEqual(t, p.Top.Children[0].Children[0].Children[0].Data(),
		" if (foo < 10) { }")
}

func TestScriptOnlyDoc(t *testing.T) {
	p := NewParserFromString(
		"<script> if (foo < 10) { var x = '<foo></foo>' }</script>")
	err := p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	//fmt.Printf("XXX doc: %s\n", p.Top)
	assertEqual(t, len(p.Top.Children), 1)
	assertEqual(t, p.Top.Children[0].Data(),
		" if (foo < 10) { var x = '<foo></foo>' }")
}

func TestSimpledocSiblings(t *testing.T) {
	p := NewParserFromString(
		"<html><body><a>foo</a><div>bar</div></body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	//fmt.Printf("XXX doc: %s\n", p.Top)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, len(p.Top.Children), 1)
	assertEqual(t, len(p.Top.Children[0].Children), 2)
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "a")
}

/* TODO Add a good test html page.
func TestParseFromReader(t *testing.T) {
	rdr, err := os.Open("test_data/page.html")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	p := NewParser(rdr)
	err = p.Parse()
	if err != nil {
		assertTrue(t, false, "Failed to parse")
	}
	assertTrue(t, p.Top != nil, "We got a parse tree back")
	//fmt.Println("Doc: ", p.Top.String())
}
*/

func TestNodeClone(t *testing.T) {
	p := NewParserFromString(
		"<html><body><a>foo</a><div>bar</div></body></html>")
	p.Parse()
	n := p.Top.Clone()
	assertTrue(t, n != nil, "n is nil")
	assertEqual(t, n.Data(), "html")
	assertEqual(t, len(n.Children), 1)
	assertEqual(t, len(n.Children[0].Children), 2)
	assertEqual(t, n.Children[0].Data(), "body")
	assertEqual(t, n.Children[0].Children[0].Data(), "a")
}

func TestNodeWalk(t *testing.T) {
	p := NewParserFromString(
		"<html><body><a>foo</a><div>bar</div></body></html>")
	p.Parse()
	i := 0
	ns := make([]string, 6)
	f := func(n *Node) {
		ns[i] = n.Data()
		i++
	}
	p.Top.Walk(f)
	assertEqual(t, i, 6)
	assertEqual(t, ns, []string{"html", "body", "a", "foo", "div", "bar"})
}

func TestSnippet(t *testing.T) {
	p := NewParserFromString("<a></a>")
	err := p.Parse()
	assertTrue(t, err == nil, "we errored while parsing snippet %s", err)
	assertTrue(
		t, p.Top != nil, "We didn't get a node tree back while parsing snippet")
	assertEqual(t, p.Top.Data(), "a")
}

func TestMeta(t *testing.T) {
	p := NewParserFromString(
		"<html><head><meta><link href='foo'></head><body><div>foo</div></body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err was not nil, %s", err)
	n := p.Top
	fmt.Println(p.Top)
	assertTrue(t, n != nil, "n is nil")
	assertEqual(t, n.Data(), "html")
	assertEqual(t, len(n.Children), 2)
	assertEqual(t, len(n.Children[0].Children), 2)
	assertEqual(t, n.Children[0].Data(), "head")
	assertEqual(t, n.Children[0].Children[0].Data(), "meta")
	assertEqual(t, n.Children[0].Children[1].Data(), "link")
	assertEqual(t, n.Children[1].Data(), "body")
	assertEqual(t, n.Children[1].Children[0].Data(), "div")
	assertEqual(t, n.Children[1].Children[0].Children[0].Data(), "foo")
}

func TestComment(t *testing.T) {
	p := NewParserFromString(
		"<html><head><!-- comment --></head><body><div>foo</div></body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err was not nil, %s", err)
	n := p.Top
	fmt.Println(p.Top)
	assertTrue(t, n != nil, "n is nil")
	assertEqual(t, n.Data(), "html")
	assertEqual(t, len(n.Children), 2)
	assertEqual(t, len(n.Children[0].Children), 1)
	assertEqual(t, n.Children[0].Data(), "head")
	assertEqual(t, n.Children[0].Children[0].Data(), " comment ")
	assertEqual(t, n.Children[0].Children[0].Type, CommentNode)
	assertEqual(t, n.Children[1].Data(), "body")
	assertEqual(t, n.Children[1].Children[0].Data(), "div")
	assertEqual(t, n.Children[1].Children[0].Children[0].Data(), "foo")
}

func TestUnclosedPTagInBody(t *testing.T) {
	p := NewParserFromString(
		"<html><body><p>foo<article></article></body></html>")
	err := p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, len(p.Top.Children[0].Children), 2)
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "p")
	assertEqual(t, len(p.Top.Children[0].Children[0].Children), 1)
	assertEqual(t, p.Top.Children[0].Children[0].Children[0].Data(), "foo")
	assertEqual(t, p.Top.Children[0].Children[1].Data(), "article")

	p = NewParserFromString(
		"<html><body><p><article></article></body></html>")
	err = p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, len(p.Top.Children[0].Children), 2)
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "p")
	assertEqual(t, len(p.Top.Children[0].Children[0].Children), 0)
	assertEqual(t, p.Top.Children[0].Children[1].Data(), "article")

	p = NewParserFromString(
		"<html><body><p>foo</body></html>")
	err = p.Parse()
	assertTrue(t, err == nil, "err is not nil: %v", err)
	assertEqual(t, p.Top.Data(), "html")
	assertEqual(t, p.Top.Children[0].Data(), "body")
	assertEqual(t, len(p.Top.Children[0].Children), 1)
	assertEqual(t, p.Top.Children[0].Children[0].Data(), "p")
	assertEqual(t, len(p.Top.Children[0].Children[0].Children), 1)
	assertEqual(t, p.Top.Children[0].Children[0].Children[0].Data(), "foo")
}

// TODO micro benchmarks
func BenchmarkDocParse(t *testing.B) {
	for i := 0; i < t.N; i++ {
		p := NewParserFromString(
			"<html><body><script> if (foo < 10) { }</script></body></html>")
		err := p.Parse()
		assertTrue(t, err == nil, "err is not nil: %v", err)
		//fmt.Printf("XXX doc: %s\n", p.Top)
		assertEqual(t, p.Top.Data(), "html")
		assertEqual(t, len(p.Top.Children), 1)
		assertEqual(t, p.Top.Children[0].Data(), "body")
		assertEqual(t, len(p.Top.Children[0].Children), 1)
		assertEqual(t, p.Top.Children[0].Children[0].Data(), "script")
		assertEqual(t, p.Top.Children[0].Children[0].Children[0].Data(),
			" if (foo < 10) { }")
	}
}
