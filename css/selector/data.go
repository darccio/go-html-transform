// Package selector contains a css3 selector parser.
//
// The package follows the CSS3 Spec at: http://www.w3.org/TR/css3-selectors/
package selector

import (
	"go.marzhillstudios.com/pkg/go-html-transform/h5"

	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type selectorType int

const (
	// Tagname selector
	Tag selectorType = iota
	// Universal selector
	Universal
	// Id Selector
	Id
	// Class Selector
	Class
	// Attribute Selector
	Attr
	// Pseudoclass Selector
	PseudoClass
	// Pseudoelement Selector
	PseudoElement
)

// combinator combines two selector sequences together
type combinator int

const (
	// Descendant combinator
	Descendant combinator = iota
	// Child Combinatory
	Child
	// An immediately adjacent sibling combinator
	AdjacentSibling
	// General sibling combinator
	Sibling
)

func (t combinator) String() string {
	switch t {
	case Child:
		return ">"
	case Descendant:
		return " "
	case AdjacentSibling:
		return "+"
	case Sibling:
		return "~"
	}
	return ""
}

// The type of matcher for Attribute selectors
type attrMatchType int

const (
	// Test for the presence of the attribute
	Presence attrMatchType = iota
	// Test for an exact value for the attribute
	Exactly
	// Test that an attribute contains a value in a whitespace seperated list.
	Contains
	// Test that an attribute starts with a value or a value with a dash.
	DashPrefix
)

func (t attrMatchType) String() string {
	switch t {
	case Presence:
		return ""
	case Exactly:
		return "="
	case Contains:
		return "~="
	case DashPrefix:
		return "|="
	}
	panic("Unreachable")
}

// SimpleSelector describes one thing about an element.
type SimpleSelector struct {
	// The type of Simple Selector.
	Type selectorType
	// The Tagname of the SimpleSelector if Type is Tag
	Tag string
	// The attribute matching type if Type is Attr
	AttrMatch attrMatchType
	// The value to match against if Type is Id, Class, Pseudoclass, Pseudoelement, or Attr
	Value string
	// The attribute name if Type is Attr
	AttrName string
}

const (
	aMul = 100000000000000
	bMul = 100000000
)

func attrDashPrefix(prefix string, a *html.Attribute) bool {
	return a.Val == prefix || strings.HasPrefix(a.Val, prefix+"-")
}

func attrContains(val string, a *html.Attribute) bool {
	for _, v := range strings.Split(a.Val, " ") {
		if val == v {
			return true
		}
	}
	return false
}

func attrExactly(val string, a *html.Attribute) bool {
	return val == a.Val
}

// Match returns true if this SimpleSelector matches this node false otherwise.
func (ss SimpleSelector) Match(n *html.Node) bool {
	if n == nil {
		return false
	}
	if ss.Type == Tag {
		return strings.ToLower(ss.Tag) == strings.ToLower(h5.Data(n))
	}
	if ss.Type == PseudoClass {
		switch ss.Value {
		case "root":
			return n.Parent == nil
		case "first-child":
			return n.Parent != nil && n.Parent.FirstChild == n
		case "last-child":
			return n.Parent != nil && n.Parent.LastChild == n
		case "only-child":
			return n.PrevSibling == nil && n.NextSibling == nil
		case "empty":
			return n.FirstChild == nil
		default:
			// TODO(jwall):
			panic(fmt.Errorf("Can't match with PseudoClass %s", ss.Value))
		}
	} else if ss.Type == PseudoElement {
		panic(fmt.Errorf("Can't match with PseudoElement %s", ss.Value))
	}
	for _, a := range n.Attr {
		switch ss.Type {
		case Id:
			if strings.ToLower(a.Key) == "id" {
				return a.Val == ss.Value
			}
		case Class:
			if strings.ToLower(a.Key) == "class" {
				return attrContains(ss.Value, &a)
			}
		case Attr:
			if strings.ToLower(a.Key) == strings.ToLower(ss.AttrName) {
				switch ss.AttrMatch {
				case Exactly:
					return attrExactly(ss.Value, &a)
				case Contains:
					return attrContains(ss.Value, &a)
				case DashPrefix:
					return attrDashPrefix(ss.Value, &a)
				}
				return true
			}
		}
	}
	return false
}

// Specificity returns the CSS3 specificity for a SimpleSelector.
func (ss SimpleSelector) Specificity() int64 {
	switch ss.Type {
	case Id:
		return aMul
	case Class, Attr, PseudoClass:
		return bMul
	case Tag, PseudoElement:
		return 1
	}
	return 0
}

func (ss SimpleSelector) String() string {
	switch ss.Type {
	case Id:
		return "#" + ss.Value
	case Class:
		return "." + ss.Value
	case Attr:
		return "[" + ss.AttrName + ss.AttrMatch.String() + ss.Value + "]"
	case PseudoClass:
		return ":" + ss.Value
	case PseudoElement:
		return "::" + ss.Value
	case Universal:
		return "*"
	case Tag:
		return ss.Tag
	}
	panic("Unreachable")
}

// Sequence is a list of SimpleSelectors describing multiple things about an
// element.
type Sequence []SimpleSelector

// Find finds all Nodes that match this sequence in the tree rooted by
// n.
func (s Sequence) Find(n *html.Node) []*html.Node {
	var found []*html.Node
	h5.WalkNodes(n, func(n *html.Node) {
		if s.Match(n) {
			found = append(found, n)
		}
	})
	return found
}

// Match returns true if this Sequence matches this node false otherwise.
func (s Sequence) Match(n *html.Node) bool {
	if n == nil {
		return false
	}
	match := true
	for _, ss := range s {
		match = match && ss.Match(n)
	}
	return match
}

func (s Sequence) String() string {
	ss := ""
	for _, sel := range s {
		ss += sel.String()
	}
	return ss
}

// Specificity returns the CSS3 specificity for a given sequence of
// SimpleSelectors.
func (s Sequence) Specificity() int64 {
	var a, b, c int64 = 0, 0, 0
	for _, sel := range s {
		switch sel.Type {
		case Id:
			a++
		case Class, Attr, PseudoClass:
			b++
		case Tag, PseudoElement:
			c++
		}
	}
	return a*aMul + b*bMul + c
}

// Link joins a sequence to another sequence with a combinator.
type Link struct {
	// Combinator used to join this link with another Link or Sequence in a Chain.
	Combinator combinator
	// Sequence that combinator will join to the previous Link or Sequence in a Chain.
	Sequence
}

// Find all the nodes in a html.Node tree that match this Selector Link.
func (l Link) Find(n *html.Node) []*html.Node {
	var found []*html.Node
	switch l.Combinator {
	case Descendant:
		// walk the node tree returning any nodes the sequence matches
		h5.WalkNodes(n, func(n *html.Node) {
			if l.Sequence.Match(n) {
				found = append(found, n)
			}
		})
	case Child:
		// iterate through the children returning any nodes the sequence matches
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if l.Sequence.Match(c) {
				found = append(found, c)
			}
		}
	case AdjacentSibling:
		// look at the two adjacent siblings if any and return any that the squence matches.
		if l.Sequence.Match(n.PrevSibling) {
			found = append(found, n.PrevSibling)
		}
		if l.Sequence.Match(n.NextSibling) {
			found = append(found, n.NextSibling)
		}
	case Sibling:
		// Look at all the siblings if any and return any that the sequence matches.
		for s := n.PrevSibling; s != nil; s = s.PrevSibling {
			if l.Sequence.Match(s) {
				found = append([]*html.Node{s}, found...)
			}
		}
		for s := n.NextSibling; s != nil; s = s.NextSibling {
			if l.Sequence.Match(s) {
				found = append(found, s)
			}
		}
	}
	return found
}

