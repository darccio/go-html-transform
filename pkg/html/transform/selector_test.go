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
	if sel.Type != typ {
		t.Errorf(msg)
		t.Logf("Type: [%s]", sel.Type)
	}
}

func assertKey(t *testing.T, sel *Selector, val string, msg string) {
	if sel.Key != val {
		t.Errorf(msg)
		t.Logf("Key: [%s]", sel.Val)
	}
}

func assertVal(t *testing.T, sel *Selector, val string, msg string) {
	if sel.Val != val {
		t.Errorf(msg)
		t.Logf("Val: [%s]", sel.Val)
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
	assertKey(t, sel, "foo", "selector key not foo")
	assertVal(t, sel, "bar", "selector val not bar")
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
	assertKey(t, sel, "foo", "selector key not foo")
	assertVal(t, sel, "bar", "selector val not bar")
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
	assertKey(t, sel, "foo", "selector key not foo")
	assertVal(t, sel, "bar", "selector val not bar")

	selString = "a"
	sel = NewSelector(selString)
	assertType(t, sel, TAGNAME, "selector type not TAGNAME")
	assertTagType(t, sel, "a", "selector TagType not a")

	selString = "a[foo=bar]"
	sel = NewSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagType(t, sel, "a", "selector TagType not a")
	assertKey(t, sel, "foo", "selector key not foo")
	assertVal(t, sel, "bar", "selector val not bar")

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
}

// TODO(jwall): tests for NewSelectorQuery
