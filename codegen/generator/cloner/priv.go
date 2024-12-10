package cloner

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/typegraph"
)

const (
	DirectivePrefix         = "cloner:"
	DirectiveCommentIgnore  = "ignore"
	DirectiveCommentCopyPtr = "copyptr"
)

var (
	ErrUnknownDirective = errors.New("unknown directive")
)

type clonerPriv struct {
	lines map[int]direction
}

type direction struct {
	Pos     int
	Ignore  bool
	CopyPtr bool
}

func (d direction) override(c MatcherConfig) MatcherConfig {
	switch {
	case d.Ignore:
		c.ChannelHandle = NoCopyHandleIgnore
		c.NoCopyHandle = NoCopyHandleIgnore
	case d.CopyPtr:
		c.ChannelHandle = NoCopyHandleCopyPointer
		c.NoCopyHandle = NoCopyHandleCopyPointer
	}
	return c
}

func parseNode(n *typegraph.Node) (any, error) {
	// store in global cache.
	dec := decorator.NewDecorator(n.Pkg.Fset)
	_, err := dec.DecorateFile(n.File)
	if err != nil {
		panic(err)
	}

	dts := dec.Dst.Nodes[n.Ts].(*dst.TypeSpec)

	st, ok := dts.Type.(*dst.StructType)
	// ignoring cases for
	//   - *Ident(alias)
	//   - *ParenExpr(grouped, wtf?)
	//   - *SelectorExpr(type based on type defined in other packages)
	//   - *StarExpr(pointer type)
	if !ok {
		return nil, nil
	}
	lines := make(map[int]direction)
	for i, f := range st.Fields.List {
		lineDirective, ok, err := codegen.ParseFieldDirectiveCommentDst(
			DirectivePrefix,
			f.Decs,
			func(lines []string) (direction, error) {
				var parsed direction
				for _, line := range lines {
					for _, directive := range strings.Split(line, ",") {
						switch directive {
						case DirectiveCommentIgnore:
							parsed.Ignore = true
						case DirectiveCommentCopyPtr:
							parsed.CopyPtr = true
						default:
							return parsed, fmt.Errorf("%w: %q", ErrUnknownDirective, directive)
						}
					}
				}
				return parsed, nil
			},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parsing %q.%s: %w",
				n.Type.Obj().Pkg().Path(), n.Type.Obj().Name(), err,
			)
		}

		if ok {
			lineDirective.Pos = i
			lines[i] = lineDirective
		}
	}

	return clonerPriv{lines}, nil
}
