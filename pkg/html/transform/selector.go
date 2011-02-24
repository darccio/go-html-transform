/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	. "html"
	v "container/vector"
	"log"
	s "strings"
)

type SelectorQuery struct {
	*v.Vector
}

type SelectorPart struct {
	Type    byte // a bitmask of the selector types 
	Val     string // the value we are matching against
}

type Selector struct {
	TagName string // "*" for any tag otherwise the name of the tag
	Parts []SelectorPart
	Attr map[string]string
}

const (
	TAGNAME byte = iota // zero value so the default
	CLASS   byte = '.'
	ID      byte = '#'
	PSEUDO  byte = ':'
	ANY     byte = '*'
	ATTR    byte = '['
)

const (
	SELECTOR_CHARS string = ".:#["
)

func matchAttrib(nodeAttr []Attribute, matchAttr map[string]string) bool {
	attribResult := true
	for key, val := range matchAttr {
		exists := false
		matched := false
		for _, attr := range nodeAttr {
			if key == attr.Key {
				exists = true
				if val == attr.Val {
					matched = true	
				}
				attribResult = attribResult && exists && matched
			}
		}
		attribResult = attribResult && exists
	}
	return attribResult
}

func (part SelectorPart) match(node *Node) bool {
	switch part.Type {
		case CLASS:
			classAttr := make(map[string]string)
			classAttr["class"] = part.Val
			return matchAttrib(node.Attr, classAttr) 
		case ID:
			idAttr := make(map[string]string)
			idAttr["id"] = part.Val
			return matchAttrib(node.Attr, idAttr)
		case PSEUDO:
	}
	return false
}

func (sel *Selector) Match(node *Node) bool {
	tagNameResult := true
	if sel.TagName != "" && sel.TagName != "*" && sel.TagName != node.Data {
		tagNameResult = tagNameResult && false
	}
	attribResult := matchAttrib(node.Attr, sel.Attr)
	// TODO(jwall): hook in the whole part matching
	return tagNameResult && attribResult
}

func newAnyTagClassOrIdSelector(str string) *Selector {
	return &Selector{
	Parts:    []SelectorPart{
			SelectorPart{
			Type: str[0],
			Val:     str[1:],
		}},
	TagName: "*",
	}
}

func newAnyTagSelector(str string) *Selector {
	return &Selector{
		TagName: "*",
	}
}

func splitAttrs(str string) []string {
	attrs := s.FieldsFunc(str[1:len(str)-1], func(c int) bool {
		if c == '=' {
			return true
		}
		return false
	})
	return attrs
}

func newAnyTagAttrSelector(str string) *Selector {
	parts := s.SplitAfter(str, "]", -1)
	sel := Selector{
		TagName: "*",
		Attr:    map[string]string{},
	}
	for _, part := range parts[0:len(parts)-1] {
		attrs := splitAttrs(part)
		sel.Attr[attrs[0]] = attrs[1]
	}
	return &sel
}

func newTagNameSelector(str string) *Selector {
	return &Selector{
		TagName: str,
	}
}

func newTagNameWithConstraints(str string, i int) *Selector {
	// TODO(jwall): indexAny use [CLASS,...]
	var selector = new(Selector)
	switch str[i] {
	case CLASS, ID, PSEUDO: // with class or id
		selector = newAnyTagClassOrIdSelector(str[i:])
	case ATTR: // with attribute
		selector = newAnyTagAttrSelector(str[i:])
	default:
		panic("Invalid constraint type for the tagname selector")
	}
	selector.TagName = str[0:i]
	//selector.Type = TAGNAME
	return selector
}

