// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.

package transform

import (
	"fmt"
	"io"

	"golang.org/x/net/html"

	"go.marzhillstudios.com/pkg/go-html-transform/css/selector"
	"go.marzhillstudios.com/pkg/go-html-transform/h5"
)

// Collector defines an interface for html node collectors.
type Collector interface {
	// Find searches a tree rooted at n and returns a slice of nodes
	// that match a criteria.
	Find(n *html.Node) []*html.Node
}

type CollectorFunc func(n *html.Node) []*html.Node

func (f CollectorFunc) Find(n *html.Node) []*html.Node {
	return f(n)
}

func FirstMatch(cs ...Collector) CollectorFunc {
	return func(n *html.Node) []*html.Node {
		for _, col := range cs {
			if ns := col.Find(n); ns != nil {
				return ns
			}
		}
		return nil
	}
}

// The TransformFunc type is the type of a html.Node transformation function.
type TransformFunc func(*html.Node)

// Transformer encapsulates a document under transformation.
type Transformer struct {
	doc *h5.Tree
}

func NewFromReader(rdr io.Reader) (*Transformer, error) {
	tree, err := h5.New(rdr)
	if err == nil {
		return New(tree), nil
	}
	return nil, err
}

// Constructor for a Transformer. It makes a copy of the document
// and transforms that instead of the original.
func New(t *h5.Tree) *Transformer {
	clone := t.Clone()
	return newTransformer(&clone)
}

func newTransformer(t *h5.Tree) *Transformer {
	return &Transformer{doc: t}
}

// The Doc method returns the document under transformation.
func (t *Transformer) Doc() *html.Node {
	return t.doc.Top()
}

func (t *Transformer) Render(w io.Writer) error {
	return t.doc.Render(w)
}

func (t *Transformer) String() string {
	return t.doc.String()
}

func (t *Transformer) Clone() *Transformer {
	return New(t.doc)
}

func applyFuncToCollector(f TransformFunc, n *html.Node, sel Collector) {
	for _, nn := range sel.Find(n) {
		f(nn)
	}
}

// The ApplyWithSelector method applies a TransformFunc to the nodes matched
// by the CSS3 Selector.
func (t *Transformer) Apply(f TransformFunc, sel string) error {
	sq, err := selector.Selector(sel)
	t.ApplyWithCollector(f, sq)
	return err
}

func (t *Transformer) ApplyToFirstMatch(f TransformFunc, sels ...string) error {
	cs := make([]Collector, 0, len(sels))
	for _, sel := range sels {
		sq, err := selector.Selector(sel)
		if err != nil {
			return err
		}
		cs = append(cs, sq)
	}
	t.ApplyWithCollector(f, FirstMatch(cs...))
	return nil
}

// ApplyWithCollector applies a TransformFunc to the tree using a Collector.
func (t *Transformer) ApplyWithCollector(f TransformFunc, coll Collector) {
	// TODO come up with a way to walk tree once?
	applyFuncToCollector(f, t.Doc(), coll)
}

// Transform is a bundle of selectors and a transform func. It forms a
// self contained Transfrom on an html document that can be reused.
type Transform struct {
	coll Collector
	f    TransformFunc
}

// Trans creates a Transform that you can apply using ApplyAll.
// It takes a TransformFunc and a valid CSS3 Selector.
// It returns a *Transform or an error if the selector wasn't valid
func Trans(f TransformFunc, sel string) (*Transform, error) {
	sq, err := selector.Selector(sel)
	return TransCollector(f, sq), err
}

// MustTrans creates a Transform.
// Panics if the selector wasn't valid.
func MustTrans(f TransformFunc, sel string) *Transform {
	t, err := Trans(f, sel)
	if err != nil {
		panic(err)
	}
	return t
}

// TransCollector creates a Transform that you can apply using ApplyAll.
// It takes a TransformFunc and a Collector
func TransCollector(f TransformFunc, coll Collector) *Transform {
	return &Transform{f: f, coll: coll}
}

// ApplyAll applies a series of Transforms to a document.
//     t.ApplyAll(Trans(f, sel1, sel2), Trans(f2, sel3, sel4))
func (t *Transformer) ApplyAll(ts ...*Transform) {
	for _, spec := range ts {
		t.ApplyWithCollector(spec.f, spec.coll)
	}
}

// AppendChildren creates a TransformFunc that appends the Children passed in.
func AppendChildren(cs ...*html.Node) TransformFunc {
	return func(n *html.Node) {
		for _, c := range cs {
			if c.Parent != nil {
				c.Parent.RemoveChild(c)
			}
			n.AppendChild(h5.CloneNode(c))
		}
	}
}

