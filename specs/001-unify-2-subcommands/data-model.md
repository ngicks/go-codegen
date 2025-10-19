# Data Model: Unified Command Structure

**Feature**: 001-unify-2-subcommands
**Date**: 2025-10-18

## Overview

This document defines the key data structures and entities for the automark and autoimpl commands. These structures coordinate the two-phase workflow while reusing existing generator infrastructure.

## Core Entities

### 1. TypeMarker (automark output / autoimpl input)

**Purpose**: Represents a directive comment that marks a type for code generation

**Structure**:
```
TypeMarker:
  - TypeName: string           # Name of the marked type
  - Generators: []GeneratorSpec # List of generators to apply
  - Location: SourceLocation    # File and position information
  - ExistingDirectives: []FieldDirective  # Existing field-level directives
```

**Relationships**:
- Created by: automark command
- Consumed by: autoimpl command
- Embedded in: Go source file as comment

**Validation Rules**:
- TypeName must be a valid Go identifier
- Generators list cannot be empty
- Generator names must be from known registry
- No duplicate generators for same type

**Example**:
```go
// TypeMarker rendered as comment:
//codegen:cloner no-copy:copy chan:make
type MyStruct struct {
    //cloner:copyptr      // FieldDirective
    Mutex *sync.Mutex
}
```

### 2. GeneratorSpec

**Purpose**: Specifies which generator to run and with what configuration

**Structure**:
```
GeneratorSpec:
  - Type: string              # e.g., "cloner", "und:patch", "und:plain"
  - Config: map[string]string # Configuration options (key=value pairs)
```

**Relationships**:
- Part of: TypeMarker
- Maps to: Registered GeneratorFunc in autoimpl

**Validation Rules**:
- Type must match a registered generator
- Config keys must be valid for that generator type
- Conflicting config keys should error

**Examples**:
```
GeneratorSpec{
  Type: "cloner",
  Config: {
    "no-copy": "copy",
    "chan": "make",
    "interface": "copy"
  }
}

GeneratorSpec{
  Type: "und:patch",
  Config: {}
}
```

### 3. FieldDirective (existing, preserved)

**Purpose**: Per-field customization directive

**Structure**:
```
FieldDirective:
  - Field: string       # Field name
  - Generator: string   # e.g., "cloner", "undgen"
  - Directive: string   # e.g., "copyptr", "make", "ignore"
  - Config: map[string]string  # Additional key=value config
```

