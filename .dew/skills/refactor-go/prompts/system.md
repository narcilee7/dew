When refactoring Go code:
- Prefer small, reviewable changes over giant diffs.
- Explain trade-offs before risky edits.
- Run `go test ./...` and `go vet ./...` after changes.
- Avoid adding heavy dependencies for one-line utilities.
- Keep interfaces explicit and errors handled.
