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

type HtmlNodeType int

const (
	TEXT HtmlNodeType = iota // 0 value so the default
	TAG
)

type HtmlNode struct {
	nodeType HtmlNodeType
	nodeValue string
	nodeAttributes map[string] string
	children v.Vector
}

func (n *HtmlNode) Copy(node HtmlNode) {
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
			if tt == ErrorToken {
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
	top *HtmlNode
}

func transformAttributes(attrs []Attribute) map[string] string {
	attributes := make(map[string] string)
	for _, attr := range attrs {
		attributes[attr.Key] = attr.Val
	}
	return attributes
}

func typeFromToken(t Token) HtmlNodeType {
	if t.Type == TextToken {
		return TEXT
	}
	return TAG
}

func nodeFromToken(t Token) *HtmlNode {
	return &HtmlNode{
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
		curr := queue.At(0).(HtmlNode)
		switch tok.Type {
		case SelfClosingTagToken, TextToken:
			curr.children.Push(nodeFromToken(tok))
		case StartTagToken:
			curr.children.Push(nodeFromToken(tok))
			queue.Push(nodeFromToken(tok))
		case EndTagToken:
			queue.Pop()
		}
	}
	return &doc
}
