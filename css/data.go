// Package css implements a css level 3 parser as described at
// http://www.w3.org/TR/css-syntax-3/
package css

import "go.marzhillstudios.com/pkg/go-html-transform/css/selector"

// Stylesheet is a list of Statements
type Stylesheet struct {
	Statements []Statement
}

// Statement is either a Ruleset or an AtRule or a comment.
type Statement struct {
	*Ruleset
	*AtRule
	*Comment
	*HtmlComment
}

// AtRule is an AtKeyword an optional param and an SimpleBlock.
// http://www.w3.org/TR/css-syntax-3/#consume-an-at-rule0
type AtRule struct {
	AtKeyword   string
	Param       []string
	SimpleBlock *SimpleBlock
}

// Ruleset is a selector followed by a Declaration Block
type Ruleset struct {
	Selector *selector.Chain
	DeclarationList
}

// BlockItem contains one and only one of DeclarationList, *Ruleset, or *AtRule
type BlockItem struct {
	DeclarationList
	*Ruleset
	*AtRule
}

// SimpleBlock contains a list of BlockItems
type SimpleBlock struct {
	Content []BlockItem
}

// DeclarationList is a list of Declarations forming a block.
type DeclarationList []Declaration

// Declaration is a Property and Value pair.
type Declaration struct {
	Property string
	Value    string
}

type Comment string
type HtmlComment string
