# Contributing

## Commit Messages

Commit messages must follow this format:

```text
verb(optional theme): imperative description
```

The verb must be one of:

```text
docs
enhance
add
delete
refactor
fix
build
```

The theme is optional and identifies the part of the project affected by the commit.

The description must use the **imperative mood** and concisely describe the change.

### Examples

```text
add(queue): support job cancellation
fix(worker): prevent duplicate job execution
refactor(queue): separate scheduling from execution
enhance(worker): reduce processing overhead
delete(queue): remove obsolete retry logic
docs(api): document task creation
```

Without a theme:

```text
add: support graceful shutdown
fix: handle invalid task input
refactor: simplify error handling
```

### Imperative Mood

Use:

```text
add task cancellation
fix invalid input handling
delete obsolete retry logic
```

Not:

```text
added task cancellation
fixed invalid input handling
adding task cancellation
task cancellation was added
```