func (l Link) String() string {
	return l.Combinator.String() + l.Sequence.String()
}

// Chain is a chain of sequences joined using Links.
type Chain struct {
	// The first Sequence of selectors in the Chain.
	Head Sequence
	// The rest of the Links in the chain.
	Tail []Link
}

// Find all the nodes in a html.Node tree that match this Selector Chain.
func (chn *Chain) Find(n *html.Node) []*html.Node {
	set := make(map[*html.Node]struct{})
	found := chn.Head.Find(n)
	for _, l := range chn.Tail {
		var interesting []*html.Node
		for _, n := range found {
			for _, n1 := range l.Find(n) {
				if _, ok := set[n1]; !ok {
					interesting = append(interesting, n1)
					set[n1] = struct{}{}
				}
			}
		}
		found = interesting
	}
	return found
}

func (chn *Chain) String() string {
	if chn == nil {
		return ""
	}
	ss := chn.Head.String()
	for _, l := range chn.Tail {
		ss += l.String()
	}
	return ss
}

// Specificity returns the CSS3 specificity of a Chain.
func (chn *Chain) Specificity() int64 {
	if chn == nil {
		return 0
	}
	sp := chn.Head.Specificity()
	for _, t := range chn.Tail {
		sp += t.Sequence.Specificity()
	}
	return sp
}
