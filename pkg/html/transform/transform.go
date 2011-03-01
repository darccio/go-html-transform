/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.

 The html transform package implements a html css selector and transformer.

 An html doc can be inspected and queried using css selectors as well as
 transformed.

 	doc := NewDoc(str)
 	sel1 := NewSelector("div.foo")
 	sel2 := NewSelector("a")
  t := NewTransform(doc)
 	newDoc := t.Apply(AppendChild, sel1)
  	.Apply(Replace, sel2)
  	.doc
*/
package transform

import (
	. "html"
)

// The TransformFunc type is the type of a Node transformation function.
type TransformFunc func(*Node)

type Transformer struct {
	doc *Document
}

func NewTransform(d *Document) *Transformer {
	return &Transformer{doc:d.Clone()}
}

func (t *Transformer) Apply(f TransformFunc, sel *SelectorQuery) *Transformer {

	return t
}

func AppendChild(c *Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Child)
		newChild := make([]*Node, sz+1)
		copy(newChild, n.Child)
		newChild[sz] = c
		n.Child = newChild
	}
}

func PrependChild(c *Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Child)
		newChild := make([]*Node, sz+1)
		copy(newChild[1:], n.Child)
		newChild[0] = c
		n.Child = newChild
	}
}

// TODO(jwall): helper transformation functions
// AppendChild(c *Node)
// PrependChild(c *Node)
// RemoveChildren()
// Replace()

// TODO(jwall): Function Modifiers
// DoTimes(TransformFunc, n)
// MakeAttribModifier(key, val)

