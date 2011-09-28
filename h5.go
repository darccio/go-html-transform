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

type NodeType int
const (
	TextNode NodeType = iota // zero value so the default
	ElementNode NodeType = iota
)

type Node struct {
	Type NodeType
	data []int
	Attr []*Attribute
	Parent *Node
	Children []*Node
}

func (n *Node) Data() string {
	return string(n.data)
}

type TokenConsumer func(*Parser, []int)

type Parser struct {
	In *bufio.Reader
	Top *Node
	curr *Node
	consumer TokenConsumer
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

// TODO(jwall): UNITTESTS!!!!
func (p *Parser) Parse() os.Error {
	// we start in the data state
	h := handleChar(dataStateHandler)
	for h != nil {
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
		return h(p, c), nil
	}
	return memoized[h]
}

// TODO(jwall): UNITTESTS!!!!
// Section 11.2.4.1
func dataStateHandler(p *Parser, c int) stateHandler {
	switch c {
	//case '&': // TODO(jwall): do we actually care for this parser?
		//return handleChar(charRefHandler)
	case '<':
		return handleChar(tagOpenHandler)
	default:
		// consume the token
		textConsumer(p, c)
		return handleChar(dataStateHandler)
	}
	panic("Unreachable")
}

// TODO(jwall):
// Section 11.2.4.2
func charRefHandler(p *Parser, c int) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ', '<', '&':
		// TODO
	case '#':
		// TODO
	default:
		// TODO
	}
	panic("Unreachable")
}

// Section 11.2.4.8
func tagOpenHandler(p *Parser, c int) stateHandler {
	curr := pushNode(p)
	switch c {
	case '!': // markup declaration state
		// TODO
	case '/': // end tag open state
		return endTagOpenHandler
		// TODO
	case '?': // TODO parse error // bogus comment state
		return bogusCommentHandler
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		curr.Type = ElementNode
		lc := c + 0x0020 // lowercase it
		curr.data = []int{lc}
		return handleChar(tagNameHandler)
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
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
		return handleChar(selfClosingStartTagHandler)
	case '>':
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
		return handleChar(selfClosingStartTagHandler)
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
		return handleChar(selfClosingStartTagHandler)
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
	// case '&':
		// TODO do we even care for this parser?
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
		//case '&':
			// TODO do we even care for this parser?
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
	//case '&':
		// TODO do we even care for this parser?
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
		return handleChar(selfClosingStartTagHandler)
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
		return handleChar(selfClosingStartTagHandler)
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
func selfClosingStartTagHandler(p *Parser, c int) stateHandler {
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
	p.curr = p.curr.Parent
	return p.curr
}
