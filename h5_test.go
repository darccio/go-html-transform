package h5

import (
	"testing"
	"testing/util"
)

func TestPushNode(t *testing.T) {
	p := new(Parser)
	util.AssertTrue(t, p.Top == nil, "Top is not nil")
	util.AssertTrue(t, p.curr == nil, "curr is not nil")
	top := pushNode(p)
	top.data = append(top.data, []int("foo")...)
	util.AssertTrue(t, p.Top != nil, "Top is still nil")
	util.AssertTrue(t, p.curr != nil, "curr is stil nil")
	util.AssertEqual(t, p.Top, top)
	util.AssertEqual(t, p.curr, top)
	next := pushNode(p)
	next.data = append(next.data, []int("bar")...)
	util.AssertEqual(t, len(top.Children), 1)
	util.AssertEqual(t, p.Top, top)
	util.AssertEqual(t, p.curr, next)
	util.AssertEqual(t, p.curr.Parent, p.Top)
}

func TestPopNode(t *testing.T) {
	p := new(Parser)
	top := pushNode(p)
	top.data = append(top.data, []int("foo")...)
	next := pushNode(p)
	next.data = append(next.data, []int("bar")...)
	popped := popNode(p)
	util.AssertEqual(t, popped, top)
	util.AssertEqual(t, p.Top, p.curr)
}

func TestAddSibling(t *testing.T) {
	p := new(Parser)
	top := pushNode(p)
	top.data = append(top.data, []int("foo")...)
	next := pushNode(p)
	next.data = append(next.data, []int("bar")...)
	sib := addSibling(p)
	sib.data = append(sib.data, []int("baz")...)
	util.AssertEqual(t, len(top.Children), 2)
	util.AssertEqual(t, top.Children[0], next)
	util.AssertEqual(t, top.Children[1], sib)
}
