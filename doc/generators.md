# Generators Documentation

## Overview

This document provides comprehensive documentation for all code generators in the go-codegen project. Each generator serves a specific purpose in eliminating boilerplate code and ensuring consistency.

## Table of Contents

- [Cloner Generator](#cloner-generator)
- [Undgen Patch Generator](#undgen-patch-generator)
- [Undgen Plain Generator](#undgen-plain-generator)
- [Undgen Validator Generator](#undgen-validator-generator)
- [Directive Comments](#directive-comments)
- [Common Patterns](#common-patterns)

---

## Cloner Generator

### Purpose

Generates deep clone methods for Go types, creating complete copies of data structures including all nested fields.

### When to Use

- Creating copies of complex data structures
- Implementing undo/redo functionality
- Avoiding shared mutable state
- Creating thread-safe copies of data

### Generated Methods

#### For Non-Generic Types

```go
// Original type
type User struct {
    Name    string
    Age     int
    Tags    []string
    Metadata map[string]any
}

// Generated method
func (u User) Clone() User {
    return User{
        Name: u.Name,
        Age:  u.Age,
        Tags: append([]string(nil), u.Tags...),
        Metadata: maps.Clone(u.Metadata),
    }
}
```

#### For Generic Types

```go
// Original type
type Container[T any] struct {
    Value T
    Items []T
}

// Generated method
func (c Container[T]) CloneFunc(cloneT func(T) T) Container[T] {
    cloned := Container[T]{
        Value: cloneT(c.Value),
    }
    cloned.Items = make([]T, len(c.Items))
    for i, item := range c.Items {
        cloned.Items[i] = cloneT(item)
    }
    return cloned
}
```

### Configuration Options

| Option         | Flag               | Description                         | Default |
| -------------- | ------------------ | ----------------------------------- | ------- |
| Channel Copy   | `--chan-copy`      | Copy channels (usually not desired) | false   |
| Function Copy  | `--func-copy`      | Copy function values                | false   |
| Interface Copy | `--interface-copy` | Attempt to clone interface values   | false   |
| Skip No-Copy   | `--skip-no-copy`   | Skip types with `Lock()` method     | false   |

### Control behavior with directive comments

Control cloning behavior with comments:

```go
type Config struct {
    //cloner:skip
    Logger *slog.Logger  // Won't be cloned

    //cloner:copyptr
    SharedRef *State     // Copies pointer, not value

    Data []byte          // Normal deep clone
}
```

### Special Type Handling

#### Built-in Types

- **Slices**: Creates new slice with cloned elements
- **Maps**: Creates new map with cloned keys/values
- **Pointers**: Allocates new memory and clones pointee
- **Arrays**: Element-by-element cloning
- **Channels**: Copied by reference (unless `--chan-copy`)
- **Functions**: Copied by reference (unless `--func-copy`)

#### Standard Library Types

```go
// Automatically handled types
time.Time     // Copied by value
*big.Int      // Uses big.Int's own copy semantics
sync.Mutex    // Skipped (no-copy type)
```

### Examples

#### Simple Struct

```go
// Input
type Person struct {
    Name    string
    Age     int
    Friends []*Person
}

// Generated
func (p Person) Clone() Person {
    cloned := p
    if p.Friends != nil {
        cloned.Friends = make([]*Person, len(p.Friends))
        for i, v := range p.Friends {
            if v != nil {
                cloned.Friends[i] = new(Person)
                *cloned.Friends[i] = v.Clone()
            }
        }
    }
    return cloned
}
```

#### With Embedded Types

```go
// Input
type Employee struct {
    Person  // Embedded
    ID      string
    Manager *Employee
}

// Generated - calls Person.Clone() for embedded field
```

---

## Undgen Patch Generator

### Purpose

Generates "patch" types with all fields optional, used for partial updates in REST APIs or database operations.

### When to Use

- HTTP PATCH endpoints
- Partial database updates
- Configuration merging
- Optional field handling

### Generated Types

```go
// Original
type User struct {
    Name  string
    Email string
    Age   int
}

// Generated patch type
type UserPatch struct {
    Name  und.Und[string]
    Email und.Und[string]
    Age   und.Und[int]
}

// Generated methods
func (p UserPatch) Merge(target *User) {
    if p.Name.IsDefined() {
        target.Name = p.Name.Value()
    }
    if p.Email.IsDefined() {
        target.Email = p.Email.Value()
    }
    if p.Age.IsDefined() {
        target.Age = p.Age.Value()
    }
}
```

### Features

- All fields become `und.Und[T]` (undefined-able)
- Merge methods for applying patches
- JSON marshaling/unmarshaling support
- Validation of defined fields

### Usage Example

```go
// API endpoint
func UpdateUser(id string, patch UserPatch) error {
    user := GetUser(id)
    patch.Merge(&user)
    return SaveUser(user)
}

// Client usage
patch := UserPatch{
    Name: und.Defined("New Name"),
    // Email and Age remain undefined/unchanged
}
```

---

## Undgen Plain Generator

### Purpose

Generates "plain" versions of types that use `und` types, converting between nullable and non-nullable representations.

### When to Use

- API response transformation
- Database model conversion
- Removing nullable wrappers for internal use
- Type bridging between layers

### Generated Types and Methods

```go
// Original with und types
type APIResponse struct {
    Data   und.Und[string]
    Count  und.Und[int]
    Status string
}

// Generated plain type
type APIResponsePlain struct {
    Data   string
    Count  int
    Status string
}

// Generated conversion methods
func (r APIResponse) ToPlain() APIResponsePlain {
    return APIResponsePlain{
        Data:   r.Data.Value(),  // Uses zero value if undefined
        Count:  r.Count.Value(),
        Status: r.Status,
    }
}

func APIResponseFromPlain(plain APIResponsePlain) APIResponse {
    return APIResponse{
        Data:   und.Defined(plain.Data),
        Count:  und.Defined(plain.Count),
        Status: plain.Status,
    }
}
```

### Conversion Rules

- `und.Und[T]` → `T` (zero value if undefined)
- Nested structs are recursively converted
- Slices and maps are handled appropriately
- Non-und fields are copied as-is

---

## Undgen Validator Generator

### Purpose

Generates validation methods for types containing `und` types, ensuring required fields are defined.

### When to Use

- Input validation
- API request validation
- Form validation
- Business rule enforcement

### Generated Methods

```go
// Original
type CreateUserRequest struct {
    Name     und.Und[string] `und:"required"`
    Email    und.Und[string] `und:"required"`
    Nickname und.Und[string] // Optional
}

// Generated validator
func (r CreateUserRequest) Validate() error {
    var errs []error

    if !r.Name.IsDefined() {
        errs = append(errs, fmt.Errorf("name is required"))
    }
    if !r.Email.IsDefined() {
        errs = append(errs, fmt.Errorf("email is required"))
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %w", errors.Join(errs...))
    }
    return nil
}
```

### Validation Rules

- Fields tagged with `und:"required"` must be defined
- Custom validation functions can be specified
- Nested struct validation
- Collection validation (slices, maps)

### Struct Tags

```go
type Example struct {
    // Required field
    Field1 und.Und[string] `und:"required"`

    // Optional with validation
    Field2 und.Und[int] `und:"min=0,max=100"`

    // Nested validation
    Nested und.Und[Inner] `und:"dive"`
}
```

---

## Directive Comments

### Overview

Special comments that control code generation behavior.

### Cloner Directives

| Directive          | Effect                                | Example            |
| ------------------ | ------------------------------------- | ------------------ |
| `//cloner:skip`    | Skip cloning this field               | `//cloner:skip`    |
| `//cloner:copyptr` | Copy pointer instead of dereferencing | `//cloner:copyptr` |
| `//cloner:shallow` | Shallow copy for this field           | `//cloner:shallow` |

### Undgen Directives

| Directive           | Effect                        | Example             |
| ------------------- | ----------------------------- | ------------------- |
| `//undgen:ignore`   | Skip this field in generation | `//undgen:ignore`   |
| `//undgen:required` | Mark as required in validator | `//undgen:required` |

### Usage Example

```go
type Config struct {
    //cloner:skip
    //undgen:ignore
    logger *slog.Logger  // Ignored by all generators

    //cloner:copyptr
    sharedCache *Cache   // Cloner copies pointer only

    //undgen:required
    apiKey und.Und[string] // Required in validator
}
```

---

## Common Patterns

### Pattern 1: DTO with Validation

```go
// Define DTO with und types
type CreateProductDTO struct {
    Name  und.Und[string] `und:"required"`
    Price und.Und[float64] `und:"required,min=0"`
    Tags  und.Und[[]string]
}

// Generate validator
//go:generate codegen undgen validator --pkg ./

// Use in handler
func CreateProduct(dto CreateProductDTO) error {
    if err := dto.Validate(); err != nil {
        return err
    }
    // Process valid DTO
}
```

### Pattern 2: Patch Updates

```go
// Original model
type User struct {
    ID        string
    Name      string
    Email     string
    UpdatedAt time.Time
}

// Generate patch type
//go:generate codegen undgen patch --pkg ./

// Use for updates
func UpdateUser(id string, patch UserPatch) error {
    user := GetUser(id)
    patch.Merge(&user)
    user.UpdatedAt = time.Now()
    return SaveUser(user)
}
```

### Pattern 3: Safe Cloning

```go
// Type with complex state
type GameState struct {
    Players []*Player
    Board   [][]Cell
    History []*Move
    //cloner:skip
    mutex   sync.RWMutex
}

// Generate clone method
//go:generate codegen cloner --pkg ./

// Use for snapshots
func (g *GameState) Snapshot() GameState {
    g.mutex.RLock()
    defer g.mutex.RUnlock()
    return g.Clone()
}
```

### Pattern 4: Type Conversion Pipeline

```go
// API → Internal → DB flow
type APIRequest struct {
    Data und.Und[json.RawMessage]
}

// Generate plain type
//go:generate codegen undgen plain --pkg ./

// Conversion pipeline
func ProcessRequest(req APIRequest) error {
    plain := req.ToPlain()         // Remove und wrapper
    internal := toInternal(plain)  // Business logic
    return saveToDb(internal)      // Persist
}
```

## Best Practices

### 1. Generator Selection

- Use **Cloner** for immutable operations and state management
- Use **Patch** for RESTful APIs and partial updates
- Use **Plain** for layer boundaries and type conversion
- Use **Validator** for input validation and business rules

### 2. Performance Considerations

- Cloner: Be aware of deep copying large structures
- Consider `//cloner:shallow` for large, immutable fields
- Use `//cloner:copyptr` for intentionally shared references

### 3. Testing Generated Code

```bash
# Always test after generation
go generate ./...
go test ./...
```

### 4. Version Control

- Commit generated files
- Include generation in CI/CD pipeline
- Document generation commands in README

### 5. Combining Generators

Generators work well together:

```go
// Original type
type Model struct {
    Field und.Und[string]
}

// Generate all
//go:generate codegen undgen patch --pkg ./
//go:generate codegen undgen plain --pkg ./
//go:generate codegen undgen validator --pkg ./
//go:generate codegen cloner --pkg ./

// Now you have:
// - ModelPatch for updates
// - ModelPlain for internal use
// - Model.Validate() for validation
// - Model.Clone() for copying
```

## Troubleshooting

### Common Issues

#### Generated Code Doesn't Compile

- Ensure `goimports` is installed
- Check for circular dependencies
- Verify all imports are available

#### Generator Skips Types

- Check matcher configuration
- Look for no-copy types
- Verify type is exported

#### Generic Types Not Working

- Ensure Go 1.18+ is used
- Check type parameter constraints
- Verify clone functions are provided

### Debug Output

```bash
# Run with verbose output
codegen cloner --pkg ./ -v

# Dry run to preview
codegen cloner --pkg ./ --dry
```

