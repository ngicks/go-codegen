# Feature Specification: Unified Command Structure

**Feature Branch**: `001-unify-2-subcommands`
**Created**: 2025-10-18
**Status**: Draft
**Input**: User description: "refactor codegen structure / code / etc to unify code generator into 2 subcommands."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Mark Types for Code Generation (Priority: P1)

Developers need to discover and mark types in their codebase that require code generation by running an automated discovery command that adds directive comments.

**Why this priority**: This is the first phase of the workflow and must work correctly before any code can be generated. Without reliable type discovery and marking, the entire generation pipeline fails.

**Independent Test**: Can be fully tested by running automark on packages with various type definitions and verifying directive comments are correctly added to eligible types.

**Acceptance Scenarios**:

1. **Given** package patterns pointing to Go packages, **When** developer runs automark, **Then** all eligible types in matching packages are annotated with directive comments
2. **Given** types already marked with directive comments, **When** developer runs automark again, **Then** existing marks are preserved or updated without duplication
3. **Given** specific type selection criteria, **When** developer runs automark with those filters, **Then** only types matching the criteria receive directive comments
4. **Given** a dry-run flag, **When** developer runs automark in dry-run mode, **Then** the tool reports what would be marked without modifying files

---

### User Story 2 - Generate Implementation from Marked Types (Priority: P2)

Developers need to generate code implementations by scanning for types marked with directive comments and producing the appropriate generated code files.

**Why this priority**: This is the second phase of the workflow that delivers actual value. Once types are marked, generating implementations completes the automation.

**Independent Test**: Can be fully tested by running autoimpl on packages containing marked types and verifying generated code compiles and functions correctly.

**Acceptance Scenarios**:

1. **Given** types marked with clone directive comments, **When** developer runs autoimpl, **Then** clone methods are generated with correct signatures and suffixes
2. **Given** types marked with und-related directive comments, **When** developer runs autoimpl, **Then** appropriate und code (patch/plain/validator) is generated
3. **Given** directive comments with configuration options, **When** developer runs autoimpl, **Then** generated code respects the specified options
4. **Given** package patterns, **When** developer runs autoimpl on those patterns, **Then** all marked types in matching packages are processed

---

### User Story 3 - Execute Complete Mark-Then-Generate Workflow (Priority: P3)

Developers need to run the complete workflow (mark types, then generate implementations) efficiently as part of their development or build process.

**Why this priority**: While individual commands work independently, a streamlined end-to-end workflow improves productivity. This is valuable but not critical since users can run commands separately.

**Independent Test**: Can be fully tested by running automark followed by autoimpl and verifying the complete pipeline produces correct results.

**Acceptance Scenarios**:

1. **Given** a fresh codebase with unmarked types, **When** developer runs automark then autoimpl sequentially, **Then** types are marked and implementations are generated correctly
2. **Given** the same package patterns for both commands, **When** developer runs both commands with those patterns, **Then** marked types match generated implementations without mismatches
3. **Given** changes to existing types, **When** developer re-runs automark then autoimpl, **Then** directive comments and generated code are updated appropriately
4. **Given** help output for the workflow, **When** developer reads documentation, **Then** the two-phase workflow (mark then implement) is clearly explained

---

### Edge Cases

