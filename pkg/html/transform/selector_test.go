/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
)

func assertTagName(t *testing.T, sel *Selector, typ string, msg string) {
	if sel.TagName != typ {
		t.Errorf(msg)
		t.Logf("TagName: [%s]", sel.TagName)
	}
}

func assertTagNameAny(t *testing.T, sel *Selector) {
	assertTagName(t, sel, "*", "selector tagType not ANY")
}

func assertClass(t *testing.T, sel *Selector, class string) {
	ok := false
	for _, part := range sel.Parts {
		if part.Type == CLASS && part.Val == class {
			ok = true
		}
	}
	if !ok {
		t.Errorf("Selector has no class constraint %s", class)
	}
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
	assertTagNameAny(t, sel)
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestNewAnyTagSelector(t *testing.T) {
	selString := "*"
	sel := newAnyTagSelector(selString)
	assertType(t, sel, ANY,"selector type not ANY")
	assertTagNameAny(t, sel)
}

func TestNewAnyTagAttrSelector(t *testing.T) {
	selString := "[foo=bar]"
	sel := newAnyTagAttrSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagNameAny(t, sel)
	assertAttr(t, sel, "foo", "bar", "selector key not foo")
}

func TestNewAnyTagMultipleAttrSelector(t *testing.T) {
	selString := "[foo=bar][baz=boo]"
	sel := newAnyTagAttrSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagNameAny(t, sel)
	assertAttr(t, sel, "foo", "bar", "selector attr foo not bar")
	assertAttr(t, sel, "baz", "boo", "selector baz not boo")
}

func TestTagNameSelector(t *testing.T) {
	selString := "a"
	sel := newTagNameSelector(selString)
	assertType(t, sel, TAGNAME, "selector type not TAGNAME")
	assertTagName(t, sel, "a", "selector TagName not a")
}

func TestTagNameWithAttr(t *testing.T) {
	selString := "a[foo=bar]"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertAttr(t, sel, "foo", "bar", "selector key not foo")
}

func TestTagNameWithClass(t *testing.T) {
	selString := "a.foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestTagNameWithId(t *testing.T) {
	selString := "a#foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, ID, "selector type not ID")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestTagNameWithPseudo(t *testing.T) {
	selString := "a:foo"
        sel := newTagNameWithConstraints(selString, 1)
	assertType(t, sel, PSEUDO, "selector type not PSEUDO")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestNewSelector(t *testing.T) {
	selString := ".foo"
	sel := NewSelector(selString)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagNameAny(t, sel)
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "*"
	sel = NewSelector(selString)
	assertType(t, sel, ANY,"selector type not ANY")
	assertTagNameAny(t, sel)

	selString = "[foo=bar]"
	sel = NewSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagNameAny(t, sel)
	assertAttr(t, sel, "foo", "bar", "selector key not foo")

	selString = "a"
	sel = NewSelector(selString)
	assertType(t, sel, TAGNAME, "selector type not TAGNAME")
	assertTagName(t, sel, "a", "selector TagName not a")

	selString = "a[foo=bar]"
	sel = NewSelector(selString)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertAttr(t, sel, "foo", "bar", "selector key not foo")

	selString = "a.foo"
	sel = NewSelector(selString)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "a#foo"
	sel = NewSelector(selString)
	assertType(t, sel, ID, "selector type not ID")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")

	selString = "a:foo"
	sel = NewSelector(selString)
	assertType(t, sel, PSEUDO, "selector type not PSEUDO")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")
}

func TestMergeSelectorsBaseCase(t *testing.T) {
	sel1 := NewSelector(".foo")
	sel2 := NewSelector("a")
	sel3 := NewSelector("[foo=bar]")

	MergeSelectors(sel1, sel2)
	MergeSelectors(sel1, sel3)

	assertType(t, sel1, TAGNAME, "selector type not TAGNAME")
	assertTagName(t, sel1, "a", "selector TagName not a")
	assertClass(t, sel1, "foo")
	assertAttr(t, sel1, "foo", "bar", "selector key not foo")
}

func TestMergeSelectorsMultipleParts(t *testing.T) {
	sel1 := NewSelector(".foo")
	sel2 := NewSelector(".bar")
	sel3 := NewSelector("[foo=bar]")

	MergeSelectors(sel1, sel2)
	MergeSelectors(sel1, sel3)

	assertClass(t, sel1, "foo")
	assertClass(t, sel1, "bar")
}

func TestMergeSelectorsEmptySelectors(t *testing.T) {
	defer func() {
		if err := recover(); err != nil{
			t.Error("Merging two Empty Selectors failed %s", err)
		}
	}()
	sel1 := new(Selector)
	sel2 := new(Selector)

	MergeSelectors(sel1, sel2)
}

func TestMergeSelectorsTwoTagNames(t *testing.T) {
	defer func() {
		if err := recover(); err == nil{
			t.Error("Merging two Selectors with tagnames did not fail")
		}
	}()
	sel1 := NewSelector("hr")
	sel2 := NewSelector("a")

	MergeSelectors(sel1, sel2)
}

func TestNewSelectorMultipleConstraints(t *testing.T) {
	selStr := "a.foo.bar"
	sel := NewSelector(selStr)
	assertClass(t, sel, "foo")
	assertClass(t, sel, "bar")
}

func TestNewSelectorQuery(t *testing.T) {
	 NewSelectorQuery("a.foo", ".bar", "[id=foobar]")
	q := NewSelectorQuery("a.foo", ".bar", "[id=foobar]")
	sel := q.At(0).(*Selector)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagName(t, sel, "a", "selector TagName not a")
	assertVal(t, sel, "foo", "selector val not foo")

	sel = q.At(1).(*Selector)
	assertType(t, sel, CLASS, "selector type not CLASS")
	assertTagNameAny(t, sel)
	assertVal(t, sel, "bar", "selector val not foo")

	sel = q.At(2).(*Selector)
	assertType(t, sel, ATTR, "selector type not ATTR")
	assertTagNameAny(t, sel)
	assertAttr(t, sel, "id", "foobar", "selector key not foo")
}

func TestPartition(t *testing.T) {
	testStr := "foo.bar,baz.blah"
	parted := partition(testStr, func(c int) bool {
		if c == '.' {
			return true
		}
		return false
	})
	if len(parted) != 3 {
		t.Errorf("partition count is not 3 but %d", len(parted))
		t.Logf("Parted: %s", parted)
	} else {
		if parted[0] != "foo" {
			t.Errorf("First partion is not foo")
		}
		if parted[1] != ".bar,baz" {
			t.Errorf("second partion is not bar,baz")
		}
		if parted[2] != ".blah" {
			t.Errorf("third partion is not blah")
		}
	}
}

func TestPartitionInitialChar(t *testing.T) {
	testStr := ".foo"
	parted := partition(testStr, func(c int) bool {
		if c == '.' {
			return true
		}
		return false
	})
	if len(parted) != 1 {
		t.Errorf("partition count is not 1 but %d", len(parted))
		t.Logf("Parted: %s", parted)
	} else {
		if parted[0] != ".foo" {
			t.Errorf("First partion is not foo")
		}
	}
}
