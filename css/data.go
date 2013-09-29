package css

import "code.google.com/p/go-html-transform/css/selector"

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

// AtRule is an AtKeyword an optional param and an optional list of blocks.
type AtRule struct {
	AtKeyword string
	param     string
	Block     []*Block
}

// Ruleset is a selector followed by a Declaration Block
type Ruleset struct {
	Selector     *selector.Chain
	Declarations []Declaration
}

// Block is either a Ruleset or an AtRule or a Block
type Block struct {
	*Ruleset
	*AtRule
	*Block
}

// Declaration is a Property and Value pair.
type Declaration struct {
	Property string
	Value    string
}

type Comment string
type HtmlComment string
