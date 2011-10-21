package h5

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"strings"
)

type ParseError struct {
	msg string
	node *Node
}

func NewParseError(n *Node, msg string, args... interface{}) *ParseError {
	return &ParseError{node:n, msg:fmt.Sprintf(msg, args...)}
}

func (e ParseError) String() string {
	return e.msg
}

type Attribute struct {
	Name string
	Value string
}

func (a *Attribute) String() string {
	return " " + a.Name + "='" + a.Value + "'"
}

type NodeType int
const (
	TextNode NodeType = iota // zero value so the default
	ElementNode NodeType = iota
	DoctypeNode NodeType = iota
	CommentNode NodeType = iota
)

type Node struct {
	Type NodeType
	data []int
	Attr []*Attribute
	Parent *Node
	Children []*Node
}

func attrString(attrs []*Attribute) string {
	if attrs == nil {
		return ""
	}
	s := ""
	for _, a := range attrs {
		s += fmt.Sprintf(" %s", a)
	}
	return s
}

func (n *Node) String() string {
	switch n.Type {
	case TextNode:
		return n.Data()
	case ElementNode, DoctypeNode:
		s :="<" + n.Data() + attrString(n.Attr) + ">"
		for _, n := range n.Children {
			s += n.String()
		}
		s += "</" + n.Data() + ">"
		return s
	case CommentNode:
		// TODO
	}
	return ""
}

func (n *Node) Data() string {
	if n.data != nil {
		return string(n.data)
	}
	return ""
}

type TokenConsumer func(*Parser, []int)

type InsertionMode int

const (
	IM_initial InsertionMode = iota
	IM_beforeHtml InsertionMode = iota
	IM_beforeHead InsertionMode = iota
	IM_inHead InsertionMode = iota
	IM_inHeadNoScript InsertionMode = iota
	IM_afterHead InsertionMode = iota
	IM_inBody InsertionMode = iota
	IM_text InsertionMode = iota
	IM_inTable InsertionMode = iota
	IM_inTableText InsertionMode = iota
	IM_inCaption InsertionMode = iota
	IM_inColumnGroup InsertionMode = iota
	IM_inTableBody InsertionMode = iota
	IM_inRow InsertionMode = iota
	IM_inCell InsertionMode = iota
	IM_inSelect InsertionMode = iota
	IM_inSelectInTable InsertionMode = iota
	IM_afterBody InsertionMode = iota
	IM_afterFrameset InsertionMode = iota
	IM_afterAfterBody InsertionMode = iota
	IM_afterAfterFrameset InsertionMode = iota
)

func insertionModeSwitch(p *Parser, n *Node) stateHandler {
	currMode := p.Mode
	switch currMode {
	case IM_initial:
		fallthrough
	case IM_beforeHtml:
		fallthrough
	case IM_beforeHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "head":
				p.Mode = IM_inHead
			default:
				// TODO(jwall): parse error
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "script":
				p.Mode = IM_text
				return handleChar(scriptDataStateHandler)
			default:
				// TODO(jwall): parse error
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inHeadNoScript:
	case IM_afterHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "body":
				p.Mode = IM_inBody
			default:
				// TODO(jwall): parse error
				// inject a body tag first?
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inTable:
		fallthrough
	case IM_inTableText:
		fallthrough
	case IM_inCaption:
		fallthrough
	case IM_inColumnGroup:
		fallthrough
	case IM_inTableBody:
		fallthrough
	case IM_inRow:
		fallthrough
	case IM_inCell:
		fallthrough
	case IM_inSelect:
		fallthrough
	case IM_inSelectInTable:
		fallthrough
	case IM_afterBody:
		fallthrough
	case IM_inBody:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "script":
				p.Mode = IM_text
				return handleChar(scriptDataStateHandler)
			default:
				// TODO(jwall): parse error
			}
		}
	case IM_afterFrameset:
		fallthrough
	case IM_afterAfterFrameset:
		fallthrough
	case IM_afterAfterBody:
		fallthrough
			// TODO(jwall): parse error
	case IM_text:
			if n.Data() == "script" {
				p.Mode = IM_inBody
			}
	}
	return handleChar(dataStateHandler)
}

func dataStateHandlerSwitch(p *Parser) stateHandler {
	n := p.curr
	if n != nil {
		return insertionModeSwitch(p, n)
	}
	return handleChar(dataStateHandler)
}

type Parser struct {
	In *bufio.Reader
	Top *Node
	curr *Node
	consumer TokenConsumer
	Mode InsertionMode
	buf []int // temporary buffer
}

// Handles the various tokenization states
type stateHandler func(p *Parser) (stateHandler, os.Error)

