/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/ 
package transform

import (
	v "container/vector"
	. "html"
	"log"
	"os"
	"strings"
)

type NodeType int

const (
	TEXT NodeType = iota // 0 value so the default
	TAG
)

type Node struct {
	nodeType NodeType
	nodeValue string
	nodeAttributes map[string] string
	children v.Vector
}

func (n *Node) Copy(node Node) {
	n.nodeType = node.nodeType
	n.nodeValue = node.nodeValue
	n.nodeAttributes = node.nodeAttributes
	n.children = node.children
}

func lazyTokens(t *Tokenizer) <-chan Token {
	tokens := make(chan Token, 1)
	go func() {
		for {
			tt := t.Next()
			if tt == Error {
				switch t.Error() {
				case os.EOF:
					break
				default:
					log.Panicf(
						"Error tokenizing string: %s",
						t.Error())
				}
			}
			tokens <- t.Token()
		}
	}()
	return tokens
}

type Document struct {
	top *Node
}

func transformAttributes(attrs []Attribute) map[string] string {
	attributes := make(map[string] string)
	for _, attr := range attrs {
		attributes[attr.Key] = attr.Val
	}
	return attributes
}

func typeFromToken(t Token) NodeType {
	if t.Type == Text {
		return TEXT
	}
	return TAG
}

func nodeFromToken(t Token) *Node {
	return &Node{
		nodeType: typeFromToken(t),
		nodeValue: t.Data,
		nodeAttributes: transformAttributes(t.Attr),
	}
}

func NewDoc(s string) *Document {
	t := NewTokenizer(strings.NewReader(s))
	tokens := lazyTokens(t)
	tok1 := <-tokens
	doc := Document{top: nodeFromToken(tok1)}

	queue := new(v.Vector)
	queue.Push(doc.top)
	for tok := range tokens {
		curr := queue.At(0).(Node)
		switch tok.Type {
		case SelfClosingTag, Text:
			curr.children.Push(nodeFromToken(tok))
		case StartTag:
			curr.children.Push(nodeFromToken(tok))
			queue.Push(nodeFromToken(tok))
		case EndTag:
			queue.Pop()
		}
	}
	return &doc
}
