# Tasks: Unified Command Structure

**Input**: Design documents from `/specs/001-unify-2-subcommands/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included for this feature as it involves a significant refactoring requiring verification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Project root: `codegen/`
- New packages: `codegen/pkg/automark/`, `codegen/pkg/autoimpl/`
- Commands: `codegen/cmd/`
- Tests: `codegen/internal/generationtests/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create package directory structure for automark at codegen/pkg/automark
- [ ] T002 Create package directory structure for autoimpl at codegen/pkg/autoimpl
- [ ] T003 Create integration test directory at codegen/internal/generationtests/automark_autoimpl
- [ ] T004 [P] Create internal test directory at codegen/pkg/automark/internal/tests
- [ ] T005 [P] Create internal test directory at codegen/pkg/autoimpl/internal/tests

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Create TypeMarker data structure in codegen/pkg/automark/marker.go
- [ ] T007 Create GeneratorSpec data structure in codegen/pkg/automark/config.go
- [ ] T008 Create MarkerConfig struct in codegen/pkg/automark/config.go
- [ ] T009 Create DispatchConfig struct in codegen/pkg/autoimpl/config.go
- [ ] T010 Create GeneratorRegistry type and default registry in codegen/pkg/autoimpl/dispatcher.go
- [ ] T011 Implement directive comment format constants in codegen/pkg/automark/format.go
- [ ] T012 Create test input files (unmarked types) in codegen/internal/generationtests/automark_autoimpl/input/types.go
- [ ] T013 Add generator registry entries for cloner in codegen/pkg/autoimpl/dispatcher.go
- [ ] T014 [P] Add generator registry entries for und:patch in codegen/pkg/autoimpl/dispatcher.go
- [ ] T015 [P] Add generator registry entries for und:plain in codegen/pkg/autoimpl/dispatcher.go
- [ ] T016 [P] Add generator registry entries for und:validator in codegen/pkg/autoimpl/dispatcher.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Mark Types for Code Generation (Priority: P1) 🎯 MVP

**Goal**: Implement automark command that discovers types and adds directive comments

**Independent Test**: Run automark on test packages, verify directive comments are correctly added to eligible types

### Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T017 [P] [US1] Create test for basic type marking in codegen/pkg/automark/internal/tests/marker_test.go
- [ ] T018 [P] [US1] Create test for idempotency (re-marking existing) in codegen/pkg/automark/internal/tests/idempotency_test.go
- [ ] T019 [P] [US1] Create test for type filtering in codegen/pkg/automark/internal/tests/filter_test.go
- [ ] T020 [P] [US1] Create test for dry-run mode in codegen/pkg/automark/internal/tests/dryrun_test.go

### Implementation for User Story 1

- [ ] T021 [P] [US1] Implement package loading logic in codegen/pkg/automark/loader.go
- [ ] T022 [P] [US1] Implement type discovery using existing type graph in codegen/pkg/automark/discovery.go
- [ ] T023 [US1] Implement type filter logic (include/exclude patterns) in codegen/pkg/automark/filter.go (depends on T021, T022)
- [ ] T024 [US1] Implement existing marker detection logic in codegen/pkg/automark/detector.go
- [ ] T025 [US1] Implement AST modification for comment insertion using dst in codegen/pkg/automark/writer.go
- [ ] T026 [US1] Implement file write-back with dst.Restorer in codegen/pkg/automark/writer.go
- [ ] T027 [US1] Implement dry-run mode (preview without writing) in codegen/pkg/automark/marker.go
- [ ] T028 [US1] Implement configuration embedding (flags to directive config) in codegen/pkg/automark/config.go
- [ ] T029 [US1] Create automark Cobra command structure in codegen/cmd/automark.go
- [ ] T030 [US1] Add --generator flag handling in codegen/cmd/automark.go
- [ ] T031 [US1] Add --dry-run flag handling in codegen/cmd/automark.go
- [ ] T032 [P] [US1] Add --include/--exclude filter flags in codegen/cmd/automark.go
- [ ] T033 [P] [US1] Add --exported-only flag in codegen/cmd/automark.go
- [ ] T034 [P] [US1] Add generator-specific configuration flags (--no-copy, --chan, etc.) in codegen/cmd/automark.go
- [ ] T035 [US1] Wire up automark command to root command in codegen/cmd/root.go
- [ ] T036 [US1] Implement help text and usage examples for automark in codegen/cmd/automark.go
- [ ] T037 [US1] Add error handling and validation for automark in codegen/pkg/automark/validator.go
- [ ] T038 [US1] Add logging support (verbose mode) in codegen/pkg/automark/marker.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Generate Implementation from Marked Types (Priority: P2)

