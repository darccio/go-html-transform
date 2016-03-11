package selector

import (
	"go.marzhillstudios.com/pkg/go-html-transform/h5"

	"testing"

	"golang.org/x/net/html"
)

type testSpec struct {
	s  string
	n  *html.Node
	n2 *html.Node
	ns []*html.Node
}

func partial(s string) *html.Node {
	ns, _ := h5.PartialFromString(s)
	return ns[0]
}

func partials(s string) []*html.Node {
	ns, _ := h5.PartialFromString(s)
	return ns
}

var matchers = []testSpec{
	testSpec{
		"a",
		partial("<a></a>"),
		partial("<div></div>"),
		nil,
	},
	testSpec{
		"a.foo",
		partial("<a class=\"foo\"></a>"),
		partial("<div class=\"foo\"></div>"),
		nil,
	},
	testSpec{
		"a[class]",
		partial("<a class=\"foo\"></a>"),
		partial("<a></a>"),
		nil,
	},
	testSpec{
		"a[class=foo]",
		partial("<a class=\"foo\"></a>"),
		partial("<a class=\"fooo\"></a>"),
		nil,
	},
	testSpec{
		"a[class|=foo]",
		partial("<a class=\"foo-bar\"></a>"),
		partial("<a class=\"foobar\"></a>"),
		nil,
	},
	testSpec{
		"a[class~=foo]",
		partial("<a class=\"foo\"></a>"),
		partial("<a class=\"fooo\"></a>"),
		nil,
	},
	testSpec{
		"a[class~=foo]",
		partial("<a class=\"foo bar\"></a>"),
		partial("<a class=\"fooo bar\"></a>"),
		nil,
	},
	testSpec{
		"a[class~=foo]",
		partial("<a class=\"baz foo bar\"></a>"),
		partial("<a class=\"baz foo0 bar\"></a>"),
		nil,
	},
}

var finders = []testSpec{
	testSpec{
		"a[class~=foo]",
		partial("<a class=\"baz foo bar\"></a>"),
		nil,
		partials("<a class=\"baz foo bar\"></a>"),
	},
	testSpec{
		"a",
		partial("<div><a>foo</a><a class=\"baz foo bar\"></a></div>"),
		nil,
		partials("<a>foo</a><a class=\"baz foo bar\"></a>"),
	},
	testSpec{
		"a.foo",
		partial("<div><a>foo</a><a class=\"baz foo bar\"></a></div>"),
		nil,
		partials("<a class=\"baz foo bar\"></a>"),
	},
	testSpec{
		"span",
		partial("<div><span>foo<span>bar</span></span><span class=\"baz foo bar\"></span></div>"),
		nil,
		partials("<span>foo<span>bar</span></span><span>bar</span><span class=\"baz foo bar\"></span>"),
	},
	testSpec{
		"div>span",
		partial("<div><span>foo<span>bar</span></span><p class=\"baz foo bar\"><span>baz</span></p></div>"),
		nil,
		partials("<span>foo<span>bar</span></span>"),
	},
	testSpec{
		"div+span",
		partial("<div><div>foo</div><span>bar</span><span>baz</span></div>"),
		nil,
		partials("<span>bar</span>"),
	},
	testSpec{
		"div+span",
		partial("<div><span>bar</span><span>baz</span><div>foo</div></div>"),
		nil,
		partials("<span>baz</span>"),
	},
	testSpec{
		"div+span",
		partial("<div><span>foobar</span><span>baz</span><div>foo</div><span>bar</span></div>"),
		nil,
		partials("<span>baz</span><span>bar</span>"),
	},
	testSpec{
		"div~span",
		partial("<div><div>foo</div><span>bar</span><span>baz</span></div>"),
		nil,
		partials("<span>bar</span><span>baz</span>"),
	},
	testSpec{
		"div span",
		partial("<div><p>foo</p><span>bar</span><span>baz<span>quux</span></span></div>"),
		nil,
		partials("<span>bar</span><span>baz<span>quux</span></span><span>quux</span>"),
	},
	testSpec{
		"div span",
		partial("<div><div><span>foo</span></div><div><span>bar</span></div></div>"),
		nil,
		partials("<span>foo</span><span>bar</span>"),
	},
	testSpec{
		":empty",
		partial("<div><div></div></div>"),
		nil,
		partials("<div></div>"),
	},
	testSpec{
		"a:first-child",
		partial("<div><a>foo</a><b>baz</b></div>"),
		nil,
		partials("<a>foo</a>"),
	},
	testSpec{
		"b:last-child",
		partial("<div><a>foo</a><b>baz</b></div>"),
		nil,
		partials("<b>baz</b>"),
	},
	testSpec{
		"a:only-child",
		partial("<div><a>foo</a></div>"),
		nil,
		partials("<a>foo</a>"),
	},
}

func TestSelectorFind(t *testing.T) {
	for _, spec := range finders {
		chn, err := Selector(spec.s)
		if err != nil {
			t.Errorf("Error parsing selector %q", err)
		}
		ns := chn.Find(spec.n)
		if len(ns) < 1 {
			t.Errorf("%q didn't find any nodes in %q",
				chn, h5.RenderNodesToString([]*html.Node{spec.n}))
		}
		if h5.RenderNodesToString(ns) != h5.RenderNodesToString(spec.ns) {
			t.Errorf("Got: %q Expected: %q",
				h5.RenderNodesToString(ns), h5.RenderNodesToString(spec.ns))
		}
	}
}

func TestSelectorMatch(t *testing.T) {
	for _, spec := range matchers {
		chn, err := Selector(spec.s)
		if err != nil {
			t.Errorf("Error parsing selector %q", err)
		}
		if !chn.Head.Match(spec.n) {
			t.Errorf("spec %q didn't match %q when it should have",
				chn, h5.RenderNodesToString([]*html.Node{spec.n}))
		}
		if chn.Head.Match(spec.n2) {
			t.Errorf("spec %q matched %q when it shouldn't have",
				chn, h5.RenderNodesToString([]*html.Node{spec.n2}))
		}
	}
}
