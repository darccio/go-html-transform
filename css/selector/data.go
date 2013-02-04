// package selector contains a css3 selector parser.
//
// The package follows the CSS3 Spec at: http://www.w3.org/TR/css3-selectors/
package selector

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
	Type      selectorType
	Tag       string
	AttrMatch attrMatchType
	Value     string
	AttrName  string
}

const (
	aMul = 100000000000000
	bMul = 100000000
)

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
	combinator
	Sequence
}

func (l Link) String() string {
	return l.combinator.String() + l.Sequence.String()
}

// Chain is a chain of Sequences joined by combinators.
type Chain struct {
	Head Sequence
	Tail []Link
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
