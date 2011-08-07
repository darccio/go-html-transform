package h5

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"strings"
)

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

func textConsumer(p *Parser, chars... int) {
	p.curr.data = append(p.curr.data, chars...) // ugly but safer
}

func handleChar(h func(*Parser, int) stateHandler) stateHandler {
	return func(p *Parser) (stateHandler, os.Error) {
			c, err := p.nextInput()
			if err != nil {
				return nil, err
			}
			return h(p, c), nil
		}
}

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
		lc := c + 0x0020 // lowercase it
		pushNode(p).data = []int{lc}
		return handleChar(tagNameHandler)
		// TODO // start new node with name set to lowercase c
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		// TODO
		pushNode(p).data = []int{c}
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
		// TODO beforeAttributeNameHandler
	case '/':
		// TODO selfClosingTagStartHandler
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

// Section 11.2.4.9
func endTagOpenHandler(p *Parser) (stateHandler, os.Error) {
	// compare to current tags name
	n := p.curr
	for i := 0; i < len(p.curr.data); i++ {
		c, err := p.nextInput()
		if err == os.EOF { // Parse Error
			// TODO
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		switch c {
		case '>':
			// TODO parse error
			return handleChar(dataStateHandler), nil
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			lc := c + 0x0020 // lowercase it
			if n.data[i] != lc {
				// TODO parse error
			} else {
				popNode(p)
			}
			return handleChar(dataStateHandler), nil
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
			if n.data[i] != c {
				// TODO parse error
			} else {
				popNode(p)
			}
			return handleChar(dataStateHandler), nil
		default: // Bogus Comment state
			return bogusCommentHandler, nil
		}
	}
	panic("unreachable")
}

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
			return handleChar(dataStateHandler), nil
		}
	}
	panic("Unreachable")
}

// TODO(jwall): UNITTESTS!!!!
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
