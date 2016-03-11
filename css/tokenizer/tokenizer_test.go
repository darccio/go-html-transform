package tokenizer

import (
	"bufio"
	"fmt"

	"io"
	"reflect"
	"strings"
	"testing"
)

func compareErrors(e1, e2 error) bool {
	if e1 == nil {
		return e2 == nil
	} else if e2 != nil {
		return e1.Error() == e2.Error()
	}
	return false
}

var simpleTokenTests = []struct {
	in  string
	out []Token
}{
	// Basic simple token cases
	{`:`, []Token{Token{Type: Colon}}},
	{`;`, []Token{Token{Type: Semicolon}}},
	{`{`, []Token{Token{Type: LBrace}}},
	{`}`, []Token{Token{Type: RBrace}}},
	{`(`, []Token{Token{Type: LParen}}},
	{`)`, []Token{Token{Type: RParen}}},
	{`[`, []Token{Token{Type: LBracket}}},
	{`]`, []Token{Token{Type: RBracket}}},
	{`||`, []Token{Token{Type: Column}}},
	{"\r\n", []Token{Token{Type: WS, String: "\n"}}},
	{"\r", []Token{Token{Type: WS, String: "\n"}}},
	{"\f", []Token{Token{Type: WS, String: "\n"}}},
	{" \r\n\t\f", []Token{Token{Type: WS, String: " \n\t\n"}}},
	{`<`, []Token{Token{Type: Delim, String: `<`}}},
	{`<!--`, []Token{Token{Type: CDO, String: `<!--`}}},
	{`-->`, []Token{Token{Type: CDC, String: `-->`}}},
	{`-css3-`, []Token{Token{Type: Ident, String: `-css3-`}}},
	{`_foo_`, []Token{Token{Type: Ident, String: `_foo_`}}},
	{`|=`, []Token{Token{Type: Dashmatch}}},
	{`^=`, []Token{Token{Type: Prefixmatch}}},
	{`$=`, []Token{Token{Type: Suffixmatch}}},
	{`*=`, []Token{Token{Type: SubstringMatch}}},
	{`~=`, []Token{Token{Type: Includes}}},
	{`@media`, []Token{Token{Type: AtKeyword, String: `@media`}}},
	{`#id`, []Token{Token{Type: Hash, String: `#id`}}},
	{`id`, []Token{Token{Type: Ident, String: `id`}}},
	{`123`, []Token{Token{Type: Number, String: `123`}}},
	{`123e10`, []Token{Token{Type: Number, String: `123e10`}}},
	{`_id`, []Token{Token{Type: Ident, String: `_id`}}},
	{`123%`, []Token{Token{Type: Percentage, String: `123%`}}},
	{`123em`, []Token{Token{Type: Dimension, String: `123em`}}},
	{`\1abcdf`, []Token{Token{Type: Ident, String: `\1abcdf`}}},
	{`\1`, []Token{Token{Type: Ident, String: `\1`}}},
	{`\z`, []Token{Token{Type: Ident, String: `\z`}}},
	{`\z `, []Token{Token{Type: Ident, String: `\z`},
		Token{Type: WS, String: ` `}}},
	{`\f`, []Token{Token{Type: Ident, String: `\f`}}},
	{`\"fo`, []Token{Token{Type: Ident, String: `\"fo`}}},
	// Strings
	{`"foo"`, []Token{Token{Type: String, String: `"foo"`}}},
	{"\"fo\\\n\"", []Token{Token{Type: String, String: "\"fo\n\""}}},
	{"'fo\\\n'", []Token{Token{Type: String, String: "'fo\n'"}}},
	{"\"\\\"\"", []Token{Token{Type: String, String: "\"\"\""}}},
	// HexDigit cases
	{`"fo\91f6o"`, []Token{Token{Type: String, String: `"fo\91f6o"`}}},
	{`"t\91f6t"`, []Token{Token{Type: String, String: `"t\91f6t"`}}},
	// TODO Unicode Range
	//{`u91f6td`, []Token{Token{Type: String, String: `u91f6td`}}},
	//{`u91f6td-A2FE63`, []Token{Token{Type: String, String: `u91f6td-A2FE63`}}},
	{`u91f651`, []Token{Token{Type: UnicodeRange, String: `u91f651`}}},
	{`u91f651-A2FE63`, []Token{Token{Type: UnicodeRange, String: `u91f651-A2FE63`}}},
	// nonascii test case
	// TODO {"\xEDfoo\x00", []Token{Token{Type: Ident, String: "\xEDfoo\uFFFD"}}},
	//{`\\foo`, []Token{Token{Type: Ident, String: `\\foo`}}},
	//{`\91f6td`, []Token{Token{Type: Ident, String: `\91f6td`}}},
	//{`td\91f6dt`, []Token{Token{Type: Ident, String: `td\91f6dt`}}},
	// TODO(jwall): Comment
	// TODO(jwall): Function
	// TODO(jwall): URL
}

