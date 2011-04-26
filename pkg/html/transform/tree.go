package transform

import (
	v "container/vector"
	. "html"
	l "log"
	"os"
	"io"
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

func parseHtml(s string) (*Node, os.Error) {
	r := strings.NewReader(s)
	return readHtml(r)
}

func emptyElement(tok *Token) bool {
	return tok.Data == "area" ||
		tok.Data == "base" ||
		tok.Data == "basefont" ||
		tok.Data == "br" ||
		tok.Data == "col" ||
		tok.Data == "frame" ||
		tok.Data == "input" ||
		tok.Data == "img" ||
		tok.Data == "link" ||
		tok.Data == "meta" ||
		tok.Data == "param"
}

var okError = "html: TODO"

func inScript(z *Tokenizer, script *Node) (ok bool, err os.Error) {
	// we need to consume the raw tokens until we come to the script token
	l.Printf("handling the script tag")
	ok = false
	node := new(Node)
	node.Type = TextNode
	node.Data = ""
	script.Child = append(script.Child, node)
	for {
		tt := z.Next()
		raw := z.Raw()
		if tt == ErrorToken {
			err = z.Error()
			errMsg := err.String()
			if len(errMsg) >= len(okError) &&
				errMsg[:len(okError)] == okError {
				// TODO(jwall): find a safe way to handle this.
				return
			} else {
				return
			}
		}
		l.Printf("adding [%s] to script contents", raw)
		tok := z.Token()
		if tok.Data == "script" {
			l.Printf("exiting script tag")
			ok = true
			return
		} else {
			node.Data += string(raw)
		}
	}
	l.Panicf("Script tag never terminated")
	// impossible to reach
	return
}

func readHtml(r io.Reader) (top *Node, err os.Error) {
	// TODO(jwall): start using the Scanner instead of the tokenizer
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
			l.Printf("handling raw token: %s", z.Raw())
			tok := z.Token()
			p := q.Last().(*Node)
			switch tok.Type {
			case TextToken, SelfClosingTagToken, StartTagToken:
				node := tokenToNode(&tok)
				if tok.Data == "script" {
					ok, sErr := inScript(z, node)
					if !ok && sErr != nil {
						l.Panicf("Error parsing script: %s", sErr)
					}
				}
				newChild := make([]*Node, len(p.Child)+1)
				copy(newChild, p.Child)
				p.Child = newChild
				node.Parent = p
				newChild[len(newChild)-1] = node
				if tok.Type != SelfClosingTagToken &&
					tok.Type != TextToken &&
					!emptyElement(&tok) {
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

func NewDocFromReader(r io.Reader) *Document {
	n, err := readHtml(r)
	if err != nil {
		l.Panicf("Failure parsing html from reader")
	}
	return &Document{top: n}
}

func (d *Document) String() string {
	return toString(d.top)
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

func attribsString(n *Node) string {
	str := ""
	for _, attr := range n.Attr {
		str += " " + attr.Key + "=\"" + attr.Val + "\""
	}
	return str
}

func toString(n *Node) string {
	str := ""
	switch n.Type {
	case DocumentNode:
		for _, c := range n.Child {
			str += toString(c)
		}
	case TextNode:
		str = n.Data
	case ElementNode:
		str += "<" + n.Data + attribsString(n)
		if len(n.Child) > 0 {
			str += ">"
			for _, c := range n.Child {
				str += toString(c)
			}
			str += "</" + n.Data + ">"
		} else {
			// this is a self-closing tag
			str += "></" + n.Data + ">"
		}
	}
	return str
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

// Constructs a TextNode for the string passed in
func Text(str string) *Node {
	return &Node{Data: str, Type: TextNode}
}

// Constructs a slice of *Nodes from a string of html.
func HtmlString(str string) []*Node {
	parsed, err := parseHtml(str)
	if err == nil {
		return parsed.Child
	}
	return nil
}

// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
