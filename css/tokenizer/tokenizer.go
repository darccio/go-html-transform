// package tokenizer tokenizes a css stream.
// Follows the spec defined at http://www.w3.org/TR/CSS21/syndata.html
package tokenizer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

type tokenType int

type Position struct {
	Line   int
	Column int
}

// TODO(jwall): Move this into some sort of utils?
type PositionTrackingScanner struct {
	// Embedded Scanner so our LineTrackingReader can be used just like
	// a Scanner.
	*bufio.Scanner
	lastL, lastCol int // Position Tracking fields.
	l, col         int
}

func NewTrackingReader(r io.Reader, splitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)) *PositionTrackingScanner {
	s := bufio.NewScanner(r)
	rdr := &PositionTrackingScanner{
		Scanner: s,
	}
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if rdr.l == 0 { // Initiliaze our position tracking fields using a split like a closure.
			rdr.l = 1
			rdr.col = 1
			rdr.lastL = 1
			rdr.lastCol = 1
		}
		advance, token, err = splitFunc(data, atEOF)
		// If we are advancing then we should see if there are any line
		// endings here
		if advance > 0 {
			rdr.lastCol = rdr.col
			rdr.lastL = rdr.l
			for _, b := range data[:advance] {
				if b == '\n' || atEOF { // Treat eof as a newline
					rdr.l++
					rdr.col = 1
				}
				rdr.col++
			}
		}
		return
	}
	s.Split(split)
	return rdr
}

func (l *PositionTrackingScanner) Position() Position {
	return Position{Line: l.l, Column: l.col}
}

type Tokenizer struct {
	p *PositionTrackingScanner
}

const (
	Ident        tokenType = iota // 0
	AtKeyword                     // 1
	String                        // 2
	BadString                     // 3
	BadUri                        // 4
	BadComment                    // 5
	Hash                          // 6
	Number                        // 7
	Percentage                    // 8
	Dimension                     // 9
	Uri                           // 10
	UnicodeRange                  // 11
	CDO                           // 12
	CDC                           // 13
	Colon                         // 14
	Semicolon                     // 15
	LBrace                        // 16
	RBrace                        // 17
	LParen                        // 18
	RParen                        // 19
	LBracket                      // 20
	RBracket                      // 21
	S                             // 22
	Includes                      // 23
	Dashmatch                     // 24
	Comment                       // 25
	Function                      // 26
	Delim                         // 27
)

type Token struct {
	Position
	Type   tokenType
	String string
}

func New(r io.Reader) *Tokenizer {
	return &Tokenizer{p: NewTrackingReader(r, splitFunc)}
}

func consumeWhitespace(data []byte) (advance int, token []byte, err error) {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\t', '\r', '\f':
		default:
			return i, data[:i], nil
		}
	}
	return len(data), data, nil
}

func consumeCdoOrDelim(data []byte) (advance int, token []byte, err error) {
	if len(data) >= 4 && string(data[:4]) == "<!--" {
		return 4, data[:4], nil
	}
	return 1, data[:1], nil
}

func consumeEscapedOrUnicode(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if data[0] != '\\' {
		return 0, nil, fmt.Errorf("Expected Escaped character got: %q", data)
	}
	c := data[1]
	switch {
	case '0' <= c && c <= '9':
		fallthrough
	case 'A' <= c && c <= 'F':
		fallthrough
	case 'a' <= c && c <= 'f':
		next, tok, err := consumeUnicode(data, atEOF)
		if err != nil {
			return 0, nil, err
		}
		return next, tok, err
	case c == '\n', c == '\r', c == '\f', c == '\t':
		return 0, nil, fmt.Errorf("Invalid escape, non escapable char %c", c)
	default:
		return 2, data[:2], nil
	}
}

func consumeEscapeOrUnicodeAnd(f bufio.SplitFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = consumeEscapedOrUnicode(data, atEOF)
		if err != nil {
			return 0, nil, err
		}
		log.Printf("Attempting to advance %d in `%s`", advance, data)
		next, tok, err := f(data[advance:], atEOF)
		if next == 0 {
			return 0, nil, nil
		}
		return advance + next, append(token, tok...), err
	}
}

func consumeUnicode(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if data[0] != '\\' {
		return 0, nil, fmt.Errorf("Not a unicode token!! %q", data)
	}
	for i, c := range data[1:] {
		switch {
		case '0' <= c && c <= '9':
			fallthrough
		case 'A' <= c && c <= 'F':
			fallthrough
		case 'a' <= c && c <= 'f':
			if i == 5 {
				return 6, data[:6], nil
			}
		default:
			return i + 1, data[:i+1], nil
		}
	}
	return 0, nil, nil
}

