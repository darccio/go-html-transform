package css

import (
	"fmt"
	"io"
	"strings"

	//"code.google.com/p/go-html-transform/css/selector"
)

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Position
	Type   string
	String string
}

type PositionByteScanner struct {
	lastLineColumn int
	lastByte       byte
	Line           int
	Column         int
	io.ByteScanner
}

func (s *PositionByteScanner) ReadByte() (c byte, err error) {
	c, err = s.ByteScanner.ReadByte()
	if s.lastByte == '\n' {
		s.lastLineColumn = s.Column
		s.Line++
		s.Column = 0
	} else {
		s.Column++
	}
	s.lastByte = c
	return
}

func (s PositionByteScanner) UnreadByte() error {
	err := s.ByteScanner.UnreadByte()
	if err != nil {
		return err
	}
	if s.lastLineColumn == -1 {
		return fmt.Errorf("Calling UnreadByte again would result in loss of PositionInformation.")
	}
	if s.Column == 0 {
		s.lastByte = '\000'
		s.Line--
		s.Column = s.lastLineColumn
		s.lastLineColumn = -1
	} else {
		s.Column--
	}
	return nil
}

type cssParseFunc func(rdr *PositionByteScanner) (cssParseFunc, *Token, error)

func ParseReader(rdr *PositionByteScanner) (<-chan Token, error) {
	f := ParseContent
	ch := make(chan Token)
	go func() {
		var tk *Token
		var err error
		for f, tk, err = f(rdr); err != nil; f, tk, err = f(rdr) {
			if tk != nil {
				ch <- *tk
			}
		}
		if err != nil {
			// Handle the error usefully
		}
		close(ch)
	}()
	return ch, nil
}

func Parse(style string) (<-chan Token, error) {
	return ParseReader(
		&PositionByteScanner{ByteScanner: strings.NewReader(style)})
}

func ParseContent(rdr *PositionByteScanner) (cssParseFunc, *Token, error) {
	c, err := rdr.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	switch c {
	case ' ', '\n', '\t', '\r', '\f':
		// consume this and go on
	case '>':
		// comment?
		return consumeComment, nil, nil
	default:
		switch {
		case 'a' <= c && c >= 'z', 'A' <= c && c >= 'Z':
			return consumeSelector, nil, nil
		case c == '@':
			return consumeAtKeyword, nil, nil
		}
	}
	return nil, nil, fmt.Errorf("Invalid character %c at line %d column %d",
		c, rdr.Line, rdr.Column)
}

func consumeString(rdr *PositionByteScanner, expected string) error {
	return consumeBytes(rdr, []byte(expected))
}

func consumeBytes(rdr *PositionByteScanner, expected []byte) error {
	for _, b := range expected {
		c, err := rdr.ReadByte()
		if err != nil {
			return err
		}
		if c != b {
			return fmt.Errorf(
				"Unexpected character expected %q got %q at line %d column %d",
				b, c, rdr.Line, rdr.Column)
		}
	}
	return nil
}

func consumeComment(rdr *PositionByteScanner) (cssParseFunc, *Token, error) {
	c, err := rdr.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	if c == '!' {
		err := consumeString(rdr, "--")
		if err != nil {
			return nil, nil, err
		}
		return consumeCommentBody, nil, nil
	}
	return nil, nil, fmt.Errorf(
		"Invalid comment character %q at line %d column %d",
		c, rdr.Line, rdr.Column)
}

func consumeCommentBody(rdr *PositionByteScanner) {
	return nil, nil, nil
}

func consumeSelector(rdr *PositionByteScanner) (cssParseFunc, *Token, error) {
	c, err := rdr.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	return nil, nil, fmt.Errorf("Invalid character %v at line %d column %d",
		c, rdr.Line, rdr.Column)
}

func consumeAtKeyword(rdr *PositionByteScanner) (cssParseFunc, *Token, error) {
	c, err := rdr.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	return nil, nil, fmt.Errorf("Invalid character %q at line %d column %d",
		c, rdr.Line, rdr.Column)
}
