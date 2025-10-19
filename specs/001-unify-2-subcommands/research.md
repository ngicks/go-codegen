# Research: Unified Command Structure

**Feature**: 001-unify-2-subcommands
**Date**: 2025-10-18
**Status**: Complete

## Overview

This document consolidates research findings for implementing the automark and autoimpl commands, including directive format design, AST manipulation strategies, parser design, and deprecation planning.

## 1. Directive Comment Format Design

### Decision: Extend Existing Format with Type-Level Marker

**Format**: Use existing `//generatorname:directive` syntax but add a special type-level marker

**Rationale**:
- The codebase already has `codegen/pkg/directive` package with robust parsing
- Existing format: `//cloner:copyptr` (field-level directives)
- Need to distinguish between:
  - **Type-level marks** (added by automark): Indicates this type needs generation
  - **Field-level directives** (existing usage): Customizes specific field behavior

**Proposed Format**:
```go
// Type-level marker (added by automark)
//codegen:cloner
type MyStruct struct {
    // Field-level directive (existing usage, can be manual or automark-generated)
    //cloner:copyptr
    Mutex *sync.Mutex
}
```

**Key Design Points**:
1. **Type-level marker**: `//codegen:generatortype` (e.g., `//codegen:cloner`, `//codegen:und:patch`)
2. **Placement**: Immediately above type declaration
3. **Backward Compatibility**: Existing field-level directives continue to work
4. **Multi-generator support**: Multiple markers for same type (e.g., both cloner and und)

```go
//codegen:cloner
//codegen:und:patch
type User struct {
    Name string
    //cloner:copyptr
    Mutex *sync.Mutex
}
```

**Alternatives Considered**:
- **Structured comment block**: More verbose, harder to parse
- **Build tags**: Not visible to AST manipulation tools
- **Separate marker file**: Defeats purpose of directive-driven approach

**Configuration Embedding**:
For options that would normally be command-line flags:
```go
//codegen:cloner no-copy:copy chan:make
type ComplexStruct struct {
    // ...
}
```

This allows automark to embed default configuration while keeping manual override possible.

## 2. AST Modification for Comment Insertion

### Decision: Use github.com/dave/dst Decorations System

**Approach**: Leverage `dst` (decorated syntax tree) which preserves all formatting and comments

**Implementation Strategy**:

1. **Loading with Decorations**:
```go
// Existing pattern in codebase
import "github.com/dave/dst/decorator"

pkgs, err := decorator.Load(cfg)
// Provides dst.File with all decorations preserved
```

2. **Comment Insertion**:
```go
// Insert type-level marker
func AddTypeMarker(typeSpec *dst.TypeSpec, generator string) {
    marker := fmt.Sprintf("//codegen:%s", generator)

    if typeSpec.Decorations().Start == nil {
        typeSpec.Decorations().Start = dst.Decorations{}
    }

    // Check for existing marker to avoid duplication
    existing := FindExistingMarker(typeSpec.Decorations().Start, "codegen")
    if existing {
        return // Idempotent
    }

    typeSpec.Decorations().Start.Prepend("//", marker)
}
```

3. **Writing Back**:
```go
// Use dst/decorator.Restorer to write modified AST
import "github.com/dave/dst/decorator"

r := decorator.NewRestorer()
err = r.Fprint(file, modifiedFile)
```

**Preserving Existing Comments**:
- dst automatically preserves all decorations
- Check for existing markers before adding new ones
- Maintain original formatting and spacing

**Detecting Existing Directive Comments**:
```go
func HasTypeMarker(typeSpec *dst.TypeSpec, generator string) bool {
    prefix := "//codegen:" + generator
    for _, dec := range typeSpec.Decorations().Start.All() {
        if strings.HasPrefix(strings.TrimSpace(dec), prefix) {
            return true
        }
    }
    return false
}
```

**Best Practices from Codebase**:
- The project already uses `dst` extensively for code generation
- Pattern: Load → Modify → Restore maintains formatting
- Decorations API handles all comment/whitespace preservation

**Alternatives Considered**:
- **go/ast** + format: Loses comments and custom formatting
- **Text-based replacement**: Fragile, not type-aware
- **go/printer**: Doesn't preserve original formatting

## 3. Parser Design for autoimpl

### Decision: Leverage Existing directive.Parse with Type-Level Scanning

**Approach**: Extend existing `codegen/pkg/directive` package patterns

**Implementation Strategy**:

1. **Type-Level Marker Detection**:
```go
// In autoimpl, scan for type-level markers
func ScanForMarkedTypes(pkg *packages.Package) []MarkedType {
    var marked []MarkedType

    for _, file := range pkg.Syntax {
        ast.Inspect(file, func(n ast.Node) bool {
            if typeSpec, ok := n.(*ast.TypeSpec); ok {
                if marker := parseTypeMarker(typeSpec.Doc); marker != nil {
                    marked = append(marked, MarkedType{
                        TypeSpec: typeSpec,
                        Generators: marker.Generators,
                        Config: marker.Config,
                    })
                }
            }
            return true
        })
    }

    return marked
}
```

2. **Parsing Type Markers**:
```go
func parseTypeMarker(doc *ast.CommentGroup) *TypeMarker {
    // Use existing directive.ParseAst
    parsed, ok := directive.ParseAst(doc, "codegen")
    if !ok {
        return nil
    }

    // Extract generator types from parsed directive
    // e.g., "cloner", "und:patch", "und:plain"
    generators := extractGenerators(parsed)
    config := extractConfig(parsed)

    return &TypeMarker{
        Generators: generators,
        Config: config,
    }
}
```

