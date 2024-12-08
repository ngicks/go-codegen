package typegraph

import "go/types"

type Option interface {
	apply(g *Graph) Option
}

type PrivParser func(external bool, parent, child *Node, childTy *types.Named) any

type privParserOption PrivParser

func (o privParserOption) apply(g *Graph) Option {
	old := g.privParser
	g.privParser = PrivParser(o)
	return privParserOption(old)
}

func WithPrivParser(privParser PrivParser) Option {
	return privParserOption(privParser)
}
