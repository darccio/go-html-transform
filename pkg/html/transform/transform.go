// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
/*

The html transform package implements a html css selector and transformer.

An html doc can be inspected and queried using css selectors as well as
transformed.

 	doc := NewDoc(str)
 	sel1 := NewSelector("div.foo")
 	sel2 := NewSelector("a")
	t := NewTransform(doc)
 	t.Apply(AppendChild, sel1)
  t..Apply(Replace, sel2)
  newDoc := t.Doc()
*/
package transform

// TODO(jwall): Documentation...
import (
	. "html"
	"log"
)

// The TransformFunc type is the type of a Node transformation function.
type TransformFunc func(*Node)

// Transformer encapsulates a document under transformation.
type Transformer struct {
	doc *Document
}

// Constructor for a Transformer. It makes a copy of the document
// and transforms that instead of the original.
func NewTransform(d *Document) *Transformer {
	return &Transformer{doc: d.Clone()}
}

// The Doc method returns the document under transformation.
func (t *Transformer) Doc() *Document {
	return t.doc
}

// The Apply method applies a TransformFunc to the nodes returned from
// the Selector query
func (t *Transformer) Apply(f TransformFunc, sel ...string) *Transformer {
	sq := NewSelectorQuery(sel...)
	nodes := sq.Apply(t.doc)
	for _, n := range nodes{
		f(n)
	}
	return t
}

// AppendChildren creates a TransformFunc that appends the Children passed in.
func AppendChildren(cs ...*Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Child)
		newChild := make([]*Node, sz+len(cs))
		copy(newChild, n.Child)
		copy(newChild[sz:], cs)
		n.Child = newChild
	}
}

// PrependChildren creates a TransformFunc that prepends the Children passed in.
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

// RemoveChildren creates a TransformFunc that removes the Children of the node
// it operates on.
func RemoveChildren() TransformFunc {
	return func(n *Node) {
		n.Child = make([]*Node, 0)
	}
}

// ReplaceChildren creates a TransformFunc that replaces the Children of the
// node it operates on with the Children passed in.
func ReplaceChildren(ns ...*Node) TransformFunc {
	return func(n *Node) {
		n.Child = ns
	}
}

// ModifyAttrb creates a TransformFunc that modifies the attributes
// of the node it operates on.
func ModifyAttrib(key string, val string) TransformFunc {
	return func(n *Node) {
		found := false
		for i, attr := range n.Attr {
			if attr.Key == key {
				n.Attr[i].Val = val
				found = true
			}
		}
		if !found {
			newAttr := make([]Attribute, len(n.Attr)+1)
			newAttr[len(n.Attr)] = Attribute{Key:key, Val:val}
			n.Attr = newAttr
		}
	}
}

func DoAll(fs ...TransformFunc) TransformFunc {
	return func(n *Node) {
		for _, f := range fs {
			f(n)
		}
	}
}

// ForEach takes a function and a list of Nodes and performs that
// function for each node in the list.
// The function should be of a type either func(...*Node) TransformFunc
// or func(*Node) TransformFunc. Any other type will panic.
// Returns a TransformFunc.
func ForEach(f interface{}, ns ...*Node) TransformFunc {
	return func(n *Node) {
		for _, n2 := range ns {
			switch t := f.(type) {
				case func(...*Node) TransformFunc:
					f1 := f.(func(...*Node) TransformFunc)
					f2 := f1(n2)
					f2(n)
				case func(*Node) TransformFunc:
					f1 := f.(func(*Node) TransformFunc)
					f2 := f1(n2)
					f2(n)
				default:
					log.Panicf("Wrong function type passed to ForEach %s", t) 
			}
		}
	}
}

// TODO(jwall): helper transformation functions
// Clone()?

// TODO(jwall): Function Modifiers
// DoTimes(n int, fs ...TransformFunc)?
