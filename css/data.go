package css

import (
	"code.google.com/p/go-html-transform/css/selector"
)

type Stylesheet struct {
	Contents []struct {
		Comment   Comment
		Statement *Statement
	}
}

type Statement struct {
	Ruleset *Ruleset
	AtRule  *AtRule
}

type AtRule struct {
	AtKeyword string
	Block
}

type Ruleset struct {
	Selector *selector.Chain
	Block    *Block
}

type Block struct {
	Body []Body
}

type Body struct {
	// Exclusively one of these?
	Block       *Block
	AtKeyword   string
	Declaration *Declaration
	BadString   string
	Comment     Comment
}

type Declaration struct {
	Property string
	Value    string
}

type Comment string
