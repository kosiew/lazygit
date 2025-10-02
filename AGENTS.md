# Agent Guidelines for Lazygit

Welcome! This document contains conventions to follow when editing files in this repository.

## General Practices
- Keep changes focused and well-documented; prefer smaller commits that are easy to review.
- Reference existing patterns in the codebase before introducing new abstractions.
- When adding or modifying Go code, run `go fmt` (or ensure your editor formats with `gofmt`) on the affected files.
- Run `go test ./...` after making Go code changes unless the change is documentation-only.

## Commenting Guidelines

When adding or modifying code, prefer clear, purposeful comments. Use the following three types of comments to make intent, assumptions, and public interfaces explicit:

- Implementation Comments
	- Purpose: Explain non-obvious choices and tricky implementations.
	- When to use: Inside function bodies, complex algorithms, workaround code, or when the reason for a particular implementation is not obvious from the code alone.
	- Style: Keep them short and focused. Prefer explaining the "why" not the "what".
	- Example: "// Use map here to avoid O(n^2) behavior when merging large trees"

- Documentation Comments
	- Purpose: Describe functions, types, packages, and modules; serve as public interface documentation.
	- When to use: On exported functions, types, or package-level behavior that other packages or contributors will rely on.
	- Style: Use Go doc conventions for Go code (comment starts with the name of the thing being documented). Keep descriptions concise and include parameter/return expectations when helpful.
	- Example: "// NewRepo creates a Repo for the path and returns an error if the path is not a git repository."

- Contextual Comments
	- Purpose: Document assumptions, preconditions, and non-obvious requirements.
	- When to use: Where a function relies on external state, ordering guarantees, or has subtle side effects.
	- Style: State the assumption or precondition explicitly and, if applicable, how to enforce or test it.
	- Example: "// Assumes caller has already locked repo.mu"

Use these comment types consistently to make the codebase easier to read and maintain. Avoid redundant comments that restate obvious code. Prefer small code refactors over long explanatory comments when possible.

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