func testStream(t *testing.T, input string, out []Token) {
	//t.Logf("Parsing (%q)", input)
	tok := New(strings.NewReader(input))
	idx := 0
	var tk *Token
	var err error
	parsed := []*Token{}
	for tk, err = tok.Next(); tk != nil && err != io.EOF; tk, err = tok.Next() {
		if len(out) == idx {
			t.Errorf("Expected %d tokens got at least %d for %q with err: %q this token: (%s)", len(out), idx+1, input, err, tk.String)
			break
		}
		if err != nil {
			t.Error(err.Error())
			break
		}
		if out[idx].String != tk.String {
			t.Errorf("Expected string (%q) got (%q)", out[idx].String, tk.String)
		}
		if out[idx].Type != tk.Type {
			t.Logf("Token %q", tk.String)
			t.Errorf("Expected type %v got %v", out[idx].Type, tk.Type)
		}
		parsed = append(parsed, tk)
		idx++
	}
	if len(out) != idx {
		t.Errorf("Expected %d tokens for %q but got %v", len(out), input, parsed)
	}
}

func TestTokenizer(t *testing.T) {
	for _, c := range simpleTokenTests {
		testStream(t, c.in, c.out)
	}
}

//func TestConsumeUnicode(t *testing.T) {
//	input := []byte(`\91f6too`)
//	n, tok, _ := consumeUnicodeRange(input, false)
//	expectedN := 5
//	if n != expectedN {
//		t.Errorf("Expected %d but advanced by %d", expectedN, n)
//	}
//	if !reflect.DeepEqual(tok, input[:expectedN]) {
//		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
//	}
//}

func TestQuoted(t *testing.T) {
	input := []byte(`"foo"`)
	n, tok, _ := consumeQuoted(input, false)
	expectedN := 5
	if n != expectedN {
		t.Errorf("Expected %d but advanced by %d", expectedN, n)
	}
	if !reflect.DeepEqual(tok, input[:expectedN]) {
		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
	}
}

var hexCases = []struct {
	input    string
	expected string
	atEOF    bool
	err      error
}{
	{`1abcdf `, `1abcdf `, true, nil},
	{"1abcdf\t", "1abcdf\t", true, nil},
	{"1abcdf\n", "1abcdf\n", true, nil},
	{`2abcdf`, `2abcdf`, true, nil},
	{`4azcdf `, `4a`, true, nil},
	{`zcdf `, ``, true, fmt.Errorf("Invalid Hex digit char %q", 'z')},
}

func TestConsumeEscapedHex(t *testing.T) {
	for _, c := range hexCases {
		n, tok, err := consumeHexDigits([]byte(c.input), c.atEOF)
		if string(tok) != c.expected {
			t.Errorf("Expected %q got %q consumed: %d", c.expected, tok, n)
		}
		if !compareErrors(err, c.err) {
			t.Errorf("Expected err %q got err %q", c.err, err)
		}
	}
}

var escapedCases = []struct {
	input    string
	expected string
	atEOF    bool
	err      error
	isHex    bool
}{
	{`\1abcdf `, `\1abcdf `, true, nil, true},
	{`\2abcdf`, `\2abcdf`, true, nil, true},
	{`\4azcdf `, `\4a`, true, nil, true},
	{`\zcdf `, `\z`, true, nil, false},
	{"\\\n ", ``, true, fmt.Errorf("Invalid escape, non escapable char %c", '\n'), false},
}

func TestEscaped(t *testing.T) {
	for _, c := range escapedCases {
		n, tok, err, isHex := consumeEscaped([]byte(c.input), c.atEOF)
		if string(tok) != c.expected {
			t.Errorf("Expected %q got %q consumed: %d", c.expected, tok, n)
		}
		if !compareErrors(err, c.err) {
			t.Errorf("Expected err %q got err %q", c.err, err)
		}
		if c.isHex != isHex {
			t.Errorf("Expected isHex %v got %v", c.isHex, isHex)
		}
	}
}

func TestPreprocess(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"\x00", "\uFFFD"},
		{"\r", "\n"},
		{"\f", "\n"},
		{"\n", "\n"},
		{"\r\n", "\n"},
		{"\r\nfoo", "\nfoo"},
	}
	for _, c := range cases {
		got := make([]byte, len(c.expected))
		r := &sanitizingReader{bufio.NewReader(strings.NewReader(c.input))}
		r.Read(got)
		if string(got) != c.expected {
			t.Errorf("Expected %q got %q", c.expected, got)
		}
	}
}
