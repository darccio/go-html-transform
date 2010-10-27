/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	l "container/list"
	s "strings"
)

type SelectorQuery struct {
	*l.List
}

type Selector struct {
	Type byte
	Tagtype string
	Key string
	Val string
}

const (
	TAGNAME byte = iota // zero value so the default
	CLASS   byte = '.'
	ID      byte = '#'
	PSEUDO  byte = ':'
	ANY     byte = '*'
	ATTR    byte = '['
)

// TODO(jwall): feels too big can I break it up?
func NewSelector(sel ...string) *SelectorQuery {
	q := SelectorQuery{List: l.New()}
	splitAttrs := func(str string) []string { 
		attrs := s.FieldsFunc(str[1:-1], func(c int) bool {
			if c == '=' {
				return true
			}
			return false
		})
		return attrs[0:1]
	}
	for _, str := range sel {
		str = s.TrimSpace(str) // trim whitespace
		var selector Selector
		switch str[0] {
		case CLASS, ID: // Any tagname with class or id
			selector = Selector{
			Type:str[0],
			Tagtype: "*",
			Val: str[1:],
			}
		case ANY: // Any tagname
			selector = Selector{
			Type: str[0],
			Tagtype: "*",
			}
		case ATTR: // any tagname with attribute
			attrs := splitAttrs(str)
			selector = Selector{
			Tagtype: "*",
			Type: str[0],
			Key: attrs[0],
			Val: attrs[1],
			}
		default: // TAGNAME
			if i := s.IndexAny(str, ".:#["); i != -1 {
				switch str[i] {
				case CLASS, ID, PSEUDO: // with class or id
					selector = Selector{
					Tagtype: str[0:i - 1],
					Val: str[i:],
					}
				case ATTR: // with attribute
					attrs := splitAttrs(str[i + 1:])
					selector = Selector{
					Tagtype: str[0:i - 1],
					Key: attrs[0],
					Val: attrs[1],
					}
				}
			} else { // just a tagname
				selector = Selector{
				Tagtype: str,
				}
			}
		}
		q.PushBack(selector)
	}
	return &q
}
