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

func TestRemoveChildren(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	f := RemoveChildren()
	f(node)
	assertEqual(t, len(node.Child), 0)
}

func TestReplaceChildren(t *testing.T) {
	doc := NewDoc("<div id=\"foo\">foo</div><")
	node := doc.top.Child[0]
	child := new(Node)
	child2 := new(Node)
	f := ReplaceChildren(child, child2)
	f(node)
	assertEqual(t, len(node.Child), 2)
	assertEqual(t, node.Child[0], child)
	assertEqual(t, node.Child[1], child2)
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

// TODO(jwall): benchmarking tests
