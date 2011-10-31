package h5

import (
	"fmt"
)

type Attribute struct {
	Name string
	Value string
}

func (a *Attribute) String() string {
	// TODO handle differnt quoting styles.
	return " " + a.Name + "='" + a.Value + "'"
}

func (a *Attribute) Clone() *Attribute {
	return &Attribute{Name: a.Name, Value: a.Value}
}

type NodeType int
const (
	TextNode NodeType = iota // zero value so the default
	ElementNode NodeType = iota
	DoctypeNode NodeType = iota
	CommentNode NodeType = iota
)

type Node struct {
	Type NodeType
	data []int
	Attr []*Attribute
	Parent *Node
	Children []*Node
	Public bool
	System bool
	Identifier []int
}

func (n *Node) SetData(rs []int) {
	n.data = rs
}

func attrString(attrs []*Attribute) string {
	if attrs == nil {
		return ""
	}
	s := ""
	for _, a := range attrs {
		s += fmt.Sprintf(" %s", a)
	}
	return s
}

func doctypeString(n *Node) string {
	keyword := ""
	identifier := string(n.Identifier)
	switch {
	case n.Public:
		keyword = "PUBLIC"
	case n.System:
		keyword = "SYSTEM"
	default:
		return "<!DOCTYPE html>"
	}
	return fmt.Sprintf("<!DOCTYPE %s=\"%s\">", keyword, identifier)
}

func (n *Node) String() string {
	switch n.Type {
	case TextNode:
		return n.Data()
	case ElementNode:
		// TODO handle the strange self close tags
		if n.Children == nil || len(n.Children) == 0 {
			name := n.Data()
			switch name {
			case "textarea":
				return "<textarea" + attrString(n.Attr) + "></textarea>"
			}
			return "<" + n.Data() + attrString(n.Attr) + "/>"
		} else {
			s :="<" + n.Data() + attrString(n.Attr) + ">"
			for _, n := range n.Children {
				s += n.String()
			}
			s += "</" + n.Data() + ">"
			return s
		}
	case DoctypeNode:
		// TODO Doctype stringification
		s := doctypeString(n)
		for _, n := range n.Children {
			s += n.String()
		}
		return s
	case CommentNode:
		// TODO
	}
	return ""
}

func (n *Node) Walk(f func(*Node)) {
	f(n)
	if len(n.Children) > 0 {
		for _, c := range n.Children {
			c.Walk(f)
		}
	}
}

func cloneNode(n, p *Node) *Node {
	clone := new(Node)
	clone.data = make([]int, len(n.data))
	clone.Attr = make([]*Attribute, len(n.Attr))
	clone.Children = make([]*Node, len(n.Children))
	clone.Parent = p
	clone.Type = n.Type
	clone.Public = n.Public
	clone.System = n.System
	clone.Identifier = n.Identifier
	copy(clone.data, n.data)
	for i, a := range n.Attr {
		clone.Attr[i] = a.Clone()
	}
	if len(n.Children) > 0 {
		for i, c := range n.Children {
			clone.Children[i] = cloneNode(c, n)
		}
	}
	return clone
}

func (n *Node) Clone() *Node {
	return cloneNode(n, n.Parent)
}

func (n *Node) Data() string {
	if n.data != nil {
		return string(n.data)
	}
	return ""
}

func Text(str string) *Node {
	return &Node{data:[]int(str)}
}
