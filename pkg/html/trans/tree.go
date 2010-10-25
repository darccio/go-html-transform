/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/ 
package trans

import (
	b "bytes"
	h "html"
	"os"
	"log"
	//s "strings"
)

type NodeType int

const (
	TEXT NodeType = iota // 0 value so the default
	TAG
)

type Node struct {
	nodeType NodeType
	nodeValue string
	nodeAttributes map[string] string
	children []Node
}

func lazyTokens(t *h.Tokenizer) <-chan h.Token {
	tokens := make(chan h.Token, 1)
	go func() {
		for {
			tt := t.Next()
			if tt == h.Error {
				switch t.Error() {
				case os.EOF:
					break
				default:
					log.Panicf("Error tokenizing string: %s",
						t.Error())
				}
			}
			tokens <- t.Token()
		}
	}()
	return tokens
}

func transformAttributes(t *h.Tokenizer) {
	attr := make(map[string] string)
	for {
		key, val, rem := t.TagAttr()
		sKey, sVal := b.NewBuffer(key).String(), b.NewBuffer(val).String()
		if sKey != "" {
			attr[sKey] = sVal
		}
		if !rem {
			break
		}
	}
}

// TODO(jwall): reducer that builds a tree out of the tokens
