# automark Command Interface

**Feature**: 001-unify-2-subcommands
**Command**: `codegen automark`
**Purpose**: Discover types and mark them with directive comments for code generation

## Command Signature

```bash
codegen automark [flags] [packages...]
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

### Generator Selection

**--generator, -g** (required, multiple allowed)
- **Type**: String
- **Description**: Generator type to mark types for
- **Allowed Values**:
  - `cloner` - Mark for clone method generation
  - `und:patch` - Mark for und patcher generation
  - `und:plain` - Mark for und plain type generation
  - `und:validator` - Mark for und validator generation
- **Examples**:
  - `-g cloner`
  - `-g und:patch -g und:plain`
- **Validation**: Must be recognized generator type

### Generator Configuration

**--no-copy** (optional)
- **Type**: Choice [ignore, disallow, copy]
- **Applies to**: cloner generator
- **Description**: How to handle no-copy types
- **Default**: Use generator default
- **Example**: `--no-copy copy`

**--chan** (optional)
- **Type**: Choice [ignore, disallow, copy, make]
- **Applies to**: cloner generator
- **Description**: How to handle channel fields
- **Default**: Use generator default
- **Example**: `--chan make`

**--func** (optional)
- **Type**: Choice [ignore, disallow, copy]
- **Applies to**: cloner generator
- **Description**: How to handle function fields
- **Default**: Use generator default

**--interface** (optional)
- **Type**: Choice [ignore, copy]
- **Applies to**: cloner generator
- **Description**: How to handle interface fields
- **Default**: Use generator default

### Type Filtering

**--include** (optional, multiple allowed)
- **Type**: Glob pattern
- **Description**: Only mark types matching these patterns
- **Examples**:
  - `--include '*Request'` - Types ending in Request
  - `--include 'User*'` - Types starting with User
- **Default**: All eligible types

**--exclude** (optional, multiple allowed)
- **Type**: Glob pattern
- **Description**: Skip types matching these patterns
- **Examples**:
  - `--exclude '*Internal'`
  - `--exclude 'test*'`
- **Default**: No exclusions
- **Precedence**: Exclude takes precedence over include

**--exported-only** (optional)
- **Type**: Boolean
- **Description**: Only mark exported (capitalized) types
- **Default**: false (mark all eligible types)

### Operation Mode

**--dry-run** (optional)
- **Type**: Boolean
- **Description**: Preview which types would be marked without modifying files
- **Default**: false
- **Output**: List of types that would be marked with their markers

**--force** (optional)
- **Type**: Boolean
- **Description**: Overwrite existing markers even if already present
- **Default**: false (skip types with existing markers)

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
   - Check generator types recognized
   - Verify package patterns valid
   - Ensure file write permissions

2. **Discovery Phase**:
   - Load packages with full type information
   - Build type graph
   - Apply type filters (include/exclude patterns)
   - Identify eligible types based on matcher rules

3. **Marking Phase**:
   - For each eligible type:
     - Check for existing marker (skip unless --force)
     - Generate marker comment with specified generators
     - Embed configuration options from flags
     - Insert marker via AST manipulation

4. **Write Phase**:
   - Write modified files back to disk (unless --dry-run)
   - Preserve original formatting and comments
   - Maintain file permissions

5. **Report Phase**:
   - Print summary: marked count, skipped count, error count
   - List newly marked types (verbose mode)

### Idempotency

- Running automark multiple times on same types is safe
- Existing markers detected and skipped (unless --force)
- Force flag allows re-marking with different configuration

### Error Handling

**Configuration Errors**:
```
Error: Unknown generator type 'foo'
Valid generators: cloner, und:patch, und:plain, und:validator
```

**File System Errors**:
```
Error: Permission denied writing to file: pkg/models/user.go
```

**Parse Errors**:
```
Error: Invalid package pattern './....'
```

**Validation Errors**:
```
Error: No packages match pattern './nonexistent'
```

## Output

### Success Output (normal mode)

```
Marked 15 types across 3 packages
  pkg/models: 8 types
  pkg/services: 5 types
  pkg/handlers: 2 types

Summary:
  Newly marked: 15
  Already marked (skipped): 3
  Errors: 0
```

### Success Output (dry-run mode)

```
[DRY RUN] Would mark 15 types:

pkg/models/user.go:
  User (line 12) → //codegen:cloner no-copy:copy
  Profile (line 45) → //codegen:cloner no-copy:copy

pkg/models/order.go:
  Order (line 8) → //codegen:cloner no-copy:copy
  OrderItem (line 23) → //codegen:und:patch

...

No files modified (dry run mode)
```

### Error Output

```
Error marking types: 2 errors occurred

1. pkg/models/invalid.go: File has syntax errors
2. pkg/services/locked.go: Permission denied

Successfully marked: 13
Failed: 2
```

### Verbose Output

Includes additional information:
- Package loading progress
- Type discovery details
- Marker generation details
- File write operations

## Examples

### Example 1: Mark all types for cloner

```bash
codegen automark ./... -g cloner
```

**Result**: All eligible types in all packages marked with `//codegen:cloner`

### Example 2: Mark with configuration

```bash
codegen automark ./pkg/models -g cloner --no-copy copy --chan make
```

**Result**: Types in pkg/models marked with `//codegen:cloner no-copy:copy chan:make`

### Example 3: Mark for multiple generators

```bash
codegen automark ./pkg/types \
  -g und:patch \
  -g und:plain \
  -g und:validator
```

**Result**: Types marked with all three und generators:
```go
//codegen:und:patch
//codegen:und:plain
//codegen:und:validator
type MyType struct { ... }
```

### Example 4: Selective marking with filters

```bash
codegen automark ./... \
  -g cloner \
  --include '*Request' \
  --include '*Response' \
  --exported-only
```

**Result**: Only exported types ending in Request or Response are marked

### Example 5: Preview changes (dry run)

```bash
codegen automark ./pkg/models -g cloner --dry-run
```

**Result**: Shows what would be marked without modifying files

### Example 6: Force re-marking

```bash
codegen automark ./pkg/models -g cloner --no-copy copy --force
```

**Result**: Updates existing markers with new configuration

## Exit Codes

- **0**: Success (all eligible types marked)
- **1**: Partial success (some types marked, some errors)
- **2**: Failure (configuration error, no types marked)
- **3**: Failure (no packages found)

## Integration with autoimpl

The markers created by automark are consumed by autoimpl:

```bash
# Step 1: Mark types
codegen automark ./pkg/models -g cloner

# Step 2: Generate implementations
codegen autoimpl ./pkg/models
```

The autoimpl command reads the `//codegen:cloner` markers and generates the appropriate code.

## Backward Compatibility

- Existing field-level directives (e.g., `//cloner:copyptr`) are preserved
- Manual markers work identically to automark-generated markers
- Existing generator behavior unchanged
