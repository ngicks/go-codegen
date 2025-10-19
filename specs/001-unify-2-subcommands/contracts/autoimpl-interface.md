# autoimpl Command Interface

**Feature**: 001-unify-2-subcommands
**Command**: `codegen autoimpl`
**Purpose**: Scan for marked types and generate implementations based on directive comments

## Command Signature

```bash
codegen autoimpl [flags] [packages...]
```

## Arguments

### Positional Arguments

**packages** (variadic, required)
- **Type**: String patterns
- **Description**: Go package patterns to process
- **Examples**:
  - `./...` - All packages recursively from current directory
  - `./pkg/models` - Single package
  - `./pkg/... ./cmd/...` - Multiple patterns
- **Validation**:
  - At least one pattern required
  - Must be relative paths from working directory
  - Invalid patterns produce error with clear message

## Flags

### Operation Mode

**--dry-run** (optional)
- **Type**: Boolean
- **Description**: Preview what would be generated without writing files
- **Default**: false
- **Output**: List of types and generators that would be invoked

**--ignore-generated** (optional)
- **Type**: Boolean
- **Description**: Skip scanning generated files for markers
- **Default**: false (scan all files)
- **Note**: Generated files have `// Code generated ...` header

### Generator Control

**--only** (optional, multiple allowed)
- **Type**: String
- **Description**: Only run specific generators (ignore others in markers)
- **Examples**:
  - `--only cloner` - Only generate clone code
  - `--only und:patch --only und:plain` - Only these two
- **Validation**: Must be recognized generator types
- **Default**: Run all generators found in markers

**--skip** (optional, multiple allowed)
- **Type**: String
- **Description**: Skip specific generators even if marked
- **Examples**:
  - `--skip und:validator`
- **Validation**: Must be recognized generator types
- **Precedence**: Skip takes precedence over only

### Standard Flags

**--dir** (optional)
- **Type**: Directory path
- **Description**: Working directory for operation
- **Default**: Current directory
- **Validation**: Must be relative path, no path traversal

**--verbose, -v** (optional)
- **Type**: Boolean
- **Description**: Enable verbose logging
- **Default**: false

**--help, -h** (optional)
- **Type**: Boolean
- **Description**: Show help message
- **Output**: Full command documentation

## Behavior

### Execution Flow

1. **Validation Phase**:
   - Validate all flags
   - Verify package patterns valid
   - Check generator registry complete

2. **Scanning Phase**:
   - Load packages with full type information
   - Scan AST for type-level markers (`//codegen:...`)
   - Parse marker comments to extract generators and config
   - Filter based on --only/--skip flags
   - Detect malformed markers (error fast)

3. **Dispatch Phase**:
   - Group marked types by generator
   - For each generator:
     - Collect all types marked for that generator
     - Invoke generator with types and config
     - Generator produces output files

4. **Write Phase**:
   - Generators write to disk (unless --dry-run)
   - Use existing suffixwriter for consistent naming
   - Files processed with goimports

5. **Report Phase**:
   - Print summary: generated count by generator, errors
   - List generated files (verbose mode)

### Marker Format Parsing

autoimpl recognizes these marker formats:

**Simple marker**:
```go
//codegen:cloner
type Foo struct { ... }
```

**Marker with configuration**:
```go
//codegen:cloner no-copy:copy chan:make
type Bar struct { ... }
```

**Multiple generators**:
```go
//codegen:cloner
//codegen:und:patch
type Baz struct { ... }
```

**Field-level directives (existing, preserved)**:
```go
//codegen:cloner
type Qux struct {
    //cloner:copyptr
    Mutex *sync.Mutex
}
```

### Generator Invocation

Each generator is invoked with:
- **Marked Types**: List of types with that generator marker
- **Packages**: Full type information from package loading
- **Config**: Parsed from marker comments

Generators use their existing logic:
- cloner generator: Calls `cloner.Generator.Generate`
- und generators: Call appropriate undgen functions

### Error Handling

**Configuration Errors**:
```
Error: Unknown generator in marker: 'foo'
  at pkg/models/user.go:12
Valid generators: cloner, und:patch, und:plain, und:validator
```

**Parse Errors**:
```
Error: Malformed directive comment
  at pkg/models/order.go:8
  //codegen:cloner no-copy=invalid
Expected format: key:value
```

**Generator Errors**:
```
Error: cloner generation failed for type User
  at pkg/models/user.go:12
  Reason: type contains uncopyable field without directive
```

**File System Errors**:
```
Error: Permission denied writing file: pkg/models/user.clone.go
```

## Output

### Success Output (normal mode)

