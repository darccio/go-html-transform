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

func TestBogusCommentHandlerNoEOF(t *testing.T) {
	p := NewParserFromString("foo comment >")
	top := pushNode(p)
	pushNode(p)
	st, err := bogusCommentHandler(p)
	util.AssertEqual(t, len(top.Children), 2)
	util.AssertEqual(t, string(top.Children[1].data), "foo comment ")
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertTrue(t, err == nil, "err is not nil")
}

// TODO error cases
func TestBogusCommentHandlerEOF(t *testing.T) {
	p := NewParserFromString("foo comment")
	top := pushNode(p)
	pushNode(p)
	st, err := bogusCommentHandler(p)
	util.AssertEqual(t, len(top.Children), 2)
	util.AssertEqual(t, string(top.Children[1].data), "foo comment")
	util.AssertTrue(t, st == nil, "next state handler is not nil")
	util.AssertTrue(t, err != nil, "err is nil")
}

// TODO the tag name too short case
// TODO the tag name too long case
// TODO the tag name different
func TestEndTagOpenHandlerOk(t *testing.T) {
	p := NewParserFromString("FoO>")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertEqual(t, err, nil)
	util.AssertTrue(t, err == nil, "err is not nil")
	util.AssertTrue(t, p.curr == nil, "did not pop node")
}

func TestEndTagOpenHandlerTrunc(t *testing.T) {
	p := NewParserFromString("fo>")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st == nil, "next state handler is not nil")
	util.AssertTrue(t, err != nil, "err is nil")
	util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerLong(t *testing.T) {
	p := NewParserFromString("fooo>")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st == nil, "next state handler is not nil")
	util.AssertTrue(t, err != nil, "err is nil")
	util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerWrong(t *testing.T) {
	p := NewParserFromString("bar>")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st == nil, "next state handler is not nil")
	util.AssertTrue(t, err != nil, "err is nil")
	util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerBogusComment(t *testing.T) {
	p := NewParserFromString("f\no>")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st != nil, "next state handler is not nil")
	util.AssertTrue(t, err != nil, "err is nil")
	util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerEOF(t *testing.T) {
	p := NewParserFromString("foo")
	curr := pushNode(p)
	curr.data = []int("foo")
	util.AssertTrue(t, p.curr != nil, "curr is not nil")
	st, err := endTagOpenHandler(p)
	util.AssertTrue(t, st == nil, "next state handler is nil")
	util.AssertTrue(t, err != nil, "err is nil")
	util.AssertEqual(t, p.curr, curr)
}

func TestTagNameHandler(t *testing.T) {
	p := NewParserFromString("f> ")
	curr := pushNode(p)
	st, err := handleChar(tagNameHandler)(p)
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertTrue(t, err == nil, "err is not nil")
	util.AssertEqual(t, curr.data[0], 'f')
	st, err = handleChar(tagNameHandler)(p)
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertTrue(t, err == nil, "err is not nil")
	util.AssertEqual(t, curr.data[0], 'f')
	p = NewParserFromString("F")
	curr = pushNode(p)
	st, err = handleChar(tagNameHandler)(p)
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertTrue(t, err == nil, "err is not nil")
	util.AssertEqual(t, curr.data[0], 'f')
}

func TestTagOpenHandler(t *testing.T) {
	p := NewParserFromString("")
	st := tagOpenHandler(p, 'f')
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertEqual(t, p.curr.data[0], 'f')
	util.AssertEqual(t, p.curr.Type, ElementNode)
}

func TestTagOpenHandlerEndTag(t *testing.T) {
	p := NewParserFromString("")
	st := tagOpenHandler(p, '/')
	util.AssertTrue(t, st != nil, "next state handler is nil")
	util.AssertTrue(t, p.curr.data == nil, "data is currently nil")
}
