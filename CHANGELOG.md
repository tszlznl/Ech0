# Changelog

All notable user-visible changes to Ech0 are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html), and this file follows the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

For releases prior to v4.6.5, see the [GitHub releases page](https://github.com/lin-snow/Ech0/releases) — earlier release notes are not retroactively imported here.

## [4.7.0]

### Added

- **About page** (`/about`) reachable from the homepage banner. Displays the running instance's version, commit hash, build time, license, copyright, author, and a source-code link pinned to the exact commit. Implements AGPL-3.0 §13 (network users may obtain the corresponding source).
- **`internal/version` package** as the single source of truth for build / release metadata (Version, License, Author, RepoURL, StartYear, plus ldflags-injected Commit and BuildTime). Replaces the version constant that used to live in `internal/model/common`.
- **`make bump NEW_VERSION=X.Y.Z`** target that prepares a clean version-bump commit (does not auto-commit or tag).
- **CI guardrail**: the release workflow now refuses to build when the pushed git tag (`vX.Y.Z`) and `internal/version.Version` disagree. Prevents publishing artifacts that lie about their own version.
- **SPDX / Copyright headers** on every `.go` / `.ts` / `.vue` source file, plus a maintenance script `scripts/add-spdx-headers.mjs` (write / `--dry-run` / `--check` modes).
- **`docs/dev/release-process.md`** documenting the standard release procedure.

### Changed

- **`/api/hello` response shape**: dropped the legacy `github` field; added `commit`, `build_time`, `license`, `author`, `repo_url`, and `copyright`. The frontend reads version metadata from this endpoint instead of hardcoding it. Pre-PR consumers of the `github` field should switch to `repo_url` (no in-tree consumer existed).
- **`web/package.json`** now declares `license`, `author`, and `homepage` so npm tooling and SPDX scanners pick up project licensing without parsing the repo.

### Security

- Pinned `serialize-javascript` to `^7.0.5` in `hub/pnpm-lock.yaml` via `pnpm.overrides`, clearing two Dependabot alerts:
  - [GHSA-5c6j-r48x-rmvq](https://github.com/advisories/GHSA-5c6j-r48x-rmvq) — RCE via `RegExp.flags` and `Date.prototype.toISOString` (HIGH)
  - [GHSA-qj8w-gfj5-8c6v](https://github.com/advisories/GHSA-qj8w-gfj5-8c6v) — CPU-exhaustion DoS via crafted array-like objects (MEDIUM)

  Practical risk in this repo was negligible (the vulnerable code only runs at PWA build time on developer-controlled input), but the alerts are now resolved at the supply-chain level.

[Unreleased]: https://github.com/lin-snow/Ech0/compare/v4.6.4...HEAD
