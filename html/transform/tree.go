// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.

package transform

import (
	"code.google.com/p/go-html-transform/h5"
	"io"
)

func NewDoc(str string) (h5.Tree, error) {
	p, err := h5.NewParserFromString(str)
	return p.Tree(), err
}

func NewDocFromReader(rdr io.Reader) (h5.Tree, error) {
	p, err := h5.NewParser(rdr)
	return p.Tree(), err
}
