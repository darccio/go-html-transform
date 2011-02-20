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
)

func assertEqual(t *testing.T, val interface{}, expected interface{}) {
	if val != expected {
		t.Errorf("NotEqual Expected: %s Actual %s",
			expected, val)
	}
}

func nodeTree() *HtmlNode {
	node := new(HtmlNode)
	node.nodeValue = "top"
	children := new(v.Vector)
	node.children = children

	child1 := new(HtmlNode)
	child1.nodeValue = "child1"
	child2 := new(HtmlNode)
	child2.nodeValue = "child2"
	children.Push(child1)
	children.Push(child2)
	return node
}

func TestWalk(t *testing.T) {
	tree := nodeTree()
	vec := new(v.StringVector)
	walkFun := func(n *HtmlNode) {
		vec.Push(n.nodeValue)
	}
	Walk(tree, walkFun)
	assertEqual(t, vec.At(0), "top")
	assertEqual(t, vec.At(1), "child1")
	assertEqual(t, vec.At(2), "child2")
}

func TestFindAll(t *testing.T) {
	tree := nodeTree()
	vec := new(v.StringVector)
	walkFun := func(n *HtmlNode) {
		if s.Contains(n.nodeValue, "child") {
			vec.Push(n.nodeValue)
		}
	}
	Walk(tree, walkFun)
	assertEqual(t, vec.At(0), "child1")
	assertEqual(t, vec.At(1), "child2")
}
