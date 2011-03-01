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
 	node := doc.top.Child[0]
	child := new(Node)
  f := AppendChild(child)
	f(node)
	t.Logf("node: %s", *node)
	t.Logf("child 1: %s", *(node.Child[0]))
  assertEqual(t, len(node.Child), 1)
}
