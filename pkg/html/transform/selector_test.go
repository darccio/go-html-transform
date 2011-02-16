/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
)

func TestNewAnyTagClassSelector(t *testing.T) {
	selString := ".foo"
	sel := newAnyTagClassSelector(selString)
        if sel.Type != '.' {
		t.Errorf("selector type not class")
	}
        if sel.TagType != "*" {
		t.Errorf("selector tagType not *")
	}
        if sel.Val != "foo" {
		t.Errorf("selector tagType not foo")
	}
}
