package transform

import (
	. "h5"
	"os"
	"io"
)

func NewDoc(str string) (*Node, os.Error) {
	p := NewParserFromString(str)
	err := p.Parse()
	return p.Top, err
}

func NewDocFromReader(rdr io.Reader) (*Node, os.Error) {
	p := NewParser(rdr)
	err := p.Parse()
	return p.Top, err
}

// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
