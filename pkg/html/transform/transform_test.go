/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.

 The html transform package implements a html css selector and transformer.

 An html doc can be inspected and queried using css selectors as well as
 transformed.

 	doc := NewDoc(str)
 	sel := NewSelector("a", ".foo")
 	node := sel.Apply(doc)

 	transformer := func(node Node) Node { ... }
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

func TestAppendChild(t *testing.T) {
	doc := NewDoc("<div id=\"foo\"></div><")
 	node := doc.top
	child := new(Node)
  f := AppendChild(child)
	f(node)
  assertEqual(t, len(node.Child), 2)
	assertEqual(t, node.Child[1], child)
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

func TestPrependChild(t *testing.T) {
	doc := NewDoc("<div id=\"foo\"></div><")
 	node := doc.top
	child := new(Node)
  f := PrependChild(child)
	f(node)
  assertEqual(t, len(node.Child), 2)
	assertEqual(t, node.Child[0], child)
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
