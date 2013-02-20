// Copyright 2011 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.

/*
Package h5 implements an wrapper for code.google.com/p/go/src/code.google.com/p/go.net/html.

    p := h5.NewParser(rdr)
    err := p.Parse()
    tree := p.Tree()

    tree.Walk(func(n *Node) {
       // do something with the node
    })

    tree2 := tree.Clone()
*/
package h5

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"io"
	"strings"
)

type Parser struct {
	Top *html.Node
}

func Partial(r io.Reader) ([]*html.Node, error) {
	b := &html.Node{}
	b.Data = "body"
	b.DataAtom = atom.Body
	b.Type = html.ElementNode
	return html.ParseFragment(r, b)
}

func PartialFromString(s string) ([]*html.Node, error) {
	return Partial(strings.NewReader(s))
}

func RenderNodes(w io.Writer, ns []*html.Node) error {
	for _, n := range ns {
		err := html.Render(w, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func RenderNodesToString(ns []*html.Node) string {
	buf := bytes.NewBufferString("")
	RenderNodes(buf, ns)
	return string(buf.Bytes())
}

// Construct a new h5 parser from a string
func NewParserFromString(s string) (*Parser, error) {
	return NewParser(strings.NewReader(s))
}

// Construct a new h5 parser from a io.Reader
func NewParser(r io.Reader) (*Parser, error) {
	n, err := html.Parse(r)
	return &Parser{Top: n}, err
}

// Tree returns the Top Node as a Tree.
func (p *Parser) Tree() Tree {
	return Tree{p.Top}
}

// TODO(jwall): Handle fragments.
