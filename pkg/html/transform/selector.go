/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	v "container/vector"
	s "strings"
)

type SelectorQuery struct {
	*v.Vector
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
	q := SelectorQuery{}
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
		q.Insert(0, selector)
	}
	return &q
}

func testNode(node *Node, sel Selector) bool {
	if sel.Tagtype == "*" {
		attrs := node.nodeAttributes
		// TODO(jwall): abstract this out
		switch sel.Type {
		case ID:
			if attrs["id"] == sel.Val {
				return true
			}
		case CLASS:
			if attrs["class"] == sel.Val {
				return true
			}
		case ATTR:
			if attrs[sel.Key] == sel.Val {
				return true
			}
		//case PSEUDO:
			//TODO(jwall): implement these
		}
	} else {
		if node.nodeValue == sel.Tagtype {
			attrs := node.nodeAttributes
			switch sel.Type {
			case ID:
				if attrs["id"] == sel.Val {
					return true
				}
			case CLASS:
				if attrs["class"] == sel.Val {
					return true
				}
			case ATTR:
				if attrs[sel.Key] == sel.Val {
					return true
				}
			//case PSEUDO:
				//TODO(jwall): implement these
			default:
				return true
			}
		}
	}
	return false;
}

func listToNodeVector(l *v.Vector) *v.Vector {
	nv := new(v.Vector)
	for true {
		nv.Push(l.Pop())
	}
	return nv
}

func (sel *SelectorQuery) Apply(doc *Document) *v.Vector {
	interesting := new(v.Vector)
	interesting.Push(doc.top.children[0])
	for i := 0; i <= sel.Len(); i++ {
		q := new(v.Vector)
		selector := sel.At(i).(Selector)
		for true {
			if interesting.Len() == 0 {
				break
			}
			node := interesting.Pop().(*Node)
			if testNode(node, selector) {
				q.Push(node)
			}
		}
		interesting = q
	}
	return listToNodeVector(interesting) // TODO(jwall): implement
}

/*
 Replace each node the selector matches with the passed in node.

 Applies the selector against the doc and replaces the returned
 Nodes with the passed in n Node.
 */
func (sel *SelectorQuery) Replace(doc *Document, n Node) {
	nv := sel.Apply(doc);
	for i := 0; i <= nv.Len(); i++ {
		nv.At(i).(*Node).Copy(n)
	}
	return
}
