# refactor-go

## When to use
Use this skill when the user asks to refactor, clean up, or restructure Go code.

## Principles
- Prefer small, reviewable changes over giant diffs.
- Explain trade-offs before risky edits.
- Run `go test ./...` after changes.
- Avoid adding heavy dependencies for one-line utilities.

## Tools
- read
- write
- edit
- bash

## Example tasks
- "Refactor this handler to use interfaces"
- "Clean up error handling in this package"
- "Split this large function"

## Output conventions
- Provide a summary of changes before running commands.
- Use `go vet ./...` and `go test ./...` to verify.
