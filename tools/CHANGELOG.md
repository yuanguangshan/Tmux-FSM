# Changelog

## [v2.1.0] - 2026-02-03

### 🚀 Major Changes
- **Rebranding**: Renamed tool from `gen-docs` to **`codoc`**.
  - Binary name: `codoc`
  - Short alias: `gd`
  - Source file: `codoc.go`
- **JSON Output**: Added `--json` flag to export project structure and statistics in machine-readable format.
- **Build System**: Added support for version injection via `-ldflags` (includes date and git hash).

### 🐛 Bug Fixes
- **Markdown Anchors**: Fixed Table of Contents links to be 100% compatible with GitHub's anchor generation rules (lowercase, sanitized).
- **Binary Detection**: 
  - Fixed logic inversion where read errors were treated as non-binary.
  - Improved minified file detection to be specific to `.min.js/.css/.html` extensions, avoiding false positives (e.g., `admin.go`).
- **File Logic**:
  - Unified file inclusion/exclusion logic between Statistics mode and Generation mode to ensure consistency.
  - Fixed directory ignore rules to correctly match substrings (e.g., ignoring `build` now correctly matches `build-output`).
- **CLI**:
  - Fixed flag overwriting issue between `-ns` and `--no-subdirs`.
  - Fixed default filename generation to correctly use `.json` extension when `--json` is active.
- **Cleanup**: Removed unused legacy code (`shouldIgnoreFile`) and duplicate logic in flag parsing.

### 🛠 Improvements
- **Standardization**: Changed Markdown code fences from 4 backticks to standard 3 backticks.
- **Installation**: Updated `install-codoc.sh` to:
  - Automatically clean up legacy `gen-docs` binaries.
  - Detect and warn about conflicting shell aliases.
  - Inject build version details.

---

## [v2.0.0] - Previous Version
- Initial "Mature Tool" release with streaming write, optimized stats, and multi-stage filtering.
