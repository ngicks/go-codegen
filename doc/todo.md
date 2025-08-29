# TODO - go-codegen Project Tasks

## Active Tasks

### High Priority
- [ ] Restructure/refactor codebase for better code reuse across generators
- [ ] Improve shared utilities and common patterns between generators
- [ ] Consolidate duplicate logic in generator implementations
- [ ] Integrate github.com/ngicks/go-fsys-helper/vroot for overlay filesystem
  - [ ] Modify SuffixWriter to maintain code changes in memory using virtual filesystem
  - [ ] Pass virtual filesystem as Overlay option to packages for type-checking
  - [ ] Type-check generated code before writing to disk
  - [ ] Only write files after successful type-checking
- [ ] Replace ad-hoc converters with named functions
  - [ ] Define converter functions with signatures like `func(in map[string][][]T) (out map[string][][]T)`
  - [ ] Generate unique names based on generator name and input/output types
  - [ ] Write conversion functions to uniquely determined file names at package level
  - [ ] Ensure converter names are deterministic and collision-free
  - [ ] One converter file per package containing all conversion functions for that package

## Completed Tasks

- [x] Implement basic cloner generator
- [x] Implement undgen generators (patch, plain, validator)
- [x] Set up type graph infrastructure
- [x] Add directive comment support
- [x] Create test framework for generators

## Documentation Tasks

- [x] Create doc/project_overview.md from serena memory
- [x] Create doc/commands.md from serena memory
- [x] Create doc/architecture.md from serena memory
- [x] Create doc/generators.md with detailed generator documentation
- [ ] Add inline code examples to all documentation
- [ ] Create troubleshooting guide for common issues

## Testing Tasks

- [ ] Increase test coverage to >80%
- [ ] Add fuzz testing for parser components
- [ ] Create end-to-end tests for full generation pipeline
- [ ] Add tests for Windows and macOS platforms
- [ ] Test with various Go versions (1.19, 1.20, 1.21, 1.22)

## Maintenance Tasks

- [ ] Update dependencies with `go get -u ./... && go mod tidy`
- [ ] Review and update Go version requirements
- [ ] Clean up deprecated code paths
- [ ] Profile memory usage during generation

## Notes

- Always run `go test ./...` before committing changes
- Use `go fmt ./...` and `go vet ./...` for code quality
- Generated files must include the proper header comment
- Update this TODO file when starting/completing tasks

Last Updated: 2025-08-27

