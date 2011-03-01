package transform

import (
	v "container/vector"
	. "html"
	l "log"
	"os"
	"strings"
)

// Document is the type of a parsed html string.
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
	node.Attr = tok.Attr
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
				copy(newChild, p.Child)
				p.Child = newChild
				node := tokenToNode(&tok)
				node.Parent = p
				newChild[len(newChild)-1] = node
				if tok.Type != SelfClosingTagToken &&
					tok.Type != TextToken {
					q.Push(node)
				}
			case EndTagToken:
				q.Pop()
			}
		}
	}
	return top, err
}

// NewDoc is a constructor for a Document.
func NewDoc(s string) *Document {
	n, err := parseHtml(s)
	if err != nil {
		l.Panicf("Failure parsing html \n %s", s)
	}
	return &Document{top: n}
}

func walk(n *Node, f func(*Node)) {
	f(n)
	c := n.Child
	if c != nil {
		for i := 0; i < len(c); i++ {
			c_node := c[i]
			walk(c_node, f)
		}
	}
}

// The Top Method returns the root node of the parsed html string.
// This node is not a parsed html node it is empty. The actual parsed
// nodes can be found by calling the Nodes method.
// This allows a Document to contain a full html document or
// partial fragment.
// Returns a *Node.
func (d *Document) Top() *Node {
	return d.top
}

// The Nodes method returns the parsed nodes of the html string.
// There may be multiple nodes if the parsed string was fragment
// and not a full document.
// Returns a []*Node.
func (d *Document) Nodes() []*Node {
	return d.Top().Child
}

// The Walk method walks a Documents node tree running
// The passed in function on it.
func (d *Document) Walk(f func(*Node)) {
	walk(d.top, f)
}

func cloneNode(n *Node, p *Node) *Node {
	node := new(Node)
	node.Parent = p
	node.Data = n.Data

	if n.Type != 0 {
		node.Type = n.Type
	}

	newChild := make([]*Node, len(n.Child))
	for i, c := range n.Child {
		newChild[i] = cloneNode(c, node)
	}
	node.Child = newChild

	newAttr := make([]Attribute, len(n.Attr))
	copy(newAttr, n.Attr)
	node.Attr = newAttr
	return node
}

// The Clone method creates a deep copy of the Document.
func (d *Document) Clone() *Document {
	doc := new(Document)
	doc.top = cloneNode(d.top, nil)
	return doc
}

// The FindAll method searches the Document's node tree for
// anything the passed in function returns true for.
// Returns a vector of the found nodes.
func (d *Document) FindAll(f func(*Node) bool) *v.Vector {
	results := new(v.Vector)
	fun := func(node *Node) {
		if f(node) {
			results.Push(node)
		}
	}
	d.Walk(fun)
	return results
}

func Text(str string) *Node {
	return &Node{Data:str, Type:TextNode}
}

func HtmlString(str string) ([]*Node, os.Error) {
	parsed, err := parseHtml(str)
	if err == nil {
		return parsed.Child, nil
	}
	return nil, err
}

// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
