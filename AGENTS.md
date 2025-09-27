# Agent Guidelines for Lazygit

Welcome! This document contains conventions to follow when editing files in this repository.

## General Practices
- Keep changes focused and well-documented; prefer smaller commits that are easy to review.
- Reference existing patterns in the codebase before introducing new abstractions.
- When adding or modifying Go code, run `go fmt` (or ensure your editor formats with `gofmt`) on the affected files.
- Run `go test ./...` after making Go code changes unless the change is documentation-only.

## Testing and Tooling
- Prefer using the repository Makefile targets when applicable (e.g. `make test`) to stay aligned with project workflows.
- Update or add tests alongside functional changes. Highlight any intentional gaps in test coverage in the PR description.

## Documentation
- Keep README and other documentation up-to-date when behavior or usage changes.
- Follow Markdown best practices: wrap lines at ~100 characters and use meaningful headings.

## Pull Request Notes
- Summarize user-facing changes succinctly.
- Mention any new dependencies or migration steps required for reviewers.

Thanks for contributing and happy hacking!
