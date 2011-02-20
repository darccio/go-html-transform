/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/ 
package transform

import (
	v "container/vector"
	. "html"
	l "log"
	"strings"
)

type Document struct {
	top *Node
}

func NewDoc(s string) *Document {
	n, err := Parse(strings.NewReader(s))
	if err != nil {
		l.Panicf("Failure parsing html \n %s", s)
	}
	return &Document{top: n}
}

func Walk(n *Node, f func(*Node)) {
	f(n)
	c := n.Child
	if c != nil {
		for i := 0; i < len(c); i++ {
			c_node := c[i]
			Walk(c_node, f);
		}
	}
}

func (n *Document) FindAll(f func(*Node) bool) *v.Vector {
	results := new(v.Vector)
	fun := func(node *Node) {
		if f(node) {
			results.Push(node)
		}
	}
	Walk(n.top, fun)
	return results
}