3. **Generator Dispatch**:
```go
type GeneratorRegistry map[string]GeneratorFunc

func (r GeneratorRegistry) Dispatch(marked MarkedType, pkgs []*packages.Package) error {
    for _, genType := range marked.Generators {
        genFunc, ok := r[genType]
        if !ok {
            return fmt.Errorf("unknown generator: %s", genType)
        }

        if err := genFunc(marked, pkgs, marked.Config); err != nil {
            return err
        }
    }
    return nil
}
```

**Error Handling for Malformed Directives**:
- Existing `directive.Parse` handles malformed key=value pairs gracefully
- Add validation for known generator types
- Fail fast with file:line information from ast.Position

**Mapping to Generators**:
```go
var DefaultRegistry = GeneratorRegistry{
    "cloner": func(marked MarkedType, pkgs []*packages.Package, cfg map[string]string) error {
        // Invoke cloner.Generator with config
        return invokeCloner(marked, pkgs, cfg)
    },
    "und:patch": func(marked MarkedType, pkgs []*packages.Package, cfg map[string]string) error {
        return invokeUndPatch(marked, pkgs, cfg)
    },
    "und:plain": func(marked MarkedType, pkgs []*packages.Package, cfg map[string]string) error {
        return invokeUndPlain(marked, pkgs, cfg)
    },
    "und:validator": func(marked MarkedType, pkgs []*packages.Package, cfg map[string]string) error {
        return invokeUndValidator(marked, pkgs, cfg)
    },
}
```

**Backward Compatibility**:
- Existing field-level directives parsed exactly as before
- Type-level markers are additive, don't break existing parsing
- Manual directives work identically to automark-generated ones

**Alternatives Considered**:
- **Regex-based parsing**: Fragile, error-prone
- **Custom parser from scratch**: Reinventing wheel, directive package exists
- **JSON in comments**: Not idiomatic Go, harder to read/write manually

## 4. Deprecation Strategy

### Decision: Gradual Deprecation with Clear Migration Path

**Timeline**:
- **Phase 1 (This Release)**: Add automark/autoimpl alongside existing commands
- **Phase 2 (Next Release)**: Mark old commands as deprecated in help text
- **Phase 3 (Future Release)**: Remove deprecated commands

**Migration Path**:

**Old Command → New Command Mapping**:
```bash
# OLD: cloner with flags
codegen cloner --dir ./... --no-copy-copy --chan-make

# NEW: automark then autoimpl
codegen automark ./... --generator cloner --no-copy copy --chan make
codegen autoimpl ./...
```

**For und generators**:
```bash
# OLD: undgen subcommands
codegen undgen patch --dir ./pkg
codegen undgen plain --dir ./pkg
codegen undgen validator --dir ./pkg

# NEW: Mark once, generate all
codegen automark ./pkg --generator und:patch --generator und:plain --generator und:validator
codegen autoimpl ./pkg
```

**Help Text Strategy**:

1. **Phase 1**: Show both command sets, mark new as "(recommended)"
2. **Phase 2**: Add deprecation warnings:
```
DEPRECATED: Use 'codegen automark' and 'codegen autoimpl' instead.
See 'codegen automark --help' for migration guide.
```

3. **Phase 3**: Remove old commands, provide error with migration instructions

**Documentation Updates**:
- README: Add "Migration Guide" section
- Help text: Include examples of old → new command equivalents
- Changelog: Clear communication about deprecation timeline

**Backward Compatibility Guarantees**:
- Generated code output remains identical
- Existing directive comments work without modification
- Tests continue to pass with both old and new commands

**Alternatives Considered**:
- **Immediate removal**: Too disruptive for existing users
- **Permanent dual commands**: Increases maintenance burden
- **Automatic migration tool**: Over-engineering for simple command changes

## 5. Additional Research Findings

### Flag Design for automark

**Required Flags**:
- `--generator` / `-g`: Specify generator type (cloner, und:patch, etc.)
  - Can be specified multiple times for multiple generators
- `--dry-run`: Preview changes without modifying files
- Standard flags: `--dir`, `--pkg`, `--verbose`

**Configuration Flags** (passed to generators):
```bash
codegen automark ./... -g cloner \
  --no-copy copy \
  --chan make \
  --interface copy
```

These flags get embedded in the type-level marker as configuration.

### Flag Design for autoimpl

**Required Flags**:
- Standard flags: `--dir`, `--pkg`, `--verbose`, `--dry`, `--ignore-generated`
- No generator-specific flags (reads from directive comments)

**Behavior**:
- Scans for all marked types
- Dispatches to appropriate generators based on markers
- Existing generators handle all the actual code generation

### Performance Considerations

**automark Performance**:
- Similar to current matcher logic (type graph traversal)
- Additional AST modification + file write
- Target: < 5 seconds for 100 types (per spec)

**autoimpl Performance**:
- Comparable to current generators (just adds dispatch layer)
- Parallel generation possible (existing generators may support)

### Testing Strategy

**Integration Tests**:
```
codegen/internal/generationtests/automark_autoimpl/
├── input/
│   └── types.go          # Unmarked types
├── marked/
│   └── types.go          # After automark
└── generated/
    └── types.clone.go    # After autoimpl
```

**Test Flow**:
1. Run automark on input
2. Verify markers added correctly
3. Run autoimpl on marked
4. Verify generated code compiles and matches expected

## Conclusion

All research questions have been resolved with concrete decisions:

1. **Directive Format**: Type-level `//codegen:generatortype` markers
2. **AST Modification**: Use existing `dst` patterns
3. **Parser**: Extend existing `directive` package
4. **Deprecation**: Gradual three-phase approach

All decisions align with constitutional principles and leverage existing infrastructure. Ready to proceed to Phase 1 design.
