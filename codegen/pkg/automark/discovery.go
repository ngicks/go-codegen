package automark

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/ngicks/go-codegen/codegen/pkg/directive"
	"github.com/ngicks/go-codegen/codegen/pkg/typegraph"
	"golang.org/x/tools/go/packages"
)

// DiscoverTypes discovers all type declarations in the loaded packages.
// It returns a slice of TypeMarker candidates (without generators assigned yet).
func DiscoverTypes(pkgs []*packages.Package, cfg *MarkerConfig) ([]TypeCandidate, error) {
	var candidates []TypeCandidate

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			// Get the file path
			filePath := pkg.Fset.File(file.Pos()).Name()

			// Inspect AST for type declarations
			ast.Inspect(file, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					return true
				}

				// Check if this declaration should be included
				// Note: ExcludeIgnoredGenDecl returns true to INCLUDE, false to EXCLUDE
				// (despite the confusing name, it's designed for typegraph filter)
				include, err := directive.ExcludeIgnoredGenDecl(genDecl)
				if err != nil || !include {
					return true
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					// Get type information
					obj := pkg.TypesInfo.Defs[typeSpec.Name]
					if obj == nil {
						continue
					}

					// Check if this type spec should be included
					// Note: ExcludeIgnoredTypeSpec returns true to INCLUDE, false to EXCLUDE
					include, err := directive.ExcludeIgnoredTypeSpec(typeSpec, obj)
					if err != nil || !include {
						continue
					}

					typeName := obj.Name()
					typeObj := obj.Type()

					// Get position information
					pos := pkg.Fset.Position(typeSpec.Pos())

					candidate := TypeCandidate{
						TypeName: typeName,
						Type:     typeObj,
						TypeSpec: typeSpec,
						GenDecl:  genDecl,
						Location: SourceLocation{
							FilePath: filePath,
							Package:  pkg.PkgPath,
							Line:     pos.Line,
							Column:   pos.Column,
						},
						IsExported: ast.IsExported(typeName),
					}

					candidates = append(candidates, candidate)
				}

				return true
			})
		}
	}

	if cfg.Verbose {
		fmt.Printf("Discovered %d type declaration(s)\n", len(candidates))
	}

	return candidates, nil
}

// TypeCandidate represents a type that might be marked for generation.
type TypeCandidate struct {
	TypeName   string
	Type       types.Type
	TypeSpec   *ast.TypeSpec
	GenDecl    *ast.GenDecl
	Location   SourceLocation
	IsExported bool
}

// BuildTypeGraph builds a type dependency graph for the packages.
// This can be used for more sophisticated type filtering based on relationships.
func BuildTypeGraph(pkgs []*packages.Package, cfg *MarkerConfig) (*typegraph.Graph, error) {
	// For now, we create a simple type graph
	// In the future, this could be used to mark dependent types automatically

	graph, err := typegraph.New(
		pkgs,
		func(node *typegraph.Node, external bool) (bool, error) {
			// Match all types - we'll filter later
			return true, nil
		},
		directive.ExcludeIgnoredGenDecl,
		directive.ExcludeIgnoredTypeSpec,
	)
	if err != nil {
		return nil, fmt.Errorf("building type graph: %w", err)
	}

	if cfg.Verbose {
		count := 0
		for range graph.EnumerateTypes() {
			count++
		}
		fmt.Printf("Type graph contains %d type(s)\n", count)
	}

	return graph, nil
}
