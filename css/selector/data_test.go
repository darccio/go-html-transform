package selector

import (
	"testing"
)

var chains = []string{
	// Universal
	"*",
	// sequence
	"ul.foo.bar:first-child::first-line",
	// combinators
	"ul>li",
	"ul li",
	"ul+li",
	"ul~li",
	// multiple link chains
	"ul.foo.bar:first-child::first-line",
	"ul.foo.bar:first-child::first-line>a.link",
	"ul.foo.bar:first-child::first-line>a.link+br.quux",
	"ul.foo.bar:first-child::first-line>a.link+br.quux~hr.sep div",
}

func TestSelectorString(t *testing.T) {
	for _, chn := range chains {
		sel, err := Selector(chn)
		if err != nil {
			t.Errorf("Error parsing %q %q", chn, err)
		}
		if sel.String() != chn {
			t.Errorf("%q != %q", sel.String(), chn)
		}
	}
	// test EOS for { characters
	sel, err := Selector("ul li {")
	if err != EOS {
		t.Errorf("Selector didn't return End of Selector %q", err)
	}
	if sel.String() != "ul li" {
		t.Errorf("Selector %q != %q", sel.String(), "ul li")
	}
}