**Goal**: Implement autoimpl command that scans for marked types and generates code

**Independent Test**: Run autoimpl on packages with marked types, verify generated code compiles and is correct

### Tests for User Story 2

- [ ] T039 [P] [US2] Create test for marker parsing in codegen/pkg/autoimpl/internal/tests/parser_test.go
- [ ] T040 [P] [US2] Create test for generator dispatch in codegen/pkg/autoimpl/internal/tests/dispatch_test.go
- [ ] T041 [P] [US2] Create test for cloner generation via dispatch in codegen/pkg/autoimpl/internal/tests/cloner_dispatch_test.go
- [ ] T042 [P] [US2] Create test for und generation via dispatch in codegen/pkg/autoimpl/internal/tests/und_dispatch_test.go

### Implementation for User Story 2

- [ ] T043 [P] [US2] Implement type-level marker scanning in AST in codegen/pkg/autoimpl/scanner.go
- [ ] T044 [P] [US2] Implement directive comment parser extending pkg/directive in codegen/pkg/autoimpl/parser.go
- [ ] T045 [US2] Implement marker validation and error handling in codegen/pkg/autoimpl/validator.go (depends on T043, T044)
- [ ] T046 [US2] Implement GeneratorSpec extraction from parsed markers in codegen/pkg/autoimpl/parser.go
- [ ] T047 [US2] Implement dispatcher that invokes generators based on markers in codegen/pkg/autoimpl/dispatcher.go
- [ ] T048 [US2] Implement generator invocation wrapper for cloner in codegen/pkg/autoimpl/invoke_cloner.go
- [ ] T049 [P] [US2] Implement generator invocation wrapper for und:patch in codegen/pkg/autoimpl/invoke_und.go
- [ ] T050 [P] [US2] Implement generator invocation wrapper for und:plain in codegen/pkg/autoimpl/invoke_und.go
- [ ] T051 [P] [US2] Implement generator invocation wrapper for und:validator in codegen/pkg/autoimpl/invoke_und.go
- [ ] T052 [US2] Create autoimpl Cobra command structure in codegen/cmd/autoimpl.go
- [ ] T053 [US2] Add --dry-run flag handling in codegen/cmd/autoimpl.go
- [ ] T054 [P] [US2] Add --only/--skip generator filter flags in codegen/cmd/autoimpl.go
- [ ] T055 [P] [US2] Add --ignore-generated flag in codegen/cmd/autoimpl.go
- [ ] T056 [US2] Wire up autoimpl command to root command in codegen/cmd/root.go
- [ ] T057 [US2] Implement help text and usage examples for autoimpl in codegen/cmd/autoimpl.go
- [ ] T058 [US2] Add error handling for malformed markers in codegen/pkg/autoimpl/validator.go
- [ ] T059 [US2] Add logging support (verbose mode) in codegen/pkg/autoimpl/dispatcher.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Execute Complete Mark-Then-Generate Workflow (Priority: P3)

**Goal**: Verify end-to-end workflow and add integration tests

**Independent Test**: Run automark followed by autoimpl on test packages, verify complete pipeline produces correct results

### Tests for User Story 3

- [ ] T060 [P] [US3] Create end-to-end integration test in codegen/internal/generationtests/automark_autoimpl/e2e_test.go
- [ ] T061 [P] [US3] Create test for workflow with configuration changes in codegen/internal/generationtests/automark_autoimpl/reconfig_test.go
- [ ] T062 [P] [US3] Create test for mixed auto+manual markers in codegen/internal/generationtests/automark_autoimpl/mixed_test.go

### Implementation for User Story 3

