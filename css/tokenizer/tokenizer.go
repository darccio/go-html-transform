// package tokenizer tokenizes a css stream.
// Follows the spec defined at http://www.w3.org/TR/css-syntax/#tokenization

//go:generate stringer -type tokenType
package tokenizer

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"strings"
)

var EUnexpextedEOF = fmt.Errorf("Unexpected EOF.")

type tokenType int

type Position struct {
	Line   int
	Column int
}

// TODO(jwall): We need to sanitize the stream first? \r, \f, or \r\f
// turn into \n
type PositionTrackingScanner struct {
	// Embedded Scanner so our LineTrackingReader can be used just like
	// a Scanner.
	*bufio.Scanner
	lastL, lastCol int // Position Tracking fields.
	l, col         int
}

type sanitizingReader struct {
	*bufio.Reader
}

var unknownRune = []byte("\uFFFD")

func preprocess(buf []byte, r *bufio.Reader) (int, error) {
	i := 0
	for c, n, err := r.ReadRune(); err == nil; c, n, err = r.ReadRune() {
		if i+n > len(buf) {
			// We don't have room so unread the rune and return.
			r.UnreadRune()
			return i, err
		}
		switch c {
		case '\x00':
			if len(buf)-1 < i+len(unknownRune) {
				copy(buf[i:len(unknownRune)], unknownRune)
			} else {
				// We don't have room so unread the rune and
				// return.
				r.UnreadRune()
				return i + n, err
			}
		case '\r':
			buf[i] = '\n'
			nxt, err := r.Peek(1)
			if err == nil && len(nxt) == 1 && nxt[0] == '\n' {
				r.ReadByte()
			}
		case '\f':
			buf[i] = '\n'
		default:
			copy(buf[i:i+n], []byte(string(c)))
		}
		i += n
	}
	return i, nil
}

func (r *sanitizingReader) Read(buf []byte) (int, error) {
	return preprocess(buf, r.Reader)
}