// PrependChildren creates a TransformFunc that prepends the Children passed in.
func PrependChildren(cs ...*html.Node) TransformFunc {
	return func(n *html.Node) {
		for _, c := range cs {
			n.InsertBefore(h5.CloneNode(c), n.FirstChild)
		}
	}
}

// RemoveChildren creates a TransformFunc that removes the Children of the node
// it operates on.
func RemoveChildren() TransformFunc {
	return func(n *html.Node) {
		removeChildren(n)
	}
}

func removeChildren(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		defer n.RemoveChild(c)
	}
}

// ReplaceChildren creates a TransformFunc that replaces the Children of the
// node it operates on with the Children passed in.
func ReplaceChildren(ns ...*html.Node) TransformFunc {
	return func(n *html.Node) {
		removeChildren(n)
		for _, c := range ns {
			n.AppendChild(h5.CloneNode(c))
		}
	}
}

func nodeToString(n *html.Node) string {
	t := h5.NewTree(n)
	return t.String()
}

// Replace constructs a TransformFunc that replaces a node with the nodes passed
// in.
func Replace(ns ...*html.Node) TransformFunc {
	return func(n *html.Node) {
		p := n.Parent
		switch p {
		case nil:
			panic(fmt.Sprintf("Attempt to replace Root node: %s", h5.RenderNodesToString([]*html.Node{n})))
		default:
			for _, nc := range ns {
				p.InsertBefore(h5.CloneNode(nc), n)
			}
			p.RemoveChild(n)
		}
	}
}

// DoAll returns a TransformFunc that combines all the TransformFuncs that are
// passed in. Doing each transform in order.
func DoAll(fs ...TransformFunc) TransformFunc {
	return func(n *html.Node) {
		for _, f := range fs {
			f(n)
		}
	}
}

// CopyAnd will construct a TransformFunc that will
// make a copy of the node for each passed in TransformFunc
// and replace the passed in node with the resulting transformed
// html.Nodes.
func CopyAnd(fns ...TransformFunc) TransformFunc {
	return func(n *html.Node) {
		for _, fn := range fns {
			node := h5.CloneNode(n)
			n.Parent.InsertBefore(node, n)
			fn(node)
		}
		n.Parent.RemoveChild(n)
	}
}

// SubTransform constructs a TransformFunc that runs a TransformFunc on
// any nodes in the tree rooted by the node the the TransformFunc is run
// against.
// This is useful for creating self contained Transforms that are
// meant to work on subtrees of the html document.
func Subtransform(f TransformFunc, sel string) (TransformFunc, error) {
	sq, err := selector.Selector(sel)
	return SubtransformCollector(f, sq), err
}

// MustSubtransform constructs a TransformFunc that runs a TransformFunc on
// any nodes in the tree rooted by the node the the TransformFunc is run
// against.
// Panics if the selector string is malformed.
func MustSubtransform(f TransformFunc, sel string) TransformFunc {
	t, err := Subtransform(f, sel)
	if err != nil {
		panic(err)
	}
	return t
}

// SubTransformSelector constructs a TransformFunc that runs a TransformFunc on
// any nodes collected, using the passed in collector, from the subtree the
// TransformFunc is run on.
// This is useful for creating self contained Transforms that are
// meant to work on subtrees of the html document.
func SubtransformCollector(f TransformFunc, coll Collector) TransformFunc {
	return func(n *html.Node) {
		applyFuncToCollector(f, n, coll)
	}
}

// ModifyAttrb creates a TransformFunc that modifies the attributes
// of the node it operates on. If an Attribute with the same name
// as the key doesn't exist it creates it.
func ModifyAttrib(key string, val string) TransformFunc {
	return func(n *html.Node) {
		found := false
		for i, attr := range n.Attr {
			if attr.Key == key {
				n.Attr[i].Val = val
				found = true
			}
		}
		if !found {
			n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
		}
	}
}

// TransformAttrib returns a TransformFunc that transforms an attribute on
// the node it operates on using the provided func. It only transforms
// the attribute if it exists.
func TransformAttrib(key string, f func(string) string) TransformFunc {
	return func(n *html.Node) {
		for i, attr := range n.Attr {
			if attr.Key == key {
				n.Attr[i].Val = f(n.Attr[i].Val)
			}
		}
	}
}

// Trace is a debugging wrapper for transform funcs.
// It calls traceFunc with debugging information before and after the
// TransformFunc is applied.
func Trace(f TransformFunc, traceFunc func(msg string, args ...interface{}), msg string, args ...interface{}) TransformFunc {
	return func(n *html.Node) {
		traceFunc(msg, args...)
		p := n.Parent
		if p == nil {
			p = n
		}
		traceFunc("Before: %s", h5.NewTree(p).String())
		f(n)
		traceFunc("After: %s", h5.NewTree(p).String())
	}
}
