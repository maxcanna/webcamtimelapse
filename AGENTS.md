# Project Guidelines & Agents Context

This document outlines the architectural rules, coding conventions, and workflow practices for contributing to this project. All agents and developers should follow these guidelines.

## General Guidelines & Workflow

- **Agent Documentation:** Always keep this `AGENTS.md` file updated with any new context, workflows, or instructions so that future agents don't have to be explicitly asked to do it.
- **Version Bumping:** The application's version is tracked in a plain text file named `VERSION` located at `internal/constants/VERSION`. Always keep this file updated so you don't have to be explicitly asked to do it. The version string is embedded into Go via `internal/constants/version.go` and exposed as `constants.Version`.
- **Workflow & Validation:**
  - Run tests after each change and before committing. Review GitHub Actions workflows to understand the testing strategy and mimic the commands locally to anticipate possible failures.
  - Build the application locally to validate your changes:
    - CLI: `go build -o webcamtimelapse-cli ./cmd/cli`
    - GUI: `go build -o webcamtimelapse-gui ./cmd/gui`
  - Building the Fyne-based GUI binary on Linux/Ubuntu environments requires system dependencies. If the build fails, install them via: `sudo apt-get install -y libgl1-mesa-dev xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev`.
- **Commits:**
  - Commit each discrete change separately rather than bundling them into a single commit.
  - Do NOT use Conventional Commits or semantic commit formats. Provide clear, descriptive, non-semantic commit messages.
  - **Strict No-Emoji Rule:** Absolutely do NOT use emojis anywhere in commit messages, pull request titles, or branch name prefixes. This rule must be strictly adhered to without exception.
  - Use fixup commits when making subsequent edits to existing code.
  - Do not create pull requests (PRs) when submitting code; commit and push changes directly to the branch.
- **Dependencies:**
  - The repository does not use Go vendoring; dependencies are resolved externally during build or test cycles.

## Golang Code & Style

- **Logging:** Use `log/slog` for logging (with timestamps) instead of `fmt.Printf`.
- **Error Handling & Defers:**
  - To satisfy `errcheck` linting, explicitly ignore error returns from deferred cleanup functions (e.g., `os.RemoveAll`) by wrapping the call in a closure: `defer func() { _ = os.RemoveAll(path) }()`.
  - To ensure robust HTTP response handling and satisfy linters, check for a non-nil response before deferring its body closure. Use a closure that logs any `Close()` errors using `slog.Warn` to avoid silent failures and resource leaks.
- **Performance & I/O:**
  - To optimize progress reporting in `io.Reader` wrappers (like `PassThruReader`), throttle channel updates by storing the last reported percentage in a field (`lastPct`, initialized to -1) and only sending updates when the integer percentage changes or the transfer is complete.
  - When extracting archives, verify that entries are regular files (not directories) and implement size limits. Extraction of downloaded binaries is protected against decompression bombs by enforcing a 500MB maximum uncompressed size limit using `io.LimitReader`.
- **Graphics & Fyne Optimizations:**
  - To optimize performance during image labeling, cache immutable components such as source images (`image.Uniform`) and font faces in long-lived context objects, while instantiating transient, non-thread-safe objects like `font.Drawer` on the stack per call.

## Separation of Concerns & Architecture

- **Project Structure:** The project is a Go application (module: `go.massi.dev/webcamtimelapse`) that builds separate CLI and Fyne-based GUI binaries. Ensure separation of concerns by keeping CLI and GUI specific logic in their respective `cmd/` packages, and shared core logic in `internal/` packages.

## Testing Style

- **Test Isolation:** Unit tests for `internal/core` use the `core_test` package (black-box testing). This requires importing the `internal/core` package and prefixing its members with `core.` in test code.
