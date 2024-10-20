package undgen

import (
	"iter"
	"slices"

	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

func GenerateValidator(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkg *packages.Package,
	imports []TargetImport,
	targetTypeNames ...string,
) error {
	for _, err := range parseStructTag(pkg, imports, targetTypeNames...) {
		if err != nil {
			return err
		}
	}
	return nil
}

func parseStructTag(pkg *packages.Package, imports []TargetImport, targetTypeNames ...string) iter.Seq2[replaceData, error] {
	return func(yield func(replaceData, error) bool) {
		for data, err := range generatorIter(
			imports,
			findTypes(pkg, targetTypeNames...),
		) {
			if err != nil {
				if !yield(data, err) {
					return
				}
				continue
			}

			var firstErr error
			data.targets = slices.Collect(
				xiter.Filter(
					func(t replacerTarget) bool {
						return t.mt.IsSomeAnd(func(rmt RawMatchedType) bool {
							if len(rmt.Field) == 0 {
								return false
							}
							return hiter.Any(
								func(f MatchedField) bool {
									if firstErr == nil && f.Tag.IsSomeAnd(func(r UndTagParseResult) bool { return r.Err != nil }) {
										firstErr = f.Tag.Value().Err
									}
									return f.Tag.IsSome()
								},
								slices.Values(rmt.Field),
							)
						})
					},
					slices.Values(data.targets),
				),
			)
			if firstErr != nil {
				if !yield(data, firstErr) {
					return
				}
				continue
			}

			addMissingImports(data.df, data.importMap)

			if !yield(data, nil) {
				return
			}
		}
	}
}
