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
	"os"
	"strings"
)

type Document struct {
	top *Node
}

func tokenToNode(tok *Token) *Node {
	node := new(Node)
	node.Data = tok.Data
	switch tok.Type {
		case TextToken:
			node.Type = TextNode
		case SelfClosingTagToken, StartTagToken:
			node.Type = ElementNode
	}
	return node
}

func parseHtml(s string) (top *Node, err os.Error) {
	r := strings.NewReader(s)
	z := NewTokenizer(r)
	top = new(Node)
	top.Type = DocumentNode
	q := new(v.Vector)
	q.Push(top)
	for {
		tt := z.Next()
		if tt == ErrorToken {
			if z.Error() != os.EOF { // some sort of error
				err = z.Error()
			} else {
				break // done parsing since end of file
			}
		} else {
			tok := z.Token()
			p := q.Last().(*Node)
			switch tok.Type {
				case TextToken, SelfClosingTagToken, StartTagToken:
					newChild := make([]*Node, len(p.Child)+1)
					copy(newChild[0:len(p.Child)], p.Child)
					p.Child = newChild
					node := tokenToNode(&tok)
				  newChild[len(newChild)-1] = node
					q.Push(node)
				case EndTagToken:
					q.Pop()
			}
		}
	}
	return top, err
}

func NewDoc(s string) *Document {
	n, err := parseHtml(s)
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
