# Changelog

All notable user-visible changes to Ech0 are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html), and this file follows the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

For releases prior to v4.6.5, see the [GitHub releases page](https://github.com/lin-snow/Ech0/releases) — earlier release notes are not retroactively imported here.

## [Unreleased]

## [4.7.1] - 2026-05-01

### Added

- **Admin tag creation** on the Tag Manager page. Admins can now create orphan tags ahead of time without having to publish an Echo first. New `POST /api/tag` endpoint (scope `echo:write`, admin-only, idempotent on duplicate name).
- **`TheEchoMeta` component** on the Echo detail page, showing creation/update time (precise to the minute), word count, the full tag list, and a private flag.
- **`TheEchoInteractions` component** that bundles share, like, and comments into one interaction zone below each Echo detail. The comment composer is collapsed by default behind a pill trigger to keep the page calm.
- **Hover "open detail" icon** on timeline Echo cards (next to the date) for one-click navigation to the detail page.

### Changed

- **Editor tag input** is now a multi-select picker over existing tags only (no free-typed `#tag` parsing). Capped at 3 tags per Echo, with a toast warning on overflow.
- **Echo detail page** redesigned: transparent canvas background (no card frame), hero header with avatar + server/username, then `TheEchoMeta`, body, and `TheEchoInteractions`.
- **Timeline cards** no longer render inline tags; tag filtering is handled via the existing sidebar / search.
- **About page footer**: "用心打造" / "Built with care" updated to use the heart glyph (`用 ❤️ 打造` / `Built with ❤️`), synced across `zh-CN`, `en-US`, `ja-JP`, and `de-DE`.

### Fixed

- **Tag picker popover** no longer overflows the right edge of the screen on mobile (< 640px). The panel now anchors to the editor toolbar and spans its full inner width on small viewports.
- **Tag picker popover** no longer renders behind the editor image preview. The action row was given an explicit stacking context so the popover layers above subsequent siblings.

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

[Unreleased]: https://github.com/lin-snow/Ech0/compare/v4.7.1...HEAD
[4.7.1]: https://github.com/lin-snow/Ech0/compare/v4.7.0...v4.7.1
