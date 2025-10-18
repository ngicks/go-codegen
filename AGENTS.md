# Repository Guidelines

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

### Task Management

Always check and update `doc/todo.md` when:

- Starting new work on this project
- Completing tasks
- Discovering new issues or requirements
- Planning future work

### TODO List Formatting

- All entries and subtasks in `doc/todo.md` must use checkbox notation (`- [ ]` for open tasks, `- [x]` for completed tasks) with concise task names only.
- You may group related tasks using indentation with `-` for parent items and `  -` for child items; keep the checkbox prefix at every level so progress stays visible.
- Prefix every task line, including nested subtasks, with a unique, stable numeric identifier (e.g. `0001`) that never changes when items are reordered so each entry can be traced back to the elaborated plan files under `doc/tasks/` without ambiguity.
- Tasks are grouped by state sections (`## Planned`, `## Active`, `## Completed`, etc.); move items between sections when their state changes to keep the list consistent.
- When a task moves into active work, create or update a plan file under `doc/tasks/` named with a zero-padded index (e.g. `doc/tasks/000_task_name.md`) to hold the elaborated task description.
- Each task file must start with a YAML header containing a `status` field (`planned`, `active`, `canceled`, or `completed`).
- Each task file must include `## parent` and `## children` sections to describe its relationships so the hierarchy can be traversed easily.

## Using Serena MCP Tool

This project has been onboarded with the Serena MCP tool. Use `mcp__serena__` commands to access and store project information. The serena memories complement the documentation in the `./doc` directory.

- You must use serena tools where possible.
- You'll have to read/update serena memory.
- You must not use built-in read / write tool
- You must not use bash to search lines, symbols. Just update serena memory and use serena tools.
