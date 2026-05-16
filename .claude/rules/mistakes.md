# Common Mistakes to Avoid

## Before Writing Code

- **Read before editing**: Always read the file and surrounding context before making changes. Never guess at file contents or structure.
- **Verify assumptions**: Check that the function, variable, or pattern you expect actually exists. Don't assume based on naming conventions alone.
- **Explore existing patterns**: Search the codebase for similar implementations before creating something new. Follow established conventions.
- **Halt on uncertainty**: If something is unclear — which module a file belongs to, which API to use, what the user intends — stop and ask instead of guessing.

## While Writing Code

- **No placeholder comments**: Never write `// ... existing code ...`, `// TODO`, or similar stubs. Always write complete, working logic.
- **Follow existing patterns**: Match the style, naming conventions, error handling, and architecture already present in the codebase.
- **Don't over-engineer**: Solve the problem that was asked about. Don't add abstractions, configuration options, or "nice-to-have" features that weren't requested.
- **Handle errors at boundaries**: Validate user input and external data. Don't add defensive checks for internal code that's already type-safe or well-tested.
- **No hardcoded values**: Avoid embedding secrets, absolute paths, environment-specific URLs, or magic numbers. Use configuration, constants, or environment variables.
- **Keep changes minimal**: Touch only the files and lines necessary. Don't reformat, rename, or refactor code adjacent to your change unless asked.
- **Don't break imports**: When moving or renaming, update all references. Search for usages before modifying exports.

## After Writing Code

- **Update env templates when adding config**: When adding new config fields (env vars) to `cmd/serve.go` or `cmd/worker.go`, update the corresponding template in `secrets/env/` (`serve.env.template` or `worker.env.template`) and **prompt the user to set the actual values** in real env files.
- **Verify before declaring done**: Run the relevant test suite and linter. If either fails, fix the issue — do not declare done until all checks pass with exit code 0. Re-read the diff to confirm it addresses the original issue.
- **Don't commit build artifacts**: Never stage `node_modules/`, `dist/`, `.env`, or other generated/secret files.
- **Keep commits focused**: One logical change per commit. Don't bundle unrelated fixes.
- **Surface blockers early**: If you hit an obstacle, say so immediately instead of silently changing approach.
- **Update mistakes.md after fixing non-trivial bugs**: When you discover a new class of mistake during debugging, add it here so it isn't repeated.
