package undgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"log/slog"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/und/undtag"
)

func generateConversionMethod(w io.Writer, data *replaceData, node *typeNode, exprMap map[string]fieldAstExprSet) (err error) {
	ts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
	plainTyName := ts.Name.Name + printTypeParamVars(ts)
	rawTyName, _ := strings.CutSuffix(ts.Name.Name, "Plain")
	rawTyName += printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		if err == nil {
			err = flush()
		}
	}()

	_generateConversionMethod(true, printf, plainTyName, rawTyName, ts, data, node, exprMap)
	_generateConversionMethod(false, printf, plainTyName, rawTyName, ts, data, node, exprMap)

	return
}

func _generateConversionMethod(
	toPlain bool,
	printf func(format string, args ...any),
	plainTyName, rawTyName string,
	ts *dst.TypeSpec,
	data *replaceData,
	node *typeNode,
	exprMap map[string]fieldAstExprSet,
) {
	printf(`func (v %s) %s() %s {
`,
		or(
			toPlain,
			[]any{rawTyName, "UndPlain", plainTyName},
			[]any{plainTyName, "UndRaw", rawTyName},
		)...,
	)
	defer printf(`}

`)

	named := node.typeInfo
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		_generateConversionMethodElemTypes(toPlain, printf, node, data.importMap, exprMap)
	case *types.Struct:
		_generateMethodToRawStructFields(toPlain, printf, ts, node, rawTyName, plainTyName, data.importMap, exprMap)
	default:
		slog.Default().Error(
			"implementation error",
			slog.String("rawTyName", rawTyName),
			slog.String("plainTyName", plainTyName),
			slog.Any("type", named),
		)
		panic("implementation error")
	}
}

func _generateConversionMethodElemTypes(
	toPlain bool,
	printf func(format string, args ...any),
	node *typeNode,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	conversionIndent, _ := importMap.Ident(UndPathConversion)

	_, edge := firstTypeIdent(node.children) // must be only one.

	rawExpr := typeToAst(
		edge.parentNode.typeInfo.Underlying(),
		edge.parentNode.typeInfo.Obj().Pkg().Path(),
		importMap,
	)

	var plainExpr ast.Expr
	for _, v := range exprMap {
		plainExpr = v.Wrapped
	}

	unwrapper := unwrapFieldAlongPath(
		or(toPlain, rawExpr, plainExpr),
		or(toPlain, plainExpr, rawExpr),
		edge,
		0,
	)

	if isUndType(edge.childType) {
		// matched, wrapped implementor
		printf(`return ` + unwrapper(
			func(s string) string {
				return fmt.Sprintf(
					`%s.Map(
						%s,
						%s.%s,
					)`,
					importIdent(namedTypeToTargetType(edge.childType), importMap),
					s,
					conversionIndent,
					or(toPlain, "ToPlain", "ToRaw"),
				)
			},
			"v",
		) + `
`)
		return
	} else {
		// implementor
		printf(`return ` + unwrapper(
			func(s string) string {
				return fmt.Sprintf(
					`%s.%s()`,
					s,
					or(toPlain, "UndPlain", "UndRaw"),
				)
			},
			"v",
		) + `
`)
		return
	}

}

func generateConversionMethodDirect(toPlain bool, edge typeDependencyEdge, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (convert func(ident string) string, needsArg bool) {
	matchUndTypeBool(
		namedTypeToTargetType(edge.childType),
		false,
		func() {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return optionToPlain(undOpt) },
				func() (func(ident string) string, bool) { return optionToRaw(undOpt, typeParam, importMap) },
			)()
		},
		func(isSlice bool) {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return undToPlain(undOpt, importMap) },
				func() (func(ident string) string, bool) { return undToRaw(isSlice, undOpt, typeParam, importMap) },
			)()
		},
		func(isSlice bool) {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return elasticToPlain(isSlice, undOpt, typeParam, importMap) },
				func() (func(ident string) string, bool) { return elasticToRaw(isSlice, undOpt, typeParam, importMap) },
			)()
		},
	)

	if edge.hasSingleNamedTypeArg(isUndConversionImplementor) {
		conversionIdent, _ := importMap.Ident(UndPathConversion)
		pkgIdent := importIdent(namedTypeToTargetType(edge.childType), importMap)
		inner := convert
		convert = or(
			toPlain,
			func(ident string) string {
				return inner(fmt.Sprintf(
					`%s.Map(
				%s,
				%s.%s,
			)`,
					pkgIdent, ident, conversionIdent, or(toPlain, "ToPlain", "ToRaw"),
				))
			},
			func(ident string) string {
				return fmt.Sprintf(
					`%s.Map(
				%s,
				%s.%s,
			)`,
					pkgIdent, inner(ident), conversionIdent, or(toPlain, "ToPlain", "ToRaw"),
				)
			},
		)
		needsArg = true
	}
	return
}