func NewParserFromString(s string) *Parser {
	return NewParser(strings.NewReader(s))
}

func NewParser(r io.Reader) *Parser {
	return &Parser{In: bufio.NewReader(r)}
}

func (p *Parser) nextInput() (int, os.Error) {
	r, _, err := p.In.ReadRune()
	return r, err
}

func (p *Parser) Parse() os.Error {
	// we start in the Doctype state
	// and in the Initial InsertionMode
	h := dataStateHandlerSwitch(p)
	for h != nil {
		//if p.curr != nil && p.curr.data != nil {
			//fmt.Printf("YYY: %v\n", p.curr.Data())
		//}
		h2, err := h(p)
		if err == os.EOF {
			return nil
		}
		if err != nil {
			// TODO parse error
			return os.NewError(fmt.Sprintf("Parse error: %s", err))
		}
		h = h2
	}
	return nil
}

// TODO(jwall): UNITTESTS!!!!
func textConsumer(p *Parser, chars... int) {
	if p.curr == nil {
		pushNode(p)
	}
	p.curr.data = append(p.curr.data, chars...) // ugly but safer
}

var memoized = make(map[func(*Parser, int) stateHandler]stateHandler)

// TODO(jwall): UNITTESTS!!!!
func handleChar(h func(*Parser, int) stateHandler) stateHandler {
	if f, ok := memoized[h]; ok {
		return f
	}
	memoized[h] = func(p *Parser) (stateHandler, os.Error) {
		c, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		//fmt.Printf("YYY: char %c\n", c)
		return h(p, c), nil
	}
	return memoized[h]
}

// Section 11.2.4.52
func doctypeStateHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeDoctypeHandler)
	default:
		// TODO parse error
		return handleChar(beforeDoctypeHandler)
	}
	panic("unreachable")
}

// Section 11.2.4.53
func beforeDoctypeHandler(p *Parser, c int) stateHandler {
	curr := pushNode(p)
	curr.Type = DoctypeNode
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeDoctypeHandler)
	case '>':
		// TODO parse error, quirks mode
		return dataStateHandlerSwitch(p)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		curr.data = append(curr.data, lc)
		return handleChar(doctypeNameState)
	default:
		curr.data = append(curr.data, c)
		return handleChar(doctypeNameState)
	}
	panic("unreachable")
}

// Section 11.2.4.54
func doctypeNameState(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return afterDoctypeNameHandler
	case '>':
		return afterDoctypeNameHandler
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		n.data = append(n.data, lc)
		return handleChar(doctypeNameState)
	default:
		n.data = append(n.data, c)
		return handleChar(doctypeNameState)
	}
	panic("unreachable")
}

var (
	PUBLIC = "public"
	SYSTEM = "system"
)

// Section 11.2.4.55
func afterDoctypeNameHandler(p *Parser) (stateHandler, os.Error) {
	firstSix := make([]int, 0, 6)
	//n := p.curr
	for {
		c, err := p.nextInput()
		if err == os.EOF {
			// TODO parse error
			return dataStateHandlerSwitch(p), nil
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			// ignore
			return afterDoctypeNameHandler, nil
		case '>':
			return dataStateHandlerSwitch(p), nil
		default:
			if len(firstSix) == cap(firstSix) {
				switch string(firstSix) {
				case PUBLIC:
					return handleChar(afterDoctypeHandler), nil
				case SYSTEM:
					return handleChar(afterDoctypeHandler), nil
				}
			} else {
				lc := c + 0x0020 // lowercase it
				firstSix = append(firstSix, lc)
			}
		}
	}
	panic("unreachable")
}

// Section 11.2.4.56
func afterDoctypeHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeDoctypeIdentHandler)
	case '"', '\'':
		// TODO parse error
		return handleChar(makeIdentQuotedHandler(c))
	case '>':
		// TODO parse error
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO bogusDoctypeState
	}
	panic("unreachable")
}

// Section 11.2.4.57
func beforeDoctypeIdentHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeDoctypeIdentHandler)
	case '"', '\'':
		return handleChar(makeIdentQuotedHandler(c))
	case '>':
		 // TODO parse error
		 return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO bogusDoctypeState
	}
	panic("unreachable")
}

// Section 11.2.4.58
func makeIdentQuotedHandler(q int) (func(*Parser, int) stateHandler) {
	return func(p *Parser, c int) stateHandler {
		if q == c {
			return handleChar(afterDoctypeIdentifierHandler)
		}
		if c == '>' {
			// TODO parse error
			return dataStateHandlerSwitch(p)
		}
		panic("unreachable")
	}
	panic("unreachable")
}

