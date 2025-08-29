# Commands Reference

## Installation

### Prerequisites

```bash
# Required: Install goimports for code formatting
go install golang.org/x/tools/cmd/goimports@latest
```

## Testing

### Testing Guidelines

When testing code that uses `und`, `option`, or `elastic` types:
- Use `und.Equal[T comparable](l, r Und[T])` for comparing `und.Und` types with comparable type parameters
- Use `option.Equal[T comparable](l, r Option[T])` for comparing `option.Option` types
- Use `option.EqualOptions[T comparable](l, r []Option[T])` for comparing slices of options
- Only use `.EqualFunc()` methods when the type parameter is not comparable or needs custom comparison logic

Example:
```go
// Preferred - use package-level Equal functions
undIntCmp := cmp.Comparer(func(a, b und.Und[int]) bool {
    return und.Equal(a, b)
})

// Only when needed for non-comparable types
undCustomCmp := cmp.Comparer(func(a, b und.Und[CustomType]) bool {
    return a.EqualFunc(b, customCompareFunc)
})
```

### Run Tests

Step into `codegen` dir before executiong test against main module.

```bash
cd codegen
```

Re-generate after each code edit.
This will also generate test targets.

```bash
go generate ./...
```

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=./.coverinfo -timeout 60s ./...
go tool cover -html=./.coverinfo -o .coverinfo.html

# Run tests for specific package
go test ./codegen/generator/cloner/...

# Run specific test by name
go test -run TestCloneSimple ./...

# Run with race detection
go test -race -timeout 60s ./...

```

## Code Quality

```bash
# Format all code
go fmt ./...

# Run static analysis
go vet ./...

# Check for module issues
go mod verify

# Update dependencies
go mod tidy

# Download dependencies
go mod download

# Check for outdated dependencies
go list -u -m all
```

## Using the Code Generators

### Cloner Generator

#### Basic Usage

```bash
# Generate clone methods for current package
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./

# Or if installed globally
codegen cloner --pkg ./
```

#### With Options

```bash
# Generate with channel copying enabled
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./ --chan-copy

# Generate with function copying enabled
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./ --func-copy

# Generate with interface copying enabled
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./ --interface-copy

# Skip no-copy types check
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./ --skip-no-copy

# Combine multiple options
go run github.com/ngicks/go-codegen/codegen cloner \
  --pkg ./ \
  --chan-copy \
  --func-copy \
  --interface-copy
```

#### Multiple Packages

```bash
# Generate for all packages recursively
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./...

# Generate for specific packages
go run github.com/ngicks/go-codegen/codegen cloner \
  --pkg ./pkg/types \
  --pkg ./internal/models
```

#### Dry Run

```bash
# Preview what would be generated without writing files
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./ --dry
```

### Undgen Generators

#### Patch Generator

```bash
# Generate patch types for partial updates
go run github.com/ngicks/go-codegen/codegen undgen patch --pkg ./

# With verbose output
go run github.com/ngicks/go-codegen/codegen undgen patch --pkg ./ -v
```

#### Plain Generator

```bash
# Generate plain types (without und wrappers)
go run github.com/ngicks/go-codegen/codegen undgen plain --pkg ./

# For multiple packages
go run github.com/ngicks/go-codegen/codegen undgen plain --pkg ./...
```

#### Validator Generator

```bash
# Generate validator methods
go run github.com/ngicks/go-codegen/codegen undgen validator --pkg ./

# With custom build flags
go run github.com/ngicks/go-codegen/codegen undgen validator \
  --pkg ./ \
  --build-flags "-tags integration"
```

### Common Generator Flags

| Flag            | Short | Description                       | Default           |
| --------------- | ----- | --------------------------------- | ----------------- |
| `--dir`         | `-d`  | Set working directory             | Current directory |
| `--pkg`         | `-p`  | Target package pattern (required) | -                 |
| `--verbose`     | `-v`  | Enable verbose logging            | false             |
| `--dry`         | -     | Dry run mode (no files written)   | false             |
| `--build-flags` | -     | Pass flags to build system        | -                 |

## Development Workflow

### Standard Development Cycle

```bash
# 1. Make changes to code
# 2. Format the code
go fmt ./...

# 3. Run static checks
go vet ./...

# 4. Run tests
go test ./...

# 5. Update dependencies if needed
go mod tidy

# 6. Generate code if types changed
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./

# 7. Test generated code
go test ./...
```

### Before Committing

```bash
# Full quality check
go fmt ./... && go vet ./... && go test ./...

# Verify no uncommitted generated files
git status

# Review changes
git diff

# Stage and commit
git add .
git commit -m "feat: description of changes"
```

## Troubleshooting

### Common Issues

#### goimports not found

```bash
# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Verify installation
which goimports
```

#### Module download errors

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

#### Generated files not updating

```bash
# Remove old generated files
find . -name "*.clone.go" -delete
find . -name "*.und_*.go" -delete

# Regenerate
go run github.com/ngicks/go-codegen/codegen cloner --pkg ./...
```

