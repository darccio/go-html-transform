/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
)

func assertTagType(t *testing.T, sel *Selector, typ string, msg string) {
	if sel.TagType != typ {
		t.Errorf(msg)
		t.Logf("TagType: [%s]", sel.TagType)
	}
}

func assertTagTypeAny(t *testing.T, sel *Selector) {
	assertTagType(t, sel, "*", "selector tagType not ANY")
}

func assertType(t *testing.T, sel *Selector, typ byte, msg string) {
	var mask byte = 0
	for _, part := range sel.Parts {
		mask = mask & part.Type
	}
	if (mask | typ) != typ {
		t.Errorf(msg)
		t.Logf("MaskResult: [%s], Type: [%s]",
			mask & typ, typ)
	}
}

func assertAttr(t *testing.T, sel *Selector, key string, val string, msg string) {
	if sel.Attr[key] != val {
		t.Errorf(msg)
		t.Logf("Key: [%s]", sel.Attr[key])
	}
}

func assertVal(t *testing.T, sel *Selector, val string, msg string) {
	for _, part := range sel.Parts {
		if part.Val != val {
			t.Errorf(msg)
			t.Logf("Val: [%s]", part.Val)
		}
	}
}

func TestNewAnyTagClassSelector(t *testing.T) {
	selString := ".foo"
	sel := newAnyTagClassOrIdSelector(selString)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagTypeAny(t, sel)
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestNewAnyTagSelector(t *testing.T) {
	selString := "*"
	sel := newAnyTagSelector(selString)
	assertType(t, sel, ANY,"selector type not ANY")
	assertTagTypeAny(t, sel)
}

func TestNewAnyTagAttrSelector(t *testing.T) {
	selString := "[foo=bar]"
	sel := newAnyTagAttrSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagTypeAny(t, sel)
	assertAttr(t, sel, "foo", "bar", "selector key not foo")
}

func TestTagNameSelector(t *testing.T) {
	selString := "a"
	sel := newTagNameSelector(selString)
	assertType(t, sel, TAGNAME, "selector type not TAGNAME")
	assertTagType(t, sel, "a", "selector TagType not a")
}

func TestTagNameWithAttr(t *testing.T) {
	selString := "a[foo=bar]"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertAttr(t, sel, "foo", "bar", "selector key not foo")
}

func TestTagNameWithClass(t *testing.T) {
	selString := "a.foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestTagNameWithId(t *testing.T) {
	selString := "a#foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, ID, "selector type not ID")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestTagNameWithPseudo(t *testing.T) {
	selString := "a:foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, PSEUDO, "selector type not PSEUDO")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestNewSelector(t *testing.T) {
	selString := ".foo"
	sel := NewSelector(selString)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagTypeAny(t, sel)
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "*"
	sel = NewSelector(selString)
	assertType(t, sel, ANY,"selector type not ANY")
	assertTagTypeAny(t, sel)

	selString = "[foo=bar]"
	sel = NewSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagTypeAny(t, sel)
	assertAttr(t, sel, "foo", "bar", "selector key not foo")

	selString = "a"
	sel = NewSelector(selString)
	assertType(t, sel, TAGNAME, "selector type not TAGNAME")
	assertTagType(t, sel, "a", "selector TagType not a")

	selString = "a[foo=bar]"
	sel = NewSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertAttr(t, sel, "foo", "bar", "selector key not foo")

	selString = "a.foo"
	sel = NewSelector(selString)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "a#foo"
	sel = NewSelector(selString)
	assertType(t, sel, ID, "selector type not ID")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "a:foo"
	sel = NewSelector(selString)
	assertType(t, sel, PSEUDO, "selector type not PSEUDO")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")

	// TODO(jwall): support combinators > + \S
}

func TestNewSelectorQuery(t *testing.T) {
	 NewSelectorQuery("a.foo", ".bar", "[id=foobar]")
	q := NewSelectorQuery("a.foo", ".bar", "[id=foobar]")
	sel := q.At(0).(*Selector)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertVal(t, sel, "foo", "selector val not foo")

	sel = q.At(1).(*Selector)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagTypeAny(t, sel)
	assertVal(t, sel, "bar", "selector val not foo")

	sel = q.At(2).(*Selector)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagTypeAny(t, sel)
	assertAttr(t, sel, "id", "foobar", "selector key not foo")
}