```
Generated code for 15 types across 3 packages

cloner (8 types):
  pkg/models/user.go → user.clone.go
  pkg/models/profile.go → profile.clone.go
  pkg/services/auth.go → auth.clone.go
  ...

und:patch (5 types):
  pkg/models/order.go → order.und_patch.go
  ...

und:plain (2 types):
  pkg/models/item.go → item.und_plain.go
  ...

Summary:
  Files generated: 15
  Errors: 0
```

### Success Output (dry-run mode)

```
[DRY RUN] Would generate code for 15 types:

cloner:
  User (pkg/models/user.go:12) → user.clone.go
  Profile (pkg/models/profile.go:45) → profile.clone.go
  ...

und:patch:
  Order (pkg/models/order.go:8) → order.und_patch.go
  ...

No files modified (dry run mode)
```

### No Markers Found Output

```
No marked types found in specified packages.

Hint: Run 'codegen automark' first to mark types for generation.
```

### Error Output

```
Error generating code: 2 errors occurred

1. pkg/models/user.go:12: Invalid marker format
2. pkg/services/auth.go:45: cloner generation failed

Successfully generated: 13
Failed: 2
```

### Verbose Output

Includes additional information:
- Package loading progress
- Marker scanning details
- Generator invocation details
- File write operations
- Each generated file path

## Examples

### Example 1: Generate all marked types

```bash
codegen autoimpl ./...
```

**Result**: Scans all packages, generates code for all marked types

### Example 2: Generate for specific package

```bash
codegen autoimpl ./pkg/models
```

**Result**: Only processes pkg/models package

### Example 3: Preview generation (dry run)

```bash
codegen autoimpl ./... --dry-run
```

**Result**: Shows what would be generated without modifying files

### Example 4: Generate only specific generator

```bash
codegen autoimpl ./... --only cloner
```

**Result**: Only generates clone code, skips und generators even if marked

### Example 5: Skip specific generator

```bash
codegen autoimpl ./... --skip und:validator
```

**Result**: Generates all except validator code

### Example 6: Ignore generated files

```bash
codegen autoimpl ./... --ignore-generated
```

**Result**: Doesn't scan `*.clone.go` or other generated files for markers

## Exit Codes

- **0**: Success (all marked types generated)
- **1**: Partial success (some generated, some errors)
- **2**: Failure (configuration error, nothing generated)
- **3**: Failure (no marked types found)
- **4**: Failure (parser errors in markers)

## Integration with automark

autoimpl is designed to work with markers created by automark:

```bash
# Complete workflow
codegen automark ./pkg/models -g cloner --no-copy copy
codegen autoimpl ./pkg/models
```

**Manual markers also supported**:
```go
// This works without running automark
//codegen:cloner no-copy:copy
type MyType struct { ... }
```

Then run:
```bash
codegen autoimpl ./pkg/models
```

## Generator Registry

autoimpl includes these built-in generators:

| Generator ID | Description | Output Suffix |
|--------------|-------------|---------------|
| `cloner` | Clone method generation | `.clone.go` |
| `und:patch` | Und patcher generation | `.und_patch.go` |
| `und:plain` | Und plain type generation | `.und_plain.go` |
| `und:validator` | Und validator generation | `.und_validator.go` |

## Backward Compatibility

### Existing Directive Support

autoimpl fully supports existing manually-written directives:

**Field-level directives** (existing usage):
```go
type Foo struct {
    //cloner:copyptr
    Mutex *sync.Mutex
}
```

These continue to work exactly as before, with or without type-level markers.

**Migration from old commands**:
Old command-line usage can be replaced by:
1. Add markers manually or via automark
2. Run autoimpl

### Generator Compatibility

- autoimpl invokes existing generators without modification
- Generator output is byte-for-byte identical to old commands
- All existing generator features and flags work via marker config

## Performance

### Scanning Performance

- Fast AST traversal for marker detection
- Only parses comments for types (not all code)
- Comparable to existing package loading times

### Generation Performance

- Identical to existing generators (delegates to same code)
- Generators may run in parallel (if supported)

## Troubleshooting

### No output produced

**Cause**: No marked types found

**Solution**:
```bash
# Check for markers
grep -r "//codegen:" pkg/

# If none found, run automark first
codegen automark ./pkg -g cloner
```

### Malformed marker error

**Cause**: Invalid directive syntax

**Solution**: Fix marker format per examples above, or re-run automark

### Generator not found error

**Cause**: Unknown generator type in marker

**Solution**: Check spelling, ensure generator is supported

### Permission denied

**Cause**: Cannot write output file

**Solution**: Check file permissions, ensure directory is writable
