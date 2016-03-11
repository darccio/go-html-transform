/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"

	"go.marzhillstudios.com/pkg/go-html-transform/h5"
)

func assertEqual(t *testing.T, val interface{}, expected interface{}) {
	if val != expected {
		t.Errorf("NotEqual Expected: [%s] Actual: [%s]",
			expected, val)
	}
}

func assertNotNil(t *testing.T, val interface{}) {
	if val == nil {
		t.Errorf("Value is Nil")
	}
}

func TestNewTransformer(t *testing.T) {
	tree, _ := h5.NewFromString("<html><body><div id=\"foo\"></div></body></html>")
	tf := New(tree)
	// hacky way of comparing an uncomparable type
	assertEqual(t, tf.Doc().Type, tree.Top().Type)
}

func TestTransformApply(t *testing.T) {
	tree, _ := h5.NewFromString("<html><body><div id=\"foo\"></div></body></html>")
	tf := New(tree)
	n := h5.Text("bar")
	tf.Apply(AppendChildren(n), "body")
	newDoc := tf.String()
	assertEqual(t, newDoc, "<html><head></head><body><div id=\"foo\"></div>bar</body></html>")
}

func TestTransformApplyAll(t *testing.T) {
	tree, _ := h5.NewFromString("<html><head></head><body><ul><li>foo</ul></body></html>")
	tf := New(tree)
	n := h5.Text("bar")
	n2 := h5.Text("quux")
	t1, _ := Trans(AppendChildren(n), "body li")
	t2, _ := Trans(AppendChildren(n2), "body li")
	tf.ApplyAll(t1, t2)
	assertEqual(t, tf.String(), "<html><head></head><body><ul><li>foobarquux</li></ul></body></html>")
}

func TestTransformApplyMulti(t *testing.T) {
	tree, _ := h5.NewFromString("<html><body><div id=\"foo\"></div></body></html>")
	tf := New(tree)
	tf.Apply(AppendChildren(h5.Text("")), "body")
	tf.Apply(TransformAttrib("id", func(val string) string {
		t.Logf("Rewriting Url")
		return "bar"
	}),
		"div")
	newDoc := tf.String()
	assertEqual(t, newDoc, "<html><head></head><body><div id=\"bar\"></div></body></html>")
}

func TestAppendChildren(t *testing.T) {
	node := h5.Anchor("", "")
	child := h5.Text("foo ")
	child2 := h5.Text("bar")
	AppendChildren(child, child2)(node)
	assertEqual(t, h5.NewTree(node).String(), "<a>foo bar</a>")
}

func TestRemoveChildren(t *testing.T) {
	node := h5.Anchor("", "foo")
	RemoveChildren()(node)
	assertEqual(t, h5.NewTree(node).String(), "<a></a>")
}

func TestReplaceChildren(t *testing.T) {
	node := h5.Anchor("", "foo")
	assertEqual(t, h5.NewTree(node).String(), "<a>foo</a>")
	child := h5.Text("baz ")
	child2 := h5.Text("quux")
	ReplaceChildren(child, child2)(node)
	assertEqual(t, h5.NewTree(node).String(), "<a>baz quux</a>")
}

func TestReplace(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("TestReplace paniced")
		}
	}()
	node := h5.Div("", nil, h5.Div("", nil, h5.Text("foo")))
	replacement := h5.Div("", nil, h5.Text("bar"))
	Replace(replacement)(node.FirstChild)
	assertEqual(t, h5.NewTree(node).String(),
		"<div><div>bar</div></div>")
}

func TestReplaceSplice(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("TestReplaceSplice paniced")
		}
	}()
	node := h5.Div("foo", nil,
		h5.Text("foo"),
		h5.Element("span", nil, h5.Text("bar")),
	)
	node2 := h5.Element("span", nil, h5.Text("foo"))
	Replace(node2)(node.FirstChild)
	assertEqual(t, h5.NewTree(node).String(),
		"<div id=\"foo\"><span>foo</span><span>bar</span></div>")
}

func TestReplaceSpliceOnRootNode(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("TestReplaceSpliceOnRootNode didn't panic")
		}
	}()
	tree, _ := h5.NewFromString("<div id=\"foo\">foo<span>bar</span></div><")
	doc := tree.Top()
	ns, _ := h5.NewFromString("<span>foo</span>")
	f := Replace(ns.Top())
	f(doc)
	assertEqual(t, h5.Data(doc.FirstChild), "span")
	assertEqual(t, h5.Data(doc.FirstChild.FirstChild), "foo")
}

func TestModifyAttrib(t *testing.T) {
	node := h5.Anchor("", "")
	ModifyAttrib("id", "bar")(node)
	assertEqual(t, node.Attr[0].Val, "bar")
	ModifyAttrib("class", "baz")(node)
	assertEqual(t, node.Attr[1].Key, "class")
	assertEqual(t, node.Attr[1].Val, "baz")
}

func TestTransformAttrib(t *testing.T) {
	node := h5.Anchor("", "")
	ModifyAttrib("id", "foo")(node)
	assertEqual(t, node.Attr[0].Val, "foo")
	TransformAttrib("id", func(s string) string { return "bar" })(node)
	assertEqual(t, node.Attr[0].Val, "bar")
}

func TestDoAll(t *testing.T) {
	tree, _ := h5.NewFromString("<div id=\"foo\">foo</div><")
	node := tree.Top()
	preNode := h5.Text("pre node")
	postNode := h5.Text("post node")
	f := DoAll(AppendChildren(postNode),
		PrependChildren(preNode))
	f(node)
	assertEqual(t, h5.Data(node.FirstChild), h5.Data(preNode))
	assertEqual(t, h5.Data(node.LastChild), h5.Data(postNode))
}

func TestCopyAnd(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("TestCopyAnd paniced %s", err)
		}
	}()
	node := h5.Div("", nil, h5.Div("", nil, h5.Text("foo")))
	assertEqual(t, h5.NewTree(node).String(),
		"<div><div>foo</div></div>")
	CopyAnd(
		AppendChildren(h5.Text("bar")),
		ReplaceChildren(h5.Text("baz")),
	)(node.FirstChild)
	assertEqual(t, h5.NewTree(node).String(),
		"<div><div>foobar</div><div>baz</div></div>")
}

func TestTransformSubtransforms(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("TestTransformSubtransforms paniced %s", err)
		}
	}()
	tree, _ := h5.NewFromString("<html><body><ul><li>foo</ul></body></html>")

	f, _ := Subtransform(CopyAnd(
		ReplaceChildren(h5.Text("bar")),
		ReplaceChildren(h5.Text("baz"), h5.Text("quux")),
	), "li")
	tf := New(tree)
	t1, _ := Trans(f, "ul")
	tf.ApplyAll(t1)
	assertEqual(t, tf.String(),
		"<html><head></head><body><ul><li>bar</li><li>bazquux</li></ul></body></html>")

}

// TODO(jwall): benchmarking tests
func BenchmarkTransformApply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree, _ := h5.NewFromString("<html><body><div id=\"foo\"></div></body></html")
		tf := New(tree)
		tf.Apply(AppendChildren(h5.Text("")), "body")
		tf.Apply(TransformAttrib("id", func(val string) string {
			return "bar"
		}),
			"div")
		tf.Doc()
	}
}
