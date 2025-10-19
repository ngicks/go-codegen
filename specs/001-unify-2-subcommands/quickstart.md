# Quickstart: Unified Command Structure

**Feature**: 001-unify-2-subcommands
**Date**: 2025-10-18

## Overview

This quickstart guide demonstrates the two-phase workflow for code generation using the new `automark` and `autoimpl` commands. These commands replace the existing `cloner` and `undgen` commands with a unified, workflow-oriented interface.

## The Two-Phase Workflow

1. **Phase 1 - automark**: Discover and mark types with directive comments
2. **Phase 2 - autoimpl**: Generate implementations for marked types

This separation allows you to:
- Review which types will be generated before committing
- Mix automatic marking with manual directive editing
- Run generation multiple times without re-specifying configuration

## Quick Start (5 Minutes)

### Step 1: Mark types for generation

Mark all types in your package for clone generation:

```bash
codegen automark ./pkg/models -g cloner
```

**What this does**:
- Scans `pkg/models` for eligible types
- Adds `//codegen:cloner` comments above type declarations
- Modified source files with markers added

### Step 2: Generate implementations

Generate the actual code:

```bash
codegen autoimpl ./pkg/models
```

**What this does**:
- Scans for types marked with `//codegen:cloner`
- Invokes cloner generator
- Creates `*.clone.go` files with Clone methods

### Step 3: Verify results

```bash
# Check generated files
ls pkg/models/*.clone.go

# Verify compilation
go build ./pkg/models

# Run tests
go test ./pkg/models
```

**Done!** Your types now have generated clone methods.

## Common Workflows

### Workflow 1: Clone Generation with Configuration

**Scenario**: Generate clone methods with specific behavior for channels and no-copy types

```bash
# Step 1: Mark with configuration
codegen automark ./pkg/services \
  -g cloner \
  --no-copy copy \
  --chan make

# Step 2: Review markers (optional)
grep -B2 "//codegen:cloner" pkg/services/*.go

# Step 3: Generate
codegen autoimpl ./pkg/services

# Step 4: Test
go test ./pkg/services
```

**Result**: Clone methods that copy no-copy type pointers and make new channels

### Workflow 2: Und Generation (Multiple Variants)

**Scenario**: Generate patch, plain, and validator code for types using `github.com/ngicks/und`

```bash
# Step 1: Mark for all three generators
codegen automark ./pkg/models \
  -g und:patch \
  -g und:plain \
  -g und:validator

# Step 2: Generate all at once
codegen autoimpl ./pkg/models

# Step 3: Verify three file types created
ls pkg/models/*.und_*.go
# user.und_patch.go
# user.und_plain.go
# user.und_validator.go
```

**Result**: All three und variants generated for each marked type

### Workflow 3: Selective Generation

**Scenario**: Mark many types but generate only some right now

```bash
# Step 1: Mark various types
codegen automark ./... \
  -g cloner \
  -g und:patch

# Step 2: Generate only cloner (skip und)
codegen autoimpl ./... --only cloner

# Step 3: Later, generate und too
codegen autoimpl ./... --only und:patch
```

**Result**: Incremental generation based on immediate needs

### Workflow 4: Dry Run (Preview Mode)

**Scenario**: See what would be marked/generated before committing

```bash
# Preview marking
codegen automark ./pkg/models -g cloner --dry-run
# Shows which types would be marked

# After reviewing, do it for real
codegen automark ./pkg/models -g cloner

# Preview generation
codegen autoimpl ./pkg/models --dry-run
# Shows which files would be generated

# After reviewing, generate
codegen autoimpl ./pkg/models
```

**Result**: Full visibility before any file modifications

### Workflow 5: Manual Marker Editing

**Scenario**: automark creates base markers, then manually customize

```bash
# Step 1: Auto-mark with defaults
codegen automark ./pkg/models -g cloner

# Step 2: Manually edit markers for specific types
# In pkg/models/user.go, change:
#   //codegen:cloner
# To:
#   //codegen:cloner no-copy:copy chan:make

# Also add field-level directive:
#   //cloner:copyptr
#   Mutex *sync.Mutex

# Step 3: Generate with mixed configuration
codegen autoimpl ./pkg/models
```

**Result**: Combination of automatic and manual configuration

### Workflow 6: Incremental Development

**Scenario**: Add new types over time, regenerate as needed

```bash
# Initial setup
codegen automark ./pkg/models -g cloner
codegen autoimpl ./pkg/models

# ...time passes, new types added to pkg/models/order.go...

# Re-run marking (idempotent)
codegen automark ./pkg/models -g cloner
# Only marks new types, skips existing

# Regenerate all
codegen autoimpl ./pkg/models
# Updates existing files, adds new ones

# Verify
go test ./pkg/models
```

**Result**: Easy to keep generated code in sync as codebase evolves

## Migration from Old Commands

### Old cloner command

**Before**:
```bash
codegen cloner --dir ./pkg/models --no-copy-copy --chan-make
```

**After**:
```bash
codegen automark ./pkg/models -g cloner --no-copy copy --chan make
codegen autoimpl ./pkg/models
```

### Old undgen commands

**Before**:
```bash
codegen undgen patch --dir ./pkg/models
codegen undgen plain --dir ./pkg/models
codegen undgen validator --dir ./pkg/models
```

