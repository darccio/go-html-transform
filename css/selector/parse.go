package selector

import (
	"fmt"
	"io"
	"strings"
)

var (
	EOS = fmt.Errorf("End of Selector")
)

// SelectorFromScanner parses an io.ByteScanner into a Chain.
func SelectorFromScanner(rdr io.ByteScanner) (*Chain, error) {
	var chn Chain
	err := parseChain(rdr, &chn)
	if err != nil && err != io.EOF && err != EOS {
		return nil, err
	}
	return &chn, err
}

// Selector parses a string into a Chain.
// if it encounters a '{' character it stops and returns
// the Chain and EOS. This is a convenience that allows you to
// use this to parse selectors from a CSS file or style tag.
func Selector(sel string) (*Chain, error) {
	return SelectorFromScanner(strings.NewReader(sel))
}

func consumeValue(rdr io.ByteScanner) ([]byte, error) {
	bs := []byte{}
	for c, err := rdr.ReadByte(); err != io.EOF; c, err = rdr.ReadByte() {
		if err != nil {
			return nil, err
		}
		switch c {
		case '{':
			rdr.UnreadByte()
			return bs, EOS
		case '>', '+', '~', ' ', '\t', '\n', '\f', ',', '.', '#', '[', ':':
			rdr.UnreadByte()
			return bs, nil
		default:
			bs = append(bs, c)
		}
	}
	return bs, nil
}

func parseSimpleTag(rdr io.ByteScanner, sel *SimpleSelector) error {
	bs, err := consumeValue(rdr)
	if err != nil && err != EOS {
		return err
	}
	sel.Tag = sel.Tag + string(bs)
	return err
}

func parseSimpleSelector(rdr io.ByteScanner, sel *SimpleSelector) error {
	b, err := rdr.ReadByte()
	if err != nil && err != EOS {
		return err
	}
	bs, err := consumeValue(rdr)
	if err != nil && err != EOS {
		return err
	}
	bs = append([]byte{b}, bs...)
	if sel.Type == PseudoClass && bs[0] == ':' {
		sel.Type = PseudoElement
		bs = bs[1:]
	}
	sel.Value = string(bs)
	return err
}

func parseSimpleAttr(rdr io.ByteScanner, sel *SimpleSelector) error {
	var name []byte
	var value []byte
	var c1 byte = 0
	for c2, err := rdr.ReadByte(); err != io.EOF; c2, err = rdr.ReadByte() {
		if err != nil {
			return err
		}
		switch c2 {
		case ']':
			sel.AttrName = string(name)
			sel.Value = string(value)
			return nil
		case '=':
			if c1 == '~' {
				sel.AttrMatch = Contains
			} else if c1 == '|' {
				sel.AttrMatch = DashPrefix
			} else {
				sel.AttrMatch = Exactly
			}
		case '{':
			rdr.UnreadByte()
			return EOS
		case '~':
		case '|':
		// TODO(jwall): Substring matchers
		default:
			if sel.AttrMatch == Presence {
				name = append(name, c2)
			} else {
				value = append(value, c2)
			}
		}
		c1 = c2
	}
	return fmt.Errorf("Didn't close Attribute Matcher")
}

func parseSequence(rdr io.ByteScanner) (Sequence, error) {
	seq := []SimpleSelector{}
	rdr.UnreadByte()
	for c, err := rdr.ReadByte(); err != io.EOF; c, err = rdr.ReadByte() {
		if err != nil {
			return nil, err
		}
		switch c {
		case '*':
			seq = append(seq, SimpleSelector{Type: Universal})
		case '#':
			sel := SimpleSelector{Type: Id, AttrName: "id"}
			if err := parseSimpleSelector(rdr, &sel); err != nil {
				return nil, err
			}
			seq = append(seq, sel)
		case '.':
			sel := SimpleSelector{Type: Class, AttrName: "class"}
			if err := parseSimpleSelector(rdr, &sel); err != nil {
				return nil, err
			}
			seq = append(seq, sel)
		case ':':
			sel := SimpleSelector{Type: PseudoClass}
			if err := parseSimpleSelector(rdr, &sel); err != nil {
				return nil, err
			}
			seq = append(seq, sel)
		case '[':
			sel := SimpleSelector{Type: Attr}
			if err := parseSimpleAttr(rdr, &sel); err != nil {
				return nil, err
			}
			seq = append(seq, sel)
		case '{':
			rdr.UnreadByte()
			return seq, EOS
		case ' ', '\t', '\n', '\r', '\f', '>', '+', '~':
			rdr.UnreadByte()
			return seq, nil
		default:
			sel := SimpleSelector{Type: Tag, Tag: string(c)}
			if err := parseSimpleTag(rdr, &sel); err != nil {
				return nil, err
			}
			seq = append(seq, sel)
		}
	}
	return seq, nil
}

var combinatorMap = map[byte]combinator{
	'>': Child,
	'+': AdjacentSibling,
	'~': Sibling,
}

func parseCombinator(rdr io.ByteScanner, p *Link) error {
	rdr.UnreadByte()
	for c, err := rdr.ReadByte(); err != io.EOF; c, err = rdr.ReadByte() {
		if err != nil {
			return err
		}
		switch c {
		case '{':
			rdr.UnreadByte()
			return EOS
		case ',':
			return fmt.Errorf("Encountered ',' after combinator")
		case ' ', '\t', '\n', '\r', '\f':
		case '>', '+', '~':
			if p.Combinator == Descendant {
				p.Combinator = combinatorMap[c]
			} else {
				return fmt.Errorf("Can't combine multiple combinators")
			}
		default:
			rdr.UnreadByte()
			return nil
		}
	}
	return nil
}

func parseChain(rdr io.ByteScanner, chn *Chain) error {
	for c, err := rdr.ReadByte(); err != io.EOF; c, err = rdr.ReadByte() {
		if err != nil {
			return err
		}
		switch c {
		case ',':
			return fmt.Errorf("Parser does not handle groups")
		case ' ', '\t', '\n', '\r', '\f', '>', '+', '~':
			if chn.Head == nil {
				return fmt.Errorf("Starting selector chain with combinator %c", c)
			}
			part := Link{}
			if err := parseCombinator(rdr, &part); err != nil {
				return err
			}
			chn.Tail = append(chn.Tail, part)
		default:
			if chn.Head == nil {
				chn.Head, err = parseSequence(rdr)
				if err != nil && err != io.EOF {
					return err
				}
			} else {
				last := last(chn.Tail)
				if last != nil {
					last.Sequence, err = parseSequence(rdr)
					if err != nil && err != io.EOF {
						return err
					}
				} else {
					return fmt.Errorf(
						"Attempt to add tail seqence without combinator char: %c", c)
				}
			}
		}
	}
	return nil
}

// Utility function to return last link in a chain
func last(ls []Link) *Link {
	l := len(ls)
	if l == 0 {
		return nil
	}
	return &(ls[l-1])
}