func consumeIdent(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, c := range data {
		// TODO(jwall): consumeIdentStart
		switch {
		case c == '\\':
			advance, token, err = consumeEscapeOrUnicodeAnd(consumeIdent)(data, atEOF)
			if advance > 0 {
				advance += i
				token = append(data[:i], token...)
				return
			}
			if err != nil {
				return 0, nil, err
			}
			return 0, nil, nil
		case c == '\n' || c == '\r' || c == '\t' || c == ' ':
			return i, data[:i], nil
		case c == '-' || c == '_':
		case '0' <= c && c <= '9':
		case 'A' <= c && c <= 'Z':
		case 'a' <= c && c <= 'z':
		case '\x00' <= c && c <= '\xed':
			// TODO(jwall) nonascii case
		default:
			return i, data[:i+1], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func consumeCdcOrIdent(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) >= 3 && string(data[:3]) == "-->" {
		return 3, data[:3], nil
	}
	return consumeIdent(data, atEOF)
}

// TODO(jwall): handle partial matches when !atEOF
func consumePrefix(data []byte, expect string, atEOF bool) (advance int, token []byte, err error) {
	el := len(expect)
	if len(data) >= el && string(data[:el]) == expect {
		return el, data[:el], nil
	}
	if atEOF {
		return 0, nil, fmt.Errorf("Invalid Token expecting %q got %q", expect, data[:2])
	}
	return 0, nil, nil
}

func consumeNumericOrUnit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, c := range data {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		case '%':
			return i + 1, data[:i+1], nil
		case ';':
			return i, data[:i], nil
		//case ' ', '\n', '\r', '\t', '\f':
		//	return i, data[:i], nil
		default:
			advance, token, err = consumeIdent(data[i:], atEOF)
			if advance > 0 {
				advance += i
				token = data[:len(token)+i]
				return
			}
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func consumeQuoted(data []byte, atEOF bool) (advance int, token []byte, err error) {
	next, tok, err := consumeQuotedBy(data[0])(data[1:], atEOF)
	if err != nil {
		return 0, nil, err
	}
	return next + 1, append(data[:1], tok...), nil
}

func consumeQuotedBy(q byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i, c := range data {
			if c == q {
				return i + 1, data[:i+1], nil
			}
			if c == '\\' {
				if len(data) > i+1 {
					next, tok, err := consumeEscapeOrUnicodeAnd(consumeQuotedBy(q))(data[i:], atEOF)
					if next > 0 {
						return next + i, append(data[:i], tok...), nil
					}
					if err != nil {
						return 0, nil, err
					}
				} else {
					return 0, nil, nil
				}
			}
		}
		return 0, nil, nil
	}
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	switch data[0] {
	case ':', ';', '{', '}', '(', ')', '[', ']':
		return 1, data[:1], nil
	case '"', '\'':
		return consumeQuoted(data, atEOF)
	case ' ', '\n', '\t', '\r', '\f':
		return consumeWhitespace(data)
	case '<':
		return consumeCdoOrDelim(data)
	case '-':
		return consumeCdcOrIdent(data, atEOF)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return consumeNumericOrUnit(data, atEOF)
	case '|':
		return consumePrefix(data, "|=", atEOF)
	case '~':
		return consumePrefix(data, "~=", atEOF)
	case '#', '@':
		advance, token, err = consumeIdent(data[1:], atEOF)
		if advance > 0 {
			advance++
			token = data[:len(token)+1]
			return
		}
	case '\\':
		fallthrough
	default:
		return consumeIdent(data, atEOF)
		// TODO(jwall): return consumeURI

	}
	if !atEOF {
		return 0, nil, nil
	}
	return 0, nil, fmt.Errorf("Invalid token %q", data)
}

func (t *Tokenizer) Next() (*Token, error) {
	if t.p.Scan() {
		tok := string(t.p.Bytes())
		switch tok {
		case ":":
			return &Token{Position: t.p.Position(), Type: Colon}, nil
		case ";":
			return &Token{Position: t.p.Position(), Type: Semicolon}, nil
		case "{":
			return &Token{Position: t.p.Position(), Type: LBrace}, nil
		case "}":
			return &Token{Position: t.p.Position(), Type: RBrace}, nil
		case "(":
			return &Token{Position: t.p.Position(), Type: LParen}, nil
		case ")":
			return &Token{Position: t.p.Position(), Type: RParen}, nil
		case "[":
			return &Token{Position: t.p.Position(), Type: LBracket}, nil
		case "]":
			return &Token{Position: t.p.Position(), Type: RBracket}, nil
		case "~=":
			return &Token{Position: t.p.Position(), Type: Includes}, nil
		case "|=":
			return &Token{Position: t.p.Position(), Type: Dashmatch}, nil
		case "<":
			return &Token{Position: t.p.Position(), Type: Delim, String: tok}, nil
		case "<!--":
			return &Token{Position: t.p.Position(), Type: CDO, String: tok}, nil
		case "-->":
			return &Token{Position: t.p.Position(), Type: CDC, String: tok}, nil
		default:
			switch tok[0] {
			case '@':
				return &Token{Position: t.p.Position(), Type: AtKeyword, String: tok}, nil
			case '#':
				return &Token{Position: t.p.Position(), Type: Hash, String: tok}, nil
			case '\n', '\t', '\r', '\f', ' ':
				return &Token{Position: t.p.Position(), Type: S, String: tok}, nil
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				return handleNumberPrefixToken(t.p.Position(), tok)
			case 'u':
				// TODO(jwall): return handleIdentOrUrlPrefixToken(tok)
			case '"', '\'':
				return &Token{Position: t.p.Position(), Type: String, String: tok}, nil
			default:
				return &Token{Position: t.p.Position(), Type: Ident, String: tok}, nil
			}
		}
	}
	return nil, t.p.Err()
}

func handleNumberPrefixToken(p Position, tok string) (*Token, error) {
	if strings.HasSuffix(tok, "%") {
		return &Token{Position: p, Type: Percentage, String: tok}, nil
	}
	for _, r := range tok {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return &Token{Position: p, Type: Dimension, String: tok}, nil
		}
	}
	return &Token{Position: p, Type: Number, String: tok}, nil
}
