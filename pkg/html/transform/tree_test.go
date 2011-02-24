/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/ 
package transform

import (
	"testing"
	v "container/vector"
	s "strings"
	. "html"
)

func assertEqual(t *testing.T, val interface{}, expected interface{}) {
	if val != expected {
		t.Errorf("NotEqual Expected: %s Actual %s",
			expected, val)
	}
}

func assertNotNil(t *testing.T, val interface{}) {
	if val == nil {
		t.Errorf("Value is Nil")
	}
}

func nodeTree() *Document {
	return NewDoc("<html><head /><body /></html>")
}

func TestWalk(t *testing.T) {
	tree := nodeTree()
	vec := new(v.StringVector)
	walkFun := func(n *Node) {
		vec.Push(n.Data)
	}
	Walk(tree.top, walkFun)
	assertEqual(t, vec.At(0), "") // first we have the root node
	assertEqual(t, vec.At(1), "html")
	assertEqual(t, vec.At(2), "head")
	assertEqual(t, vec.At(3), "body")
}

func TestFindAll(t *testing.T) {
	tree := nodeTree()
	vec := new(v.StringVector)
	walkFun := func(n *Node) {
		if s.Contains(n.Data, "head") {
			vec.Push(n.Data)
		}
	}
	Walk(tree.top, walkFun)
	assertEqual(t, vec.At(0), "head")
}

func TestParseHtml(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("TestParseHtml paniced: %s", err)
		}
	}()
	docStr := "<a href=\"bar\" class=\"baz\">foo</a>"
	node, _ := parseHtml(docStr)
	if node == nil {
		t.Error("Node was nil")
	}
	assertEqual(t, node.Type, DocumentNode)
	assertEqual(t, node.Child[0].Data, "a")
	assertEqual(t, len(node.Child), 1)
	assertEqual(t, len(node.Child[0].Attr), 2)
	assertEqual(t, node.Child[0].Attr[0].Key, "href")
	assertEqual(t, node.Child[0].Attr[0].Val, "bar")
	assertEqual(t, node.Child[0].Attr[1].Key, "class")
	assertEqual(t, node.Child[0].Attr[1].Val, "baz")
	assertEqual(t, node.Child[0].Type, ElementNode)
	assertEqual(t, node.Child[0].Child[0].Data, "foo")
	assertEqual(t, len(node.Child[0].Child), 1)
	assertEqual(t, node.Child[0].Child[0].Type, TextNode)
}

func TestNewDoc(t *testing.T) {
	docStr := "<a>foo</a>"
	doc := NewDoc(docStr)
	node := doc.top
	assertEqual(t, node.Child[0].Data, "a")
	assertEqual(t, len(node.Child), 1)
	assertEqual(t, node.Child[0].Type, ElementNode)
	assertEqual(t, node.Child[0].Child[0].Data, "foo")
	assertEqual(t, len(node.Child[0].Child), 1)
	assertEqual(t, node.Child[0].Child[0].Type, TextNode)
}
