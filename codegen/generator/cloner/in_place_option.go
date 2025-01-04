package cloner

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"sync"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/msg"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/option"
)

var (
	ErrUnknownDirective = errors.New("unknown directive")
)

const (
	DirectivePrefix = "cloner:"
)

type InPlaceOptionKind string

const (
	InPlaceOptionKindAssign  InPlaceOptionKind = "assign"
	InPlaceOptionKindIgnore  InPlaceOptionKind = "ignore"
	InPlaceOptionKindCopyPtr InPlaceOptionKind = "copyptr"
	InPlaceOptionKindMake    InPlaceOptionKind = "make"
)

var exclusiveDirectives = [...]InPlaceOptionKind{
	InPlaceOptionKindAssign,
	InPlaceOptionKindIgnore,
	InPlaceOptionKindCopyPtr,
	InPlaceOptionKindMake,
}

type ClonerOverlay struct {
	MatcherConfig MatcherConfigOverlay
	AllowList     []string
	DenyList      []string
	PerPkg        map[string]PkgOverlay
}

type MatcherConfigOverlay struct {
	NoCopyHandle    CopyHandle
	ChannelHandle   CopyHandle
	FuncHandle      CopyHandle
	InterfaceHandle CopyHandle
}

type PkgOverlay struct {
	MatcherConfig MatcherConfigOverlay
	AllowList     []string
	DenyList      []string
	PerType       map[string]TypeOverlay
}

// TypeOverlay defines per-type configuration.
type TypeOverlay struct {
	// true if the type is ignored.
	Ignored bool
	// TypeOption defines OverlayOption for the type itself.
	// Effective only when the type is based on map, slice, array.
	TypeOption option.Option[*FieldOverlay] `json:",omitzero"`
	Fields     map[string]FieldOverlay      `json:",omitzero"`
}

type FieldOverlay struct {
	Pos  option.Option[int]
	Kind InPlaceOptionKind
	// TODO: add custom cloner
}

func (o FieldOverlay) shouldIgnoreMatcher() bool {
	return o.Kind == InPlaceOptionKindAssign
}

func (o FieldOverlay) overlay(c *MatcherConfig) *MatcherConfig {
	// TODO: add custom cloner anyhow
	cc := c.fallback()

	assign := func(h CopyHandle) {
		cc.NoCopyHandle = h
		cc.ChannelHandle = h
		cc.FuncHandle = h
		cc.InterfaceHandle = h
	}

	var h CopyHandle
	switch o.Kind {
	default:
		return cc
	case InPlaceOptionKindIgnore:
		h = CopyHandleIgnore
	case InPlaceOptionKindCopyPtr:
		h = CopyHandleCopyPointer
	case InPlaceOptionKindMake:
		h = CopyHandleMake
	}

	assign(h)

	return cc
}

func parseNode(n *typegraph.Node) (any, error) {
	dec, _, err := loadOrParseFile(n.Pkg.Fset, n.File)
	if err != nil {
		return nil, err
	}

	dts := dec.Dst.Nodes[n.Ts].(*dst.TypeSpec)

	st := unwrapStructType(dts)
	if st == nil {
		return nil, nil
	}
	lines := make(map[string]FieldOverlay)
	for i, f := range hiter.Enumerate(codegen.FieldDst(st)) {
		fieldDirection := codegen.FieldDirectiveCommentDst(
			DirectivePrefix,
			f.Field.Decs,
		)
		if len(fieldDirection) == 0 {
			continue
		}
		direction, err := parseLines(fieldDirection)
		if err != nil {
			return nil, fmt.Errorf(
				"parsing %s: %w",
				msg.PkgPathPrefixedName(n.Type), err,
			)
		}

		direction.Pos = option.Some(i)
		lines[f.Name] = direction
	}

	return TypeOverlay{Fields: lines}, nil
}

var (
	decMap         sync.Map
	parseResultMap sync.Map
)

type parserResult struct {
	once sync.Once
	df   *dst.File
	err  error
}

func loadOrParseFile(fset *token.FileSet, file *ast.File) (*decorator.Decorator, *dst.File, error) {
	dec := loadOrStoreDec(fset)
	v, ok := parseResultMap.Load(file)
	if !ok {
		v, _ = parseResultMap.LoadOrStore(file, &parserResult{})
	}
	r := v.(*parserResult)
	r.once.Do(func() {
		r.df, r.err = dec.DecorateFile(file)
	})
	return dec, r.df, r.err
}

func loadOrStoreDec(fset *token.FileSet) *decorator.Decorator {
	v, ok := decMap.Load(fset)
	if !ok {
		dec := decorator.NewDecorator(fset)
		v, _ = decMap.LoadOrStore(fset, dec)
	}
	return v.(*decorator.Decorator)
}

// unwrapStructType unwraps ts to *dst.StructType.
// If ts's Type is wrapped in one or more parentheses, they'll be removed.
// If ts's Type is other than struct-type then it returns nil.
func unwrapStructType(ts *dst.TypeSpec) *dst.StructType {
	unwrapped := ts.Type

	var loopCount int
	const maxDepth = 100
	for {
		if loopCount >= maxDepth {
			panic("too deep parentheses")
		}
		paren, ok := unwrapped.(*dst.ParenExpr)
		if !ok {
			break
		}
		unwrapped = paren.X
		loopCount++
	}
	st, ok := unwrapped.(*dst.StructType)
	if !ok {
		return nil
	}
	return st
}

func parseLines(lines []string) (FieldOverlay, error) {
	var parsed FieldOverlay
	for _, line := range lines {
		for _, directive := range strings.Split(line, ",") {
			idx := slices.Index(exclusiveDirectives[:], InPlaceOptionKind(directive))
			switch {
			default:
				return parsed, fmt.Errorf("%w: %q", ErrUnknownDirective, directive)
			case idx >= 0:
				parsed.Kind = exclusiveDirectives[idx]
			}
		}
	}
	return parsed, nil
}
