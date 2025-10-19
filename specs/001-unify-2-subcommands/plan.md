# Implementation Plan: Unified Command Structure

**Branch**: `001-unify-2-subcommands` | **Date**: 2025-10-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-unify-2-subcommands/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Refactor the codegen CLI structure to expose exactly two subcommands: `automark` (discovers and marks types with directive comments) and `autoimpl` (generates implementations for marked types). This introduces a two-phase workflow that separates type discovery from code generation, improving user experience and reducing repetitive command invocations.

**Technical Approach**: Leverage existing generator infrastructure (cloner, undgen) but create new CLI commands that coordinate them. automark will use matcher logic to identify eligible types and write directive comments to source files. autoimpl will parse these comments and delegate to appropriate generators.

## Technical Context

**Language/Version**: Go 1.25.0 (as specified in go.mod)
**Primary Dependencies**:
- `github.com/spf13/cobra` v1.9.1 (CLI framework)
- `github.com/dave/dst` v0.27.3 (AST manipulation with decorations)
- `golang.org/x/tools` v0.36.0 (package loading and analysis)
- `github.com/ngicks/und` v1.0.0-alpha8 (und type support)

**Storage**: Source file modification (directive comments), generated Go files with specific suffixes
**Testing**: `gotest.tools/v3` for assertions, `internal/generationtests` and `internal/tests` for generation tests, compilation verification required
**Target Platform**: Cross-platform CLI tool (Linux, macOS, Windows)
**Project Type**: Single project - command-line code generator
**Performance Goals**: automark < 5 seconds for 100 types, autoimpl comparable to current generators
**Constraints**:
- Must maintain byte-for-byte compatibility with existing generators
- Must preserve backward compatibility with existing directive comments
- Must not break existing codebases
- Security: relative paths only, no path traversal

**Scale/Scope**:
- Refactor existing 5 commands (cloner, undgen + 3 subcommands) into 2 commands
- Support all existing generator types (clone, und patch/plain/validator)
- Handle single and multi-package operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Generation as Framework Infrastructure ✅
**Status**: COMPLIANT
- automark and autoimpl will reuse existing generator infrastructure (cloner, undgen packages)
- New directive comment parsing utilities will be extracted as reusable components
- No modification to core generator logic required

### II. Type Safety and AST-Based Generation ✅
**Status**: COMPLIANT
- automark will use existing package loading with full type information
- Directive comment writing will use AST manipulation via `github.com/dave/dst`
- autoimpl delegates to existing AST-based generators
- Type graph analysis remains unchanged

### III. Directive-Driven Customization ✅
**Status**: COMPLIANT
- automark generates directive comments in the established `//generatorname:directivename` format
- autoimpl respects existing directive syntax (backward compatible)
- Manual directive editing fully supported
- Field-level directives continue to override global config

### IV. Generated Code Discipline ✅
**Status**: COMPLIANT
- autoimpl uses existing suffixwriter for consistent file naming
- Generated files retain standard "DO NOT EDIT" header
- Suffixes unchanged: `.clone.go`, `.und_patch.go`, `.und_plain.go`, `.und_validator.go`
- goimports processing maintained

### V. Security Constraints for Path Handling ✅
**Status**: COMPLIANT
- Both commands use existing path validation (relative paths only)
- automark validates paths before file modification
- autoimpl inherits existing security constraints
- Working directory root enforcement maintained

### VI. Testing with Generated Artifacts (NON-NEGOTIABLE) ✅
**Status**: COMPLIANT
- Will add integration tests in `internal/generationtests`
- Tests will verify automark → autoimpl pipeline produces correct output
- Generated code must compile successfully
- Test workflow: mark types → generate → compile → verify

### VII. Matcher System for Selective Generation ✅
**Status**: COMPLIANT
- automark leverages existing MatcherConfig logic
- Type eligibility rules remain consistent with current generators
- autoimpl reads directive comments to determine what to generate
- Default-to-safe behavior preserved

**GATE RESULT**: ✅ ALL PRINCIPLES COMPLIANT - Proceed with Phase 0

## Project Structure

### Documentation (this feature)

