package tokenizer

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

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
	{" \r\n\t\f", []Token{Token{Type: S, String: " \r\n\t\f"}}},
	{`<`, []Token{Token{Type: Delim, String: `<`}}},
	{`<!--`, []Token{Token{Type: CDO, String: `<!--`}}},
	{`-->`, []Token{Token{Type: CDC, String: `-->`}}},
	{`-css3-`, []Token{Token{Type: Ident, String: `-css3-`}}},
	{`_foo_`, []Token{Token{Type: Ident, String: `_foo_`}}},
	{`|=`, []Token{Token{Type: Dashmatch}}},
	{`~=`, []Token{Token{Type: Includes}}},
	{`@media`, []Token{Token{Type: AtKeyword, String: `@media`}}},
	{`#id`, []Token{Token{Type: Hash, String: `#id`}}},
	{`id`, []Token{Token{Type: Ident, String: `id`}}},
	{`123`, []Token{Token{Type: Number, String: `123`}}},
	{`_id`, []Token{Token{Type: Ident, String: `_id`}}},
	{`123%`, []Token{Token{Type: Percentage, String: `123%`}}},
	{`123em`, []Token{Token{Type: Dimension, String: `123em`}}},
	// nonascii test case
	{"\xEDfoo\x00", []Token{Token{Type: Ident, String: "\xEDfoo\x00"}}},
	{`\\foo`, []Token{Token{Type: Ident, String: `\\foo`}}},
	// Unicode
	{`\91f6td`, []Token{Token{Type: Ident, String: `\91f6td`}}},
	{`td\91f6dt`, []Token{Token{Type: Ident, String: `td\91f6dt`}}},
	// Strings
	{`"foo"`, []Token{Token{Type: String, String: `"foo"`}}},
	{`"\""`, []Token{Token{Type: String, String: `"\""`}}},
	{`\"fo\\r\n"`, []Token{Token{Type: String, String: `\"fo\\r\n"`}}},
	{`\"fo\\r\"`, []Token{Token{Type: String, String: `\"fo\\r\"`}}},
	{`\"fo\\n\"`, []Token{Token{Type: String, String: `\"fo\\n\"`}}},
	{`"fo\91f6o"`, []Token{Token{Type: String, String: `"fo\91f6o"`}}},
	{`"t\91f6t\""`, []Token{Token{Type: String, String: `"t\91f6t\""`}}},
	// TODO(jwall): Comment
	// TODO(jwall): Function
}

func testStream(t *testing.T, input string, out []Token) {
	//t.Logf("Parsing (%q)", input)
	tok := New(strings.NewReader(input))
	idx := 0
	var tk *Token
	var err error
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
		idx++
	}
	if len(out) > idx+1 {
		t.Errorf("Expected %d tokens but got %d", len(out), idx+1)
	}
}

func TestTokenizer(t *testing.T) {
	for _, c := range simpleTokenTests {
		testStream(t, c.in, c.out)
	}
}

func TestConsumeEscapedTwice(t *testing.T) {
	input := []byte(`\"\"`)
	n, tok, _ := consumeEscapedOrUnicode(input, false)
	if n != 2 {
		t.Errorf("Expected advance by 1 but got %d", n)
	}
	if !reflect.DeepEqual(tok, input[:n]) {
		t.Errorf("Expected token: %v but got %v", input[:n], tok)
	}
	n, tok, _ = consumeEscapedOrUnicode(input[:n], false)
	if n != 2 {
		t.Errorf("Expected advance by 1 but got %d", n)
	}
	if !reflect.DeepEqual(tok, input[n:]) {
		t.Errorf("Expected token: %v but got %v", input[n:], tok)
	}
}

func TestConsumeEscapedAnd(t *testing.T) {
	input := []byte(`\""`)
	n, tok, _ := consumeEscapeOrUnicodeAnd(consumeQuotedBy('"'))(input, false)
	if n != 3 {
		t.Errorf("Expected advance by 1 but got %d", n)
	}
	if !reflect.DeepEqual(tok, input) {
		t.Errorf("Expected token: %v but got %v", input, tok)
	}
}

func TestConsumeUnicode(t *testing.T) {
	input := []byte(`\91f6too`)
	n, tok, _ := consumeUnicode(input, false)
	expectedN := 5
	if n != expectedN {
		t.Errorf("Expected %d but advanced by %d", expectedN, n)
	}
	if !reflect.DeepEqual(tok, input[:expectedN]) {
		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
	}
}

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

func TestQuotedUnicode(t *testing.T) {
	input := []byte(`"\91f6"`)
	n, tok, _ := consumeQuoted(input, false)
	expectedN := 7
	if n != expectedN {
		t.Errorf("Expected %d but advanced by %d", expectedN, n)
	}
	if !reflect.DeepEqual(tok, input[:expectedN]) {
		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
	}
}

func TestQuotedEscapedQuote(t *testing.T) {
	input := []byte(`"\""`)
	n, tok, _ := consumeQuoted(input, false)
	expectedN := 4
	if n != expectedN {
		t.Errorf("Expected %d but advanced by %d", expectedN, n)
	}
	if !reflect.DeepEqual(tok, input[:expectedN]) {
		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
	}
}

func TestQuotedUnicodeAndEscapedQuote(t *testing.T) {
	input := []byte(`"\91f6\""`)
	n, tok, _ := consumeQuoted(input, false)
	expectedN := 9
	if n != expectedN {
		t.Errorf("Expected %d but advanced by %d", expectedN, n)
	}
	if !reflect.DeepEqual(tok, input[:expectedN]) {
		t.Errorf("Expected token `%s` got `%s`", input[:expectedN], tok)
	}
}
