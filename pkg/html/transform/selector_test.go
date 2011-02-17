/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
)

func assertTagTypeAny(t *testing.T, sel *Selector) {
	if sel.TagType != "*" {
		t.Errorf("selector tagType not ANY")
	}
}

func assertType(t *testing.T, sel *Selector, typ byte, msg string) {
	if sel.Type != typ {
		t.Errorf(msg)
	}
}

func assertVal(t *testing.T, sel *Selector, val string, msg string) {
	if sel.Val != val {
		t.Errorf(msg)
	}
}

func TestNewAnyTagClassSelector(t *testing.T) {
	selString := ".foo"
	sel := newAnyTagClassOrIdSelector(selString)
	assertType(t, sel, CLASS,"selector type not CLASS")
	assertTagTypeAny(t, sel)
	assertVal(t, sel, "foo", "selector tagType not foo")
}

func TestNewAnyTagSelector(t *testing.T) {
	selString := "*"
	sel := newAnyTagSelector(selString)
	assertType(t, sel, ANY,"selector type not ANY")
	assertTagTypeAny(t, sel)
}
