/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
	. "html"
)

func TestNewTransform(t *testing.T) {
	doc := NewDoc("<html><body><div id=\"foo\"></div></body></html")
	tf := NewTransform(doc)
	// hacky way of comparing an uncomparable type
	assertEqual(t, (*tf.doc.top).Type, (*doc.top).Type)
}

func TestTransformApply(t *testing.T) {
	doc := NewDoc("<html><body><div id=\"foo\"></div></body></html")
	tf := NewTransform(doc)
	newDoc := tf.Apply(AppendChildren(new(Node)), "body").doc
	assertEqual(t, len(newDoc.top.Child[0].Child[0].Child), 2)
}

func TestAppendChildren(t *testing.T) {
	doc := NewDoc("<div id=\"foo\"></div><")
	node := doc.top
	child := new(Node)
	child2 := new(Node)
	f := AppendChildren(child, child2)
	f(node)
	assertEqual(t, len(node.Child), 3)
	assertEqual(t, node.Child[1], child)
	assertEqual(t, node.Child[2], child2)
}

func TestDocStringification(t *testing.T) {
	str := "<selfClosed /><div id=\"foo\">foo<a href=\"bar\"> bar</a></div>"
	doc := NewDoc(str)
	assertEqual(t, str, doc.String())
}

func TestRemoveChildren(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div>")
	node := doc.top.Child[0]
	f := RemoveChildren()
	f(node)
	assertEqual(t, len(node.Child), 0)
}

func TestReplaceChildren(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div>")
	node := doc.top.Child[0]
	child := new(Node)
	child2 := new(Node)
	f := ReplaceChildren(child, child2)
	f(node)
	assertEqual(t, len(node.Child), 2)
	assertEqual(t, node.Child[0], child)
	assertEqual(t, node.Child[1], child2)
}

func TestReplace(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("TestReplace paniced")
		}
	}()
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	ns := HtmlString("<span>foo</span>")
	f := Replace(ns...)
	f(node)
	assertEqual(t, len(doc.top.Child), 1)
	assertEqual(t, doc.top.Child[0].Data, "span")
}

func TestReplaceSplice(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("TestReplaceSplice paniced")
		}
	}()
	doc := NewDoc("<div id=\"foo\">foo<span>bar</span></div><")
	node := doc.top.Child[0].Child[0]
	ns := HtmlString("<span>foo</span>")
	f := Replace(ns...)
	f(node)
	assertEqual(t, len(doc.top.Child[0].Child), 2)
	assertEqual(t, doc.top.Child[0].Child[0].Data, "span")
	assertEqual(t, doc.top.Child[0].Child[0].Child[0].Data, "foo")
	assertEqual(t, doc.top.Child[0].Child[1].Child[0].Data, "bar")
}

func TestModifyAttrib(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	assertEqual(t, node.Attr[0].Val, "foo")
	f := ModifyAttrib("id", "bar")
	f(node)
	assertEqual(t, node.Attr[0].Val, "bar")
	f = ModifyAttrib("class", "baz")
	f(node)
	assertEqual(t, node.Attr[1].Key, "class")
	assertEqual(t, node.Attr[1].Val, "baz")
}

func TestTransformAttrib(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	assertEqual(t, node.Attr[0].Val, "foo")
	f := TransformAttrib("id", func(s string) string { return "bar"})
	f(node)
	assertEqual(t, node.Attr[0].Val, "bar")
}

func TestDoAll(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	preNode := new(Node)
	preNode.Data = "pre node"
	postNode := new(Node)
	postNode.Data = "post node"
	f := DoAll(AppendChildren(postNode),
		         PrependChildren(preNode))
	f(node)
	assertEqual(t, len(node.Child), 3)
	assertEqual(t, node.Child[0].Data, preNode.Data)
	assertEqual(t, node.Child[len(node.Child)-1].Data, postNode.Data)
}

func TestForEach(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	txtNode1 := Text(" bar")
	txtNode2 := Text(" baz")
	f := ForEach(AppendChildren, txtNode1, txtNode2)
	f(node)
	assertEqual(t, len(node.Child), 3)
	assertEqual(t, node.Child[1].Data, txtNode1.Data)
	assertEqual(t, node.Child[2].Data, txtNode2.Data)
}

func TestForEachSingleArgFuncs(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	txtNode1 := Text(" bar")
	txtNode2 := Text(" baz")
	singleArgFun := func(n *Node) TransformFunc {
		return AppendChildren(n)
	}
	f := ForEach(singleArgFun, txtNode1, txtNode2)
	f(node)
	assertEqual(t, len(node.Child), 3)
	assertEqual(t, node.Child[1].Data, txtNode1.Data)
	assertEqual(t, node.Child[2].Data, txtNode2.Data)
}

func TestForEachPanic(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("ForEach Failed to panic")
		}
	}()
	txtNode1 := Text(" bar")
	txtNode2 := Text(" baz")
	ForEach("foo", txtNode1, txtNode2)
}

func TestCopyAnd(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("TestCopyAnd paniced %s", err)
		}
	}()
	doc := NewDoc("<ul><li class=\"item\">item1</li></ul>")
	ul := doc.top.Child[0]
	node := ul.Child[0]
	fn1 := func(n *Node) {
		n.Child[0].Data = "foo"
	}
	fn2 := func(n *Node) {
		n.Child[0].Data = "bar"
	}
	f := CopyAnd(fn1, fn2)

	assertEqual(t, len(ul.Child), 1)
	f(node)
	assertEqual(t, len(ul.Child), 2)
	assertEqual(t, ul.Child[0].Data, "li")
	assertEqual(t, ul.Child[0].Attr[0].Key, "class")
	assertEqual(t, ul.Child[0].Attr[0].Val, "item")
	assertEqual(t, ul.Child[0].Child[0].Data, "foo")

	assertEqual(t, ul.Child[1].Data, "li")
	assertEqual(t, ul.Child[1].Attr[0].Key, "class")
	assertEqual(t, ul.Child[1].Attr[0].Val, "item")
	assertEqual(t, ul.Child[1].Child[0].Data, "bar")
}

// TODO(jwall): benchmarking tests