func partition(s string, f func(c int) bool) []string {
	// TODO(jwall): make this more efficient
	parts := []string{}
	start := 0
	for i, char := range s {
		if i < 1 { // we don't want empty first partitions
			continue
		}
		if f(char) {
			parts = append(parts, s[start:i])
			start = i
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func MergeSelectors(sel1 *Selector, sel2 *Selector) {
	if sel2.TagName != "" && sel2.TagName != "*" {
		if sel1.TagName == "" || sel1.TagName == "*" {
			sel1.TagName = sel2.TagName
		} else {
			log.Panicf("Can't merge multiple tagnames in" +
				" selectors: First [%s] Second [%s]",
				sel1.TagName, sel2.TagName)
		}
	}
	if sel1.Parts == nil {
		sel1.Parts = make([]SelectorPart, 0)
	}
	newParts := make([]SelectorPart, len(sel1.Parts) + len(sel2.Parts))
	copy(newParts, sel1.Parts)
	copy(newParts[len(sel1.Parts):], sel2.Parts)
	sel1.Parts = newParts
	if sel1.Attr == nil {
		sel1.Attr = make(map[string]string)
	}
	for key, val := range sel2.Attr {
		sel1.Attr[key] = val
	}
}

func NewSelector(str string) *Selector {
	str = s.TrimSpace(str) // trim whitespace
	// TODO(jwall): support combinators > + \S
	// TODO(jwall): splitAfter one of ".:#["
	parts := partition(str, func(c int) bool {
		for _, c2 := range SELECTOR_CHARS {
			if c == c2 {
				return true
			}
		}
		return false
	})
	result := new(Selector)
	result.TagName = "*"
	for _, p := range parts {
		selector := new(Selector)
		switch p[0] {
		case CLASS, ID: // Any tagname with class or id
			selector = newAnyTagClassOrIdSelector(p)
		case ANY: // Any tagname
			selector = newAnyTagSelector(p)
		case ATTR: // any tagname with attribute
			// TODO(jwall): expanded attribute selectors
			selector = newAnyTagAttrSelector(p)
		default: // TAGNAME
			// TODO(jwall): indexAny use [CLASS,...]
			if i := s.IndexAny(p, SELECTOR_CHARS); i != -1 {
				selector = newTagNameWithConstraints(p, i)
			} else { // just a tagname
				selector = newTagNameSelector(p)
			}
		}
		MergeSelectors(result, selector)
	}
	return result
}

func NewSelectorQuery(sel ...string) *SelectorQuery {
	q := SelectorQuery{Vector: new(v.Vector)}
	for _, str := range sel {
		selPart := NewSelector(str)
		if selPart == nil {
			log.Panic("Invalid Selector in query")
		}
		q.Push(selPart)
	}
	return &q
}

func testSelectorAttrs(attrs []Attribute, sel *Selector) bool {
	result := false
	for key, val := range sel.Attr {
		result = result || testAttr(attrs, key, val)
	}
	return result
}

func testAttr(attrs []Attribute, key string, val string) bool {
	for _, attr := range attrs {
		// TODO(jwall): we need to handle the multiple match types
		// [att] [att=val] [att~=val] [att|=val]?
		if attr.Key == key && attr.Val == val {
			return true
		}
	}
	return false
}

func testNode(node *Node, sel Selector) bool {
	/*
	if sel.TagName == "*" {
		attrs := node.Attr
		switch sel.Type {
		case ID:
			if testAttr(attrs, "id",  sel.Val) {
				return true
			}
		case CLASS:
			if testAttr(attrs, "class", sel.Val) {
				return true
			}
		case ATTR:
			if testSelectorAttrs(attrs, &sel) {
				return true
			}
			//case PSEUDO:
			//TODO(jwall): implement these
		}
	} else {
		if node.Data == sel.TagName {
			attrs := node.Attr
			switch sel.Type {
			case ID:
				if testAttr(attrs, "id", sel.Val) {
					return true
				}
			case CLASS:
				if testAttr(attrs, "class", sel.Val) {
					return true
				}
			case ATTR:
				if testSelectorAttrs(attrs, &sel) {
					return true
				}
			//case PSEUDO:
			//TODO(jwall): implement these
			default:
				return true
			}
		}
	}
	*/
	return false
}

/*
 Apply the css selector to a document.

 Returns a Vector of nodes that the selector matched.
*/
func (sel *SelectorQuery) Apply(doc *Document) *v.Vector {
	interesting := new(v.Vector)
	return interesting
}

/*
 Replace each node the selector matches with the passed in node.

 Applies the selector against the doc and replaces the returned
 Nodes with the passed in n HtmlNode.
*/
func (sel *SelectorQuery) Replace(doc *Document, n *Node) {
	nv := sel.Apply(doc)
	for i := 0; i <= nv.Len(); i++ {
		// Change to take into account new usage of *Node
		//nv.At(i).(*Node).Copy(n)
	}
	return
}