- [ ] T063 [US3] Add workflow documentation to README.md
- [ ] T064 [US3] Create migration guide from old commands to new in docs/MIGRATION.md
- [ ] T065 [US3] Update root command help text to explain two-phase workflow in codegen/cmd/root.go
- [ ] T066 [P] [US3] Add usage examples to automark help output in codegen/cmd/automark.go
- [ ] T067 [P] [US3] Add usage examples to autoimpl help output in codegen/cmd/autoimpl.go
- [ ] T068 [US3] Create test scenario files with expected markers in codegen/internal/generationtests/automark_autoimpl/marked/
- [ ] T069 [US3] Create test scenario files with expected generated code in codegen/internal/generationtests/automark_autoimpl/generated/
- [ ] T070 [US3] Verify generated code compilation in integration tests in codegen/internal/generationtests/automark_autoimpl/e2e_test.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T071 [P] Add deprecation warnings to old cloner command in codegen/cmd/cloner.go
- [ ] T072 [P] Add deprecation warnings to old undgen commands in codegen/cmd/undgen.go
- [ ] T073 [P] Update quickstart.md with migration examples
- [ ] T074 Run go fmt on all new packages
- [ ] T075 Run go vet on all new packages
- [ ] T076 Verify all tests pass with go test ./...
- [ ] T077 Run go generate ./... and commit any generated test files
- [ ] T078 [P] Update CLAUDE.md with new command information
- [ ] T079 [P] Create examples directory with sample usage in examples/automark-autoimpl/
- [ ] T080 Performance benchmark for automark on 100 types in codegen/pkg/automark/internal/tests/bench_test.go
- [ ] T081 [P] Add input validation and security checks for path traversal in codegen/pkg/automark/validator.go
- [ ] T082 [P] Add input validation and security checks for path traversal in codegen/pkg/autoimpl/validator.go
- [ ] T083 Verify byte-for-byte compatibility with existing generators in codegen/internal/generationtests/automark_autoimpl/compat_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 can proceed independently (automark)
  - US2 can proceed independently (autoimpl) but conceptually builds on US1
  - US3 depends on US1 and US2 being complete (tests full workflow)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Should integrate with US1 but independently testable
- **User Story 3 (P3)**: Depends on US1 and US2 complete - Tests the integration

### Within Each User Story

- Tests (included in this feature) should be written and FAIL before implementation
- Core data structures before algorithms
- Algorithm implementation before command wiring
- Command wiring before help text
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks (T001-T005) can run in parallel
- Foundational generator registry tasks (T013-T016) can run in parallel
- Test creation for US1 (T017-T020) can run in parallel
- Type discovery components (T021-T022) can run in parallel
- Flag additions (T032-T034) can run in parallel
- Test creation for US2 (T039-T042) can run in parallel
- Scanner and parser (T043-T044) can run in parallel
- Und generator wrappers (T049-T051) can run in parallel
- Integration tests for US3 (T060-T062) can run in parallel
- Help documentation (T066-T067) can run in parallel
- Deprecation warnings (T071-T072) can run in parallel
- Security checks (T081-T082) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Create test for basic type marking in codegen/pkg/automark/internal/tests/marker_test.go"
Task: "Create test for idempotency (re-marking existing) in codegen/pkg/automark/internal/tests/idempotency_test.go"
Task: "Create test for type filtering in codegen/pkg/automark/internal/tests/filter_test.go"
Task: "Create test for dry-run mode in codegen/pkg/automark/internal/tests/dryrun_test.go"

# Launch parallelizable implementation tasks together:
Task: "Implement package loading logic in codegen/pkg/automark/loader.go"
Task: "Implement type discovery using existing type graph in codegen/pkg/automark/discovery.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (automark command)
4. **STOP and VALIDATE**: Test automark independently
5. Verify markers are correctly added to test files
6. Demo automark functionality

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Functional automark (MVP!)
3. Add User Story 2 → Test independently → Functional autoimpl
4. Add User Story 3 → Test independently → Complete workflow verified
5. Polish phase → Production ready

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (automark)
   - Developer B: User Story 2 (autoimpl)
   - Developer C: Prepare US3 integration tests
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests are written FIRST and must FAIL before implementation (TDD)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution principle VI: All tests must verify actual generated code compiles
- Generated code must be byte-for-byte identical to existing generators
- Backward compatibility with existing directive comments is mandatory
