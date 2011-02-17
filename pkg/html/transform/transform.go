/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.

 The html transform package implements a html css selector and transformer.

 An html doc can be inspected and queried using css selectors as well as
 transformed.

 	doc := Document{}
 	sel := NewSelector("a", ".foo")
 	node := sel.Apply(doc)

 	transformer := func(node Node) Node { ... }
 	Transform(doc, sel, transformer)
 	doc.ToString()
*/
package transform

import (
	v "container/vector"
)

func Transform(doc *Document, sel *SelectorQuery, f func(*v.Vector) *HtmlNode) {
	sel.Replace(doc, f(sel.Apply(doc)))
}

// TODO(jwall): helper transformation functions