func NewTrackingReader(r io.Reader, splitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)) *PositionTrackingScanner {
	s := bufio.NewScanner(&sanitizingReader{bufio.NewReader(r)})
	//s := bufio.NewScanner(r)
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
	Ident          tokenType = iota // 0
	AtKeyword                       // 1
	String                          // 2
	BadString                       // 3
	BadUri                          // 4
	BadComment                      // 5
	Hash                            // 6
	Number                          // 7
	Percentage                      // 8
	Dimension                       // 9
	Uri                             // 10
	UnicodeRange                    // 11
	CDO                             // 12
	CDC                             // 13
	Colon                           // 14
	Semicolon                       // 15
	Comma                           // 16
	LBrace                          // 17
	RBrace                          // 18
	LParen                          // 19
	RParen                          // 20
	LBracket                        // 21
	RBracket                        // 22
	Includes                        // 23
	Prefixmatch                     // 24
	Suffixmatch                     // 25
	Dashmatch                       // 26
	Comment                         // 27
	Function                        // 28
	Delim                           // 29
	SubstringMatch                  // 30
	Column                          // 31
	WS                              // 32
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

func consumeHexDigits(data []byte, atEOF bool) (int, []byte, error) {
	n := 7
	if len(data) < n {
		n = len(data)
	}
	tok := data
LOOP:
	for i, c := range data[:n] {
		switch {
		case '0' <= c && c <= '9':
			fallthrough
		case 'A' <= c && c <= 'F':
			fallthrough
		case 'a' <= c && c <= 'f':
			tok = data[:i+1]
		case c == ' ', c == '\n', c == '\t':
			if i == 0 {
				return 0, nil, fmt.Errorf("Invalid Hex digit char %q", c)
			}
			tok = data[:i+1]
			break LOOP
		default:
			if i == 0 {
				return 0, nil, fmt.Errorf("Invalid Hex digit char %q", c)
			}
			break LOOP
		}
	}
	return len(tok), tok, nil
}

func consumeEscaped(data []byte, atEOF bool) (advance int, token []byte, err error, isHex bool) {
	if data[0] != '\\' {
		return 0, nil, fmt.Errorf("Expected Escaped character got: %q", data), false
	}
	c := data[1]
	switch {
	case '0' <= c && c <= '9':
		fallthrough
	case 'A' <= c && c <= 'F':
		fallthrough
	case 'a' <= c && c <= 'f':
		n, _, err := consumeHexDigits(data[1:], atEOF)
		return n + 1, data[:n+1], err, true
	case c == '\n':
		return 0, nil, fmt.Errorf("Invalid escape, non escapable char %c", c), false
	default:
		// consume two codepoints
		tok := []byte(string(data)[:2])
		return len(tok), tok, nil, false
	}
}

func consumeUnicodeRange(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !(data[0] == 'u' || data[0] == 'U') {
		return 0, nil, nil
	}
	log.Printf("datalen %d", len(data))
	n, tok, err := consumeHexDigits(data[1:], atEOF)
	log.Printf("xxx: Consumed %d hex digits %q", n, tok)
	if tok != nil && n <= 6 {
		if len(data) > n+1 && data[n+1] == '-' {
			log.Printf("data[n+1 %q", data[n+1])
			n2, tok2, err := consumeHexDigits(data[n+2:], atEOF)
			log.Printf("Consumed %d hex digits %q", n2, tok2)
			if tok2 != nil && n2 <= 6 {
				advance = n + n2 + 2
				return advance, data[:advance], nil
			} else {
				return 0, nil, err
			}
		} else {
			return n + 1, data[:n+1], err
		}
	}
	return 0, nil, err
}

func consumeIdent(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, c := range data {
		// TODO(jwall): consumeIdentStart
		switch {
		case c == '\\':
			advance, token, err, _ = consumeEscaped(data, atEOF)
			if advance > 0 {
				advance += i
				token = append(data[:i], token...)
				continue
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
func consumeOnePrefix(data []byte, expected []string, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, fmt.Errorf("Invalid Token expecting one of %v got %q", expected, data[:2])
	}
	for _, expect := range expected {
		el := len(expect)
		if len(data) >= el && string(data[:el]) == expect {
			return el, data[:el], nil
		}
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
					//log.Printf("consuming escaped %q", data[i:])
					next, _, err, isHex := consumeEscaped(data[i:], atEOF)
					//log.Printf("next: %d, tok: %q, err: %v", next, tok, err)
					if err != nil {
						if len(data) > next+1 && data[i+1] == '\n' {
							next, tok, err := consumeQuotedBy(q)(data[i+2:], atEOF)
							if err != nil {
								return 0, nil, err
							}
							escaped := append(append(data[:i], data[i+1]), tok...)
							next = next + i + 2
							return next, escaped, nil
						}
						return 0, nil, err
					}
					if next > 0 {
						next, tok, err := consumeQuotedBy(q)(data[i+2:], atEOF)
						if err != nil {
							return 0, nil, err
						}
						escaped := data[:i]
						if isHex {
							escaped = append(escaped, c)
						}
						escaped = append(append(escaped, data[i+1]), tok...)
						next = next + i + 2
						return next, escaped, nil
						//return next + i, append(data[:i], tok...), nil
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
	// TODO(jwall): Do preprocessing before we get to this splitfunc.
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}
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
		return consumeOnePrefix(data, []string{"|=", "||"}, atEOF)
	case '~':
		return consumeOnePrefix(data, []string{"~="}, atEOF)
	case '#', '@':
		advance, token, err = consumeIdent(data[1:], atEOF)
		if advance > 0 {
			advance++
			token = data[:len(token)+1]
			return
		}
	case 'u', 'U':
		log.Printf("Consuming Unicode %q", data)
		if n, tok, err := consumeUnicodeRange(data, atEOF); tok != nil {
			log.Printf("Consumed Unicode Range %q", tok)
			return n, tok, err
		} else {
			// Ident?
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
		case "^=":
			return &Token{Position: t.p.Position(), Type: Prefixmatch}, nil
		case "$=":
			return &Token{Position: t.p.Position(), Type: Suffixmatch}, nil
		case "*=":
			return &Token{Position: t.p.Position(), Type: SubstringMatch}, nil
		case "|=":
			return &Token{Position: t.p.Position(), Type: Dashmatch}, nil
		case "||":
			return &Token{Position: t.p.Position(), Type: Column}, nil
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
				return &Token{Position: t.p.Position(), Type: WS, String: tok}, nil
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				return handleNumberPrefixToken(t.p.Position(), tok)
			case 'u':
				log.Printf("Encountered possible unicode token %q", tok)
				// TODO(jwall): return handleIdentUnicodeOrUrlPrefixToken(tok)
				return handleUnicodeToken(t.p.Position(), tok)
			case '"', '\'':
				return &Token{Position: t.p.Position(), Type: String, String: tok}, nil
			default:
				return &Token{Position: t.p.Position(), Type: Ident, String: tok}, nil
			}
		}
	}
	return nil, t.p.Err()
}

func handleUnicodeToken(p Position, tok string) (*Token, error) {
	dashCount := 0
	nonHexChar := 0
	for _, r := range tok {
		switch {
		case r == '-':
			dashCount += 1
		case r == 'u' || r == 'U':
		case '0' <= r && r <= '9':
		case 'A' <= r && r <= 'F':
		case 'a' <= r && r <= 'f':
		default:
			nonHexChar += 1
		}
	}
	log.Printf("dashCount = %d", dashCount)
	log.Printf("nonHexChar = %d", nonHexChar)
	if dashCount <= 1 && nonHexChar == 0 {
		log.Printf("Emiting unicode token: %q", tok)
		return &Token{Position: p, Type: UnicodeRange, String: tok}, nil
	}
	return nil, nil
}

func handleNumberPrefixToken(p Position, tok string) (*Token, error) {
	if strings.HasSuffix(tok, "%") {
		return &Token{Position: p, Type: Percentage, String: tok}, nil
	}
	eCount := 0
	for _, r := range tok {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		case 'e', 'E':
			// This might be a number using scientific notation.
			eCount += 1
		default:
			switch tok[len(tok)-1] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// It's an Ident token if there is non numerical in the middle.
				return &Token{Position: p, Type: Ident, String: tok}, nil
			}
			return &Token{Position: p, Type: Dimension, String: tok}, nil
		}
	}
	if eCount <= 1 {
		return &Token{Position: p, Type: Number, String: tok}, nil
	}
	return &Token{Position: p, Type: Ident, String: tok}, nil
}