// section 11.2.4.59
func afterDoctypeIdentifierHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(afterDoctypeIdentifierHandler)
    case '>':
		p.Mode = IM_beforeHtml
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO switch to bogus Doctype state
	}
	panic("unreachable")
}

func scriptDataStateHandler(p *Parser, c int) stateHandler {
	pushNode(p) // push a text node onto the stack
	switch c {
	case '<':
		return handleChar(scriptDataLessThanHandler)
	default:
		// consume the token
		textConsumer(p, c)
		return handleChar(dataStateHandler)
	}
	panic("unreachable")
}

func scriptDataLessThanHandler(p *Parser, c int) stateHandler {
	switch c {
	case '/':
		p.buf = make([]int, 1)
		return handleChar(scriptDataEndTagOpenHandler)
	default:
		textConsumer(p, '<')
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

func scriptDataEndTagOpenHandler(p *Parser, c int) stateHandler {
	switch c {
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		p.buf = append(p.buf, lc)
		return handleChar(scriptDataEndTagNameHandler)
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.buf = append(p.buf, c)
		return handleChar(scriptDataEndTagNameHandler)
	default:
		textConsumer(p, '<', '/')
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

func scriptDataEndTagNameHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\f', '\n', ' ':
		if n.Data() == string(p.buf) {
			return handleChar(beforeAttributeNameHandler)
		} else {
			return handleChar(scriptDataStateHandler)
		}
	case '/':
		if n.Data() == string(p.buf) {
			return handleChar(selfClosingTagStartHandler)
		} else {
			return handleChar(scriptDataStateHandler)
		}
	case '>':
		if n.Data() == string(p.buf) {
			return dataStateHandlerSwitch(p)
		} else {
			return handleChar(scriptDataStateHandler)
		}
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		p.buf = append(p.buf, lc)
		return handleChar(scriptDataEndTagNameHandler)
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.buf = append(p.buf, c)
		return handleChar(scriptDataEndTagNameHandler)
	default:
		textConsumer(p, '<', '/')
		textConsumer(p, p.buf...)
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

// TODO(jwall): UNITTESTS!!!!
// Section 11.2.4.1
func dataStateHandler(p *Parser, c int) stateHandler {
	switch c {
	case '<':
		return handleChar(tagOpenHandler)
	default:
		// consume the token
		textConsumer(p, c)
		return handleChar(dataStateHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.8
func tagOpenHandler(p *Parser, c int) stateHandler {
	switch c {
	case '!': // markup declaration state
		// TODO
	case '/': // end tag open state
		//fmt.Printf("ZZZ: closing a tag\n")
		popNode(p)
		return endTagOpenHandler
	case '?': // TODO parse error // bogus comment state
		return bogusCommentHandler
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		//fmt.Printf("ZZZ: opening a new tag\n")
		curr := pushNode(p)
		curr.Type = ElementNode
		lc := c + 0x0020 // lowercase it
		curr.data = []int{lc}
		return handleChar(tagNameHandler)
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		//fmt.Printf("ZZZ: opening a new tag\n")
		curr := pushNode(p)
		curr.Type = ElementNode
		curr.data = []int{c}
		return handleChar(tagNameHandler)
	default: // parse error // recover using Section 11.2.4.8 rules
		// TODO
	}
	return nil
}

// Section 11.2.4.10
func tagNameHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		pushNode(p)
		return handleChar(dataStateHandler)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		n.data = append(n.data, lc)
		return handleChar(tagNameHandler)
	default:
		n.data = append(n.data, c)
		return handleChar(tagNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.34
func beforeAttributeNameHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		return handleChar(dataStateHandler)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		newAttr := new(Attribute)
		newAttr.Name = string(lc)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	case '=', '"', '\'', '<':
		// TODO parse error
		fallthrough
	default:
		newAttr := new(Attribute)
		newAttr.Name = string(c)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.35
func attributeNameHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(afterAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		return handleChar(dataStateHandler)
	case '=':
		return handleChar(beforeAttributeValueHandler)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Name += string(lc)
		return handleChar(attributeNameHandler)
	case '"', '\'', '<':
		// TODO parse error
		fallthrough
	default:	
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Name += string(c)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.37
func beforeAttributeValueHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeValueHandler)
	case '"', '\'':
		return handleChar(makeAttributeValueQuotedHandler(c))
	case '>':
		return handleChar(dataStateHandler)
	case '<', '=', '`':
		// todo parse error
		fallthrough
	default:
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Value += string(c)
		return handleChar(attributeValueUnquotedHandler)
	}	
	panic("Unreachable")
}

var memoizedQuotedAttributeHandlers = make(map[int]func(p *Parser, c int) stateHandler)
// Section 11.2.4.3{8,9}
func makeAttributeValueQuotedHandler(c int) (func(p *Parser, c int) stateHandler) {
	if memoizedQuotedAttributeHandlers[c] != nil {
		return memoizedQuotedAttributeHandlers[c]
	}
	f := func(p *Parser, c2 int) stateHandler {
		n := p.curr
		switch c2 {
		case '"', '\'':
			if c2 == c {
				return handleChar(afterAttributeValueQuotedHandler)
			}
			fallthrough
		default:
			currAttr := n.Attr[len(n.Attr)-1]
			currAttr.Value += string(c2)
			return handleChar(makeAttributeValueQuotedHandler(c))
		}
		panic("Unreachable")
	}
	memoizedQuotedAttributeHandlers[c] = f
	return f
}

// Section 11.2.4.40
func attributeValueUnquotedHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeNameHandler)
	case '>':
		return handleChar(dataStateHandler)
	case '"', '\'', '<', '=', '`':
		// TODO parse error
		fallthrough
	default:
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Value += string(c)
		return handleChar(attributeValueUnquotedHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.42
func afterAttributeValueQuotedHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		return handleChar(dataStateHandler)
	default:
		// TODO Parse error Reconsume the Character in the before attribute name state
		return handleChar(beforeAttributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.36
func afterAttributeNameHandler(p *Parser, c int) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(afterAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		return handleChar(dataStateHandler)
	case '=':
		return handleChar(beforeAttributeValueHandler)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		lc := c + 0x0020 // lowercase it
		newAttr := new(Attribute)
		newAttr.Name = string(lc)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	case '"', '\'', '<':
		// TODO parse error
		fallthrough
	default:
		newAttr := new(Attribute)
		newAttr.Name = string(c)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.43
func selfClosingTagStartHandler(p *Parser, c int) stateHandler {
	switch c {
		case '>':
			return handleChar(dataStateHandler)
		default:
			// TODO parse error reconsume as before attribute handler
			return handleChar(beforeAttributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.9
func endTagOpenHandler(p *Parser) (stateHandler, os.Error) {
	// compare to current tags name
	n := p.curr
	tag := make([]int, len(n.data))
	for i := 0; i <= len(n.data); i++ {
		c, err := p.nextInput()
		if err == os.EOF { // Parse Error
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		if i > len(n.data) {
				return nil, NewParseError(
					n, "End Tag does not match Start Tag start:[%s] end:[%s]",
					n.data, tag)
		}
		switch c {
		case '>':
			if i != len(n.data) {
				return nil, NewParseError(n, "End Tag Truncated: [%s]", tag)
			}
			if string(n.data) != string(tag) {
				return nil, NewParseError(
					n, "End Tag does not match Start Tag start:[%s] end:[%s]",
					n.data, tag)
			}
			popNode(p)
			return handleChar(dataStateHandler), nil
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			lc := c + 0x0020 // lowercase it
			if i == len(n.data) {
				return nil, NewParseError(
					n, "End Tag does not match Start Tag start:[%s] end:[%s]",
					n.data, tag)
			}
			tag[i] = lc
			if n.data[i] != lc {
				return nil, NewParseError(
					n, "End Tag does not match Start Tag start:[%s] end:[%s]",
					n.data, tag)
			}
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
			if i == len(n.data) {
				return nil, NewParseError(
					n, "End Tag does not match Start Tag start:[%s] end:[%s]",
					n.data, tag)
			}
			tag[i] = c
		default: // Bogus Comment state
			tag[i] = c
			return bogusCommentHandler, NewParseError(n,
				"Strange characters in end tag: [%s] switching to BogusCommentState", tag)
		}
	}
	panic("Unreachable")
}

// Section 11.2.4.44
func bogusCommentHandler(p *Parser) (stateHandler, os.Error) {
	n := addSibling(p)
	for {
		c, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		switch c {
		case '>':
			return handleChar(dataStateHandler), nil
		default:
			n.data = append(n.data, c)
		}
	}
	panic("Unreachable")
}

func addSibling(p *Parser) *Node {
	n := new(Node)
	p.curr.Parent.Children = append(p.curr.Parent.Children, n)
	return n
}

func pushNode(p *Parser) *Node {
	n := new(Node)
	if p.Top == nil {
		p.Top = n
	}
	if p.curr == nil {
		p.curr = n
	} else {
		n.Parent = p.curr
		n.Parent.Children = append(n.Parent.Children, n)
		p.curr = n
	}
	return n
}

func popNode(p *Parser) *Node {
	if p.curr != nil && p.curr.Parent != nil {
		p.curr = p.curr.Parent
	}
	return p.curr
}
