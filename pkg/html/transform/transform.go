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

func AppendChildren(cs ...*Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Child)
		newChild := make([]*Node, sz+len(cs))
		copy(newChild, n.Child)
		copy(newChild[sz:], cs)
		n.Child = newChild
	}
}

func AppendChild(c *Node) TransformFunc {
	return AppendChildren(c)
}

func PrependChildren(cs ...*Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Child)
		sz2 := len(cs)
		newChild := make([]*Node, sz+len(cs))
		copy(newChild[sz2:], n.Child)
		copy(newChild[0:sz2], cs)
		n.Child = newChild
	}
}

func PrependChild(c *Node) TransformFunc {
	return PrependChildren(c)
}

func RemoveChildren() TransformFunc {
	return func(n *Node) {
		n.Child = make([]*Node, 0)
	}
}
// TODO(jwall): helper transformation functions
// RemoveChildren()
// ReplaceChildren()

// TODO(jwall): Function Modifiers
// DoTimes(TransformFunc, n)
// MakeAttribModifier(key, val)