**After**:
```bash
# Mark once for all three
codegen automark ./pkg/models -g und:patch -g und:plain -g und:validator

# Generate all at once
codegen autoimpl ./pkg/models
```

**Benefit**: One marking command replaces three generation commands

## Advanced Usage

### Filtering Types

**Mark only specific types**:

```bash
# Only mark types ending in "Request" or "Response"
codegen automark ./pkg/api \
  -g cloner \
  --include '*Request' \
  --include '*Response'

# Only mark exported types
codegen automark ./pkg/internal \
  -g cloner \
  --exported-only
```

### Force Re-marking

**Update existing markers with new configuration**:

```bash
# Initial marking
codegen automark ./pkg/models -g cloner --no-copy ignore

# Later, change behavior
codegen automark ./pkg/models -g cloner --no-copy copy --force
# Overwrites existing markers

# Regenerate with new config
codegen autoimpl ./pkg/models
```

### Multi-Package Operations

**Process entire codebase**:

```bash
# Mark everything
codegen automark ./... -g cloner

# Generate everything
codegen autoimpl ./...

# Or specific patterns
codegen automark ./pkg/... ./cmd/... -g cloner
codegen autoimpl ./pkg/... ./cmd/...
```

### Ignoring Generated Files

**Prevent scanning generated code**:

```bash
codegen autoimpl ./... --ignore-generated
```

**Use case**: If generated files somehow got markers, ignore them

## Troubleshooting

### No types marked

**Problem**: automark reports "0 types marked"

**Solutions**:
```bash
# Check if types exist
ls pkg/models/*.go

# Verbose mode for details
codegen automark ./pkg/models -g cloner --verbose

# Try dry-run to see eligibility
codegen automark ./pkg/models -g cloner --dry-run
```

### No code generated

**Problem**: autoimpl reports "No marked types found"

**Solutions**:
```bash
# Check for markers
grep -r "//codegen:" pkg/models/

# If none, run automark first
codegen automark ./pkg/models -g cloner

# Then generate
codegen autoimpl ./pkg/models
```

### Compilation errors in generated code

**Problem**: Generated `.clone.go` files don't compile

**Solutions**:
```bash
# Check error messages
go build ./pkg/models

# Verbose generation for details
codegen autoimpl ./pkg/models --verbose

# Verify marker configuration is correct
grep -A5 "//codegen:" pkg/models/*.go
```

### Marker syntax errors

**Problem**: autoimpl reports "Malformed directive"

**Solution**: Fix marker format:
```go
// WRONG
//codegen:cloner no-copy=copy    // Uses = instead of :

// CORRECT
//codegen:cloner no-copy:copy
```

## Integration with Build Process

### go:generate

Add to your source files:

```go
package models

//go:generate codegen automark . -g cloner
//go:generate codegen autoimpl .

type User struct {
    Name string
}
```

Then run:
```bash
go generate ./pkg/models
```

### Makefile

```makefile
.PHONY: generate
generate:
	codegen automark ./... -g cloner
	codegen autoimpl ./...
	go fmt ./...
	go build ./...

.PHONY: generate-check
generate-check:
	codegen autoimpl ./... --dry-run
	@echo "All generated code is up to date"
```

### CI/CD

```yaml
# .github/workflows/ci.yml
steps:
  - name: Install codegen
    run: go install github.com/ngicks/go-codegen/codegen@latest

  - name: Check generated code
    run: |
      codegen autoimpl ./... --dry-run
      git diff --exit-code  # Fail if any generated files changed
```

## Best Practices

### 1. Commit markers with source

**Do**: Commit the `//codegen:` comments alongside your types
```bash
git add pkg/models/user.go  # Contains markers
git commit -m "Add User type with cloner marker"
```

**Why**: Markers document generation intent and survive across branches

### 2. Regenerate after type changes

**Do**: After modifying a marked type, regenerate:
```bash
# Edit pkg/models/user.go (add field)
codegen autoimpl ./pkg/models
go test ./pkg/models
```

**Why**: Keeps generated code in sync with source

### 3. Use dry-run for exploration

**Do**: Preview before committing to changes:
```bash
codegen automark ./new-package -g cloner --dry-run
```

**Why**: Understand what will happen before files are modified

### 4. Review markers after automark

**Do**: Check marker placement and configuration:
```bash
codegen automark ./pkg -g cloner
git diff pkg/  # Review changes
```

**Why**: Verify automatic marking matches expectations

### 5. Leverage both auto and manual

**Do**: Let automark create baseline, then manually tune:
```bash
codegen automark ./pkg -g cloner  # Baseline
# Edit markers for special cases
codegen autoimpl ./pkg
```

**Why**: Best of both worlds - automation + control

## Next Steps

- **Read [automark-interface.md](contracts/automark-interface.md)** for complete flag reference
- **Read [autoimpl-interface.md](contracts/autoimpl-interface.md)** for generation options
- **Read [data-model.md](data-model.md)** for understanding marker structure
- **See [research.md](research.md)** for technical design decisions

## Summary

The two-phase workflow (mark → generate) provides:
- ✅ Clear separation of concerns
- ✅ Reviewable intent (markers in source)
- ✅ Flexible configuration (auto + manual)
- ✅ Reduced typing (mark once, generate multiple times)
- ✅ Backward compatible (existing directives work)

**Key Commands**:
```bash
codegen automark [packages] -g [generator]  # Mark types
codegen autoimpl [packages]                  # Generate code
```

That's all you need to get started!