**Relationships**:
- Part of: Type definition
- Overrides: GeneratorSpec config for specific field
- Preserved by: automark (doesn't modify existing field directives)

**Validation Rules**:
- Field must exist in type
- Directive must be valid for generator
- Takes precedence over type-level config

**Example**:
```go
type Foo struct {
    //cloner:copyptr      // FieldDirective
    Mutex *sync.Mutex
}
```

### 4. SourceLocation

**Purpose**: Tracks where in source code a marker exists

**Structure**:
```
SourceLocation:
  - FilePath: string    # Absolute path to file
  - Package: string     # Go package path
  - Line: int           # Line number of type declaration
  - Column: int         # Column number
```

**Relationships**:
- Part of: TypeMarker
- Used for: Error reporting, conflict detection

**Validation Rules**:
- FilePath must be within working directory
- Package must be valid Go package path
- Line/Column must be positive

### 5. MarkerConfig (automark input)

**Purpose**: Configuration for automark command execution

**Structure**:
```
MarkerConfig:
  - PackagePatterns: []string    # e.g., ["./...", "./pkg/models"]
  - Generators: []GeneratorSpec  # Which generators to mark types for
  - TypeFilters: TypeFilterConfig # Rules for which types to mark
  - DryRun: bool                 # Preview mode
  - Verbose: bool                # Logging level
  - WorkingDir: string           # Base directory
```

**Relationships**:
- Input to: automark command
- Determines: Which types get TypeMarkers

**Validation Rules**:
- PackagePatterns cannot be empty
- Generators cannot be empty
- WorkingDir must exist
- TypeFilters must be valid matcher rules

**State Transitions**:
```
Initial → ValidateConfig → LoadPackages → DiscoverTypes →
ApplyFilters → GenerateMarkers → WriteToFiles → Complete
```

### 6. TypeFilterConfig

**Purpose**: Rules for determining which types are eligible for marking

**Structure**:
```
TypeFilterConfig:
  - IncludeTypes: []string    # Type name patterns (glob)
  - ExcludeTypes: []string    # Type name patterns (glob)
  - RequireExported: bool     # Only mark exported types
  - MatcherRules: MatcherConfig # Reuse existing matcher logic
```

**Relationships**:
- Part of: MarkerConfig
- Maps to: Existing matcher package functionality

**Validation Rules**:
- Include/Exclude patterns must be valid globs
- MatcherRules must be compatible with generator types
- Defaults follow existing generator defaults (safe-by-default)

### 7. DispatchConfig (autoimpl input)

**Purpose**: Configuration for autoimpl command execution

**Structure**:
```
DispatchConfig:
  - PackagePatterns: []string    # Where to scan for marked types
  - GeneratorRegistry: map[string]GeneratorFunc  # Available generators
  - Verbose: bool
  - DryRun: bool
  - IgnoreGenerated: bool        # Skip generated files
  - WorkingDir: string
```

**Relationships**:
  - Input to: autoimpl command
  - Uses: GeneratorRegistry to dispatch work

**Validation Rules**:
- PackagePatterns cannot be empty
- GeneratorRegistry must include all expected generators
- WorkingDir must exist

**State Transitions**:
```
Initial → ValidateConfig → LoadPackages → ScanForMarkers →
ParseMarkers → DispatchGenerators → Complete
```

### 8. GeneratorRegistry

**Purpose**: Maps generator type strings to executable generator functions

**Structure**:
```
GeneratorRegistry:
  - Entries: map[string]GeneratorFunc

Where GeneratorFunc signature:
  func(marked MarkedType, pkgs []*packages.Package, config map[string]string) error
```

**Relationships**:
- Part of: DispatchConfig
- Maps: GeneratorSpec.Type to actual generator implementation

**Validation Rules**:
- Each entry must be a valid, non-nil function
- Generator names must be unique
- Default registry includes: cloner, und:patch, und:plain, und:validator

**Default Registry**:
```
{
  "cloner": invokeCloner,
  "und:patch": invokeUndPatch,
  "und:plain": invokeUndPlain,
  "und:validator": invokeUndValidator
}
```

## Entity Relationships Diagram

```
MarkerConfig → automark → TypeMarker (in source)
                               ↓
                          (persisted as comments)
                               ↓
DispatchConfig → autoimpl → Scan for TypeMarker →
                             Parse to GeneratorSpec →
                             Lookup in GeneratorRegistry →
                             Invoke Generator → Generated Code
```

## Data Flow

### automark Flow:
```
1. User provides MarkerConfig (flags)
2. Load packages matching PackagePatterns
3. For each package:
   a. Discover types using type graph
   b. Apply TypeFilterConfig
   c. For each eligible type:
      - Check for existing TypeMarker (idempotency)
      - Generate TypeMarker with GeneratorSpecs
      - Insert marker comment via dst
4. Write modified files back
5. Report results (marked count, skipped, errors)
```

### autoimpl Flow:
```
1. User provides DispatchConfig (flags)
2. Load packages matching PackagePatterns
3. For each package:
   a. Scan AST for TypeMarkers (//codegen: comments)
   b. Parse each marker to GeneratorSpec list
   c. For each GeneratorSpec:
      - Lookup generator in GeneratorRegistry
      - Invoke with type info and config
      - Generator writes output files
4. Report results (generated count, skipped, errors)
```

## Validation and Error Handling

### automark Validation:
- **Pre-execution**: Validate MarkerConfig, check file permissions
- **During execution**: Check for conflicting markers, validate generator names
- **Post-execution**: Verify all target files remain valid Go syntax

### autoimpl Validation:
- **Pre-execution**: Validate DispatchConfig, check generator registry complete
- **During execution**: Parse markers robustly, fail fast on malformed directives
- **Post-execution**: Ensure generated files compile

### Error Categories:
1. **Configuration Errors**: Invalid flags, missing generators
2. **File System Errors**: Permission denied, path traversal attempts
3. **Parse Errors**: Malformed directive comments
4. **Generator Errors**: Generation failures from underlying generators
5. **Validation Errors**: Generated code doesn't compile

## Persistence

### TypeMarker Persistence:
- **Format**: Go comments in source files
- **Location**: Immediately above type declaration
- **Durability**: Persisted until file modified or removed
- **Version Control**: Part of source, tracked in git

### Generated Code Persistence:
- **Format**: Go source files with specific suffixes
- **Location**: Same directory as source type
- **Overwrite Policy**: Regenerated on each autoimpl run
- **Version Control**: Typically committed (as per existing practice)

## Backward Compatibility

### Existing FieldDirective Support:
- automark never modifies existing field directives
- autoimpl passes field directives to generators unchanged
- Generators continue to respect field-level overrides

### Generator Interface Compatibility:
- autoimpl invokes generators using their existing public APIs
- No changes to generator internal logic required
- Configuration mapping from TypeMarker to generator config

## Performance Characteristics

### Memory:
- **automark**: O(types) for marker storage before write
- **autoimpl**: O(marked_types) for dispatch queue

### Disk I/O:
- **automark**: Writes modified source files (one per package with marks)
- **autoimpl**: Delegates to existing generators (existing I/O patterns)

### CPU:
- **automark**: Type graph traversal + AST modification (< 5s for 100 types)
- **autoimpl**: Marker parsing + generator dispatch (comparable to current)

## Future Extensions

Potential future additions without breaking changes:

1. **Conditional Markers**: Build-tag-conditional generation
2. **Multi-package Markers**: Reference types across packages
3. **Marker Inheritance**: Parent type markers inherited by embedded types
4. **Marker Versioning**: Track directive format version for migrations

All extensions would be additive and maintain backward compatibility with existing markers.