```
specs/001-unify-2-subcommands/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── automark-interface.md    # automark command contract
│   └── autoimpl-interface.md    # autoimpl command contract
└── checklists/
    └── requirements.md  # Specification validation checklist
```

### Source Code (repository root)

This is a refactoring of existing code, not new source structure. Key modifications:

```
codegen/
├── cmd/
│   ├── root.go          # Base command (unchanged)
│   ├── automark.go      # NEW: Discovery and marking command
│   ├── autoimpl.go      # NEW: Implementation generation command
│   ├── cloner.go        # DEPRECATED: Will be replaced by automark/autoimpl
│   ├── undgen.go        # DEPRECATED: Will be replaced by automark/autoimpl
│   └── undgen_*.go      # DEPRECATED: Subcommands replaced
├── pkg/
│   ├── automark/        # NEW: Marking logic package
│   │   ├── marker.go    # Type discovery and directive writing
│   │   ├── config.go    # Marking configuration
│   │   └── internal/
│   │       └── tests/   # Marking tests
│   └── autoimpl/        # NEW: Impl coordination package
│       ├── parser.go    # Directive comment parsing
│       ├── dispatcher.go # Generator dispatch logic
│       └── internal/
│           └── tests/   # Implementation tests
├── generator/
│   ├── cloner/          # UNCHANGED: Core generator logic
│   └── undgen/          # UNCHANGED: Core generator logic
└── internal/
    └── generationtests/ # NEW TESTS: automark→autoimpl integration tests
```

**Structure Decision**: This is a command structure refactoring within the existing single-project layout. New packages (`pkg/automark`, `pkg/autoimpl`) will be added to house the mark-and-dispatch logic, while existing generator packages remain untouched. The refactoring isolates changes to the CLI layer and adds coordinating packages without modifying core generation logic.

## Complexity Tracking

*No constitutional violations - this section is empty.*

## Phase 0: Research (Next Step)

Research tasks to be completed in `research.md`:

1. **Directive Comment Format Design**
   - Research: What structured format should automark use for directive comments?
   - Research: How to ensure backward compatibility with existing manually-written directives?
   - Research: Format for embedding generator type and configuration options

2. **AST Modification for Comment Insertion**
   - Research: Best practices for inserting comments via `github.com/dave/dst`
   - Research: Preserving existing comments and formatting
   - Research: Detecting and updating existing directive comments

3. **Parser Design for autoimpl**
   - Research: Reliable parsing of directive comments from AST
   - Research: Error handling for malformed directives
   - Research: Mapping directive content to generator dispatch

4. **Deprecation Strategy**
   - Research: Migration path from old commands to automark/autoimpl
   - Research: Documentation and help text strategy
   - Research: Timeline for removing deprecated commands

## Phase 1: Design & Contracts (After Research)

Design artifacts to be created:

1. **data-model.md**: Command entities (automark config, autoimpl dispatch, directive structure)
2. **contracts/**: Command interfaces and flag specifications
3. **quickstart.md**: Usage examples for mark-then-generate workflow

## Phase 2: Task Generation (Not part of this command)

Will be handled by `/speckit.tasks` command after plan is complete.

---

## Post-Design Constitution Check ✅

*Re-evaluation after completing Phase 0 and Phase 1*

### All Principles: COMPLIANT ✅

**No new concerns identified**. The detailed design reinforces initial assessment:

1. **Framework Infrastructure** ✅: New `pkg/automark` and `pkg/autoimpl` packages follow existing patterns
2. **Type Safety & AST** ✅: Research confirms use of existing `dst` and `directive` packages
3. **Directive-Driven** ✅: Type-level `//codegen:` markers extend existing format without breaking changes
4. **Generated Code Discipline** ✅: autoimpl delegates to existing generators, maintaining all standards
5. **Security** ✅: Both commands inherit existing path validation
6. **Testing** ✅: Integration tests planned for mark→impl pipeline
7. **Matcher System** ✅: automark leverages existing matcher logic

**Quality Gates**: All gates remain satisfied. Design phase introduced no constitutional violations.

**Complexity Tracking**: Still empty - no violations to justify.

**FINAL GATE RESULT**: ✅ APPROVED FOR TASK GENERATION

The design is constitutionally sound and ready for implementation tasks.