- What happens when automark finds no eligible types in the specified packages? (Should complete successfully with informative message, no files modified)
- What happens when autoimpl scans packages but finds no marked types? (Should complete successfully with informative message, no files generated)
- What happens when a developer manually edits directive comments after automark runs? (autoimpl should respect manual edits and generate accordingly)
- What happens when automark encounters types that are already marked? (Should detect existing marks, avoid duplication, optionally update if format changed)
- What happens when autoimpl encounters malformed directive comments? (Should fail fast with clear error message indicating which file/line has the issue)
- What happens when running automark on an invalid package path? (Should fail with actionable error message about the path issue)
- What happens when autoimpl generates code but the marked type has changed since marking? (Should generate based on current type structure, potentially warn about mismatches)
- What happens when a developer runs autoimpl before running automark? (Should work fine if types are already manually marked; otherwise produces no output with informative message)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tool MUST expose exactly two primary subcommands: automark and autoimpl
- **FR-002**: automark subcommand MUST discover types in packages matching provided patterns
- **FR-003**: automark subcommand MUST add directive comments to eligible types indicating what generation is needed
- **FR-004**: automark subcommand MUST detect and preserve existing directive comments to avoid duplication
- **FR-005**: automark subcommand MUST support a dry-run mode that reports intended changes without modifying files
- **FR-006**: autoimpl subcommand MUST scan packages for types with directive comments in the specific format created by automark
- **FR-007**: autoimpl subcommand MUST generate appropriate implementations based on directive comment content (clone, und variants, etc.)
- **FR-008**: Both subcommands MUST accept package patterns to control which packages are processed
- **FR-009**: Both subcommands MUST validate inputs and fail fast with clear error messages for invalid configurations
- **FR-010**: Tool MUST maintain backward compatibility with existing manually-written directive comment syntax
- **FR-011**: Help output MUST clearly document the two-phase workflow (mark then implement) and usage patterns
- **FR-012**: Generated files MUST continue to use established suffixes (.clone.go, .und_patch.go, .und_plain.go, .und_validator.go)
- **FR-013**: autoimpl MUST support directive comments with configuration options that control generation behavior
- **FR-014**: Both subcommands MUST support both single-package and multi-package operation modes
- **FR-015**: autoimpl MUST handle cases where marked types are manually modified after marking by generating based on current type structure

### Key Entities

- **automark Subcommand**: First phase operation that discovers types and annotates them with directive comments; accepts package patterns and type selection filters
- **autoimpl Subcommand**: Second phase operation that reads directive comments and generates implementations; accepts package patterns and delegates to appropriate generators
- **Directive Comment**: Structured comment annotation added by automark or manually; contains generator type and configuration; consumed by autoimpl
- **Package Pattern**: Specification of which Go packages to process; supports wildcards and multiple patterns in one invocation
- **Type Marker**: Logic for determining which types are eligible for marking based on their structure, interfaces, or other characteristics

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can complete the full workflow (mark + generate) using exactly two subcommand invocations
- **SC-002**: automark correctly identifies and marks 100% of eligible types in tested packages without false positives
- **SC-003**: autoimpl generates code for 100% of marked types that match the expected directive format
- **SC-004**: Existing users can migrate from old command structure to new structure in under 10 minutes by reading help output and running one mark+impl cycle
- **SC-005**: All existing generator features (clone, und variants) remain accessible through the new workflow
- **SC-006**: Help output for each subcommand fits on a standard terminal screen (80x24) without scrolling for overview information
- **SC-007**: 100% of existing manually-written directive comments continue to work with autoimpl without modification
- **SC-008**: Generated code is byte-for-byte identical to previous version when using equivalent directive configurations
- **SC-009**: automark completes in under 5 seconds for packages containing up to 100 types
- **SC-010**: The two-phase workflow reduces repetitive command typing by at least 50% for projects with multiple generator types

## Assumptions

The following assumptions were made to complete this specification where details were not explicitly provided:

1. **Directive Comment Format**: Assuming automark will generate directive comments in a specific, parseable format that autoimpl can reliably detect and parse (exact format to be determined during planning)
2. **Type Eligibility**: Assuming automark will use similar matcher logic to the current generators to determine which types are eligible for marking (e.g., no-copy rules, interface rules)
3. **Manual Override Support**: Assuming developers should be able to manually edit or add directive comments after automark runs, with autoimpl respecting those manual changes
4. **Backward Compatibility**: Assuming existing manually-written directive comments should work with autoimpl without modification, but old command-line flags may not be directly transferable
5. **Idempotency**: Assuming automark should be idempotent (running multiple times produces same result) and autoimpl should be safe to re-run without accumulating duplicate generated code
6. **Package Pattern Syntax**: Assuming package patterns will follow Go's standard package path conventions with support for ... wildcards (e.g., ./... for all packages)
7. **Error Handling Philosophy**: Assuming both commands should prefer completing successfully with warnings over failing fast, except for truly invalid inputs
8. **Workflow Flexibility**: Assuming developers can run autoimpl without running automark if they manually write directive comments, supporting both automated and manual workflows
