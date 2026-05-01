# Release Process

This document describes how to cut a new release of Ech0.

There are **two paths**:

- **Automated (default)** — [release-please](https://github.com/googleapis/release-please) watches `main` and opens a release PR for you. You merge the release PR; tag, GitHub release, and CI builds happen automatically.
- **Manual fallback** — `make bump` plus hand-rolled commit/tag, kept for hot-fixes and ad-hoc bumps off a non-`main` branch.

If you don't have a strong reason otherwise, use the automated path.

## Versioning

Ech0 follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html):

- **MAJOR** (`X.Y.Z` → `(X+1).0.0`) — breaking changes to the API, auth model, storage layout, config keys, or anything else self-hosted operators rely on. Requires migration notes.
- **MINOR** (`X.Y.Z` → `X.(Y+1).0`) — backwards-compatible new features.
- **PATCH** (`X.Y.Z` → `X.Y.(Z+1)`) — backwards-compatible bug fixes.

Pre-release tags (`v5.0.0-rc.1`, `v5.0.0-beta.1`) work via standard suffixes. The `internal/version.Version` const must include the suffix verbatim — the [verify-version CI job](../../.github/workflows/release.yml) checks this.

The version is declared **once**, in [`internal/version/version.go`](../../internal/version/version.go), with a `// x-release-please-version` marker so release-please knows where to bump it. Everywhere else (`/hello`, `/healthz`, About page, `ech0 version`, MCP server identification, connect federation) reads from this const. Build-time `Commit` and `BuildTime` are injected via `-ldflags -X`.

**Never edit the version string in any other file.** If you find a hardcoded version anywhere, that is a bug to fix, not a place to also bump.

## Conventional commits drive everything

Both paths rely on commit messages being [conventional](https://www.conventionalcommits.org/):

- `feat(scope): ...` → MINOR bump, `Added` section in CHANGELOG
- `fix(scope): ...` → PATCH bump, `Fixed` section
- `perf(scope): ...` → PATCH bump, `Performance` section
- `refactor(scope): ...` → PATCH bump, `Changed` section
- `feat!: ...` or `BREAKING CHANGE:` in footer → MAJOR bump
- `chore` / `docs` / `style` / `build` / `ci` / `test` → no bump (hidden from CHANGELOG by default)

Configured in [`.release-please-config.json`](../../.release-please-config.json).

## Path A — Automated (default)

```
1. Merge feature/fix PRs to main using conventional-commit messages.
2. release-please opens (or updates) a "chore(main): release X.Y.Z" PR.
   It contains: bumped version.go, bumped manifest, new CHANGELOG entry.
3. You review the release PR. If the auto-generated CHANGELOG needs editing
   (rewording, grouping, adding migration notes), commit edits to the release
   PR branch directly.
4. Merge the release PR (regular merge, not squash).
5. release-please tags vX.Y.Z and creates a published GitHub release.
6. The tag push triggers .github/workflows/release.yml, which builds
   linux/amd64 + linux/arm64 binaries with ldflags-injected commit/build_time
   and pushes Docker images to GHCR / Docker Hub.
7. release.yml uploads the .tar.gz artifacts to the release-please-created
   GitHub release.
8. Promote / verify (see "After the release" below).
```

The release PR is **always open** as long as there are unreleased changes on `main`. Each new merged commit refreshes it. You can leave it sitting for days while you batch up changes; merging a one-line `fix:` doesn't force you to release immediately.

### Editing the auto-generated CHANGELOG entry

release-please's CHANGELOG entries come from commit subjects. If a subject is unclear or you want to add migration notes, edit `CHANGELOG.md` on the release PR branch. release-please preserves your edits across re-runs.

For breaking changes, document the migration explicitly:

```markdown
## [5.0.0] - 2026-XX-XX

### ⚠ BREAKING CHANGES

- **Storage layer**: `STORAGE_TYPE=local` no longer accepts relative paths
  — set `STORAGE_LOCAL_ROOT=/absolute/path` before upgrading. Existing
  installs with a relative `data/` path will fail at startup.

### Added
- ...
```

### Skipping a release

If you've merged something to main and don't want it released yet (e.g. waiting for a related PR), close the release PR. release-please will reopen it with the next merge.

### Tag signing in this path

Tags created via release-please go through the GitHub API and inherit GitHub's **web-flow signed Verified badge**. This is the same trust model as commits made through the GitHub web UI.

Cryptographic GPG/SSH signing of the tag git object is **not** configured in this path — that would require provisioning a bot account with a signing key in repo secrets, which is overkill for this project. If you want a signed tag for a particular release, switch to the manual fallback path below for that release.

## Path B — Manual fallback

Use this path when:

- You're cutting a hot-fix from a non-`main` branch (e.g. `release/4.6` for backporting).
- release-please's auto-bump pick is wrong (rare — usually the conventional commits got mis-typed).
- You want a GPG-signed tag for a particular release.

### Procedure

```bash
# starting from a clean tree on main, in sync with origin
make check                                              # full pre-release lint pass
make bump NEW_VERSION=4.6.5                             # edits internal/version/version.go
$EDITOR CHANGELOG.md                                    # rename [Unreleased] → [4.6.5] - YYYY-MM-DD; add a fresh empty [Unreleased]
git commit -am 'chore(release): v4.6.5'
git tag -s v4.6.5 -m 'Release v4.6.5'                   # -s to GPG/SSH-sign; -a if no key configured
git push origin main v4.6.5
gh release view v4.6.5
```

`make bump` validates semver, refuses to run on a dirty tree, edits the single Version const, sanity-checks `go build`, and prints next-step commands. It deliberately never auto-commits, never tags, never pushes — those remain explicit human actions.

### Setting up GPG/SSH tag signing locally

```bash
# GPG (recommended if you have a key on a hardware token)
git config --global user.signingkey YOUR_GPG_KEY_ID
git config --global commit.gpgsign true
git config --global tag.gpgsign true

# OR SSH (simpler if you don't already have GPG)
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
git config --global tag.gpgsign true
```

Then upload the public key to GitHub → Settings → SSH and GPG keys → "New signing key". Tags pushed afterward will show a Verified badge.

## After the release

Whether automated or manual, once the tag is pushed:

1. **Watch the build.**
   ```bash
   gh run watch
   ```
   `verify-version` should pass (it does when `Version` and the tag agree).

2. **Verify the published artifacts.**
   ```bash
   gh release view vX.Y.Z
   ```
   Check that:
   - Both `ech0-linux-amd64.tar.gz` and `ech0-linux-arm64.tar.gz` are attached.
   - The release notes match the CHANGELOG entry.
   - The release is marked Latest if it's a stable version.

3. **Sanity-check a binary.** Download one, run `./ech0 version` and `./ech0 info`, confirm version + commit hash match the tag.

4. **Verify the Docker image.**
   ```bash
   docker pull ghcr.io/lin-snow/ech0:vX.Y.Z
   docker run --rm ghcr.io/lin-snow/ech0:vX.Y.Z version
   ```

5. **Announce** (Discussions / Discord / blog) for non-trivial releases.

## Troubleshooting

### release-please doesn't open a release PR

- Check that the commits since the last tag are conventional. `chore:` and `docs:` are hidden by default; `wip:` isn't recognized at all. Run `git log $(git describe --tags --abbrev=0)..HEAD --oneline` to see what landed since.
- Check that the `release-please` workflow ran for the latest push. Actions tab → release-please.
- If a release PR is already open and stale, delete the branch (`release-please--branches--main--components--ech0` or similar) and the next push to main will recreate it.

### `verify-version` job fails on tag push

The tag (e.g. `v4.6.5`) and `internal/version.Version` (e.g. `4.6.4`) disagree. This shouldn't happen via release-please (it bumps both atomically). If it does:
- Manual flow forgot `make bump` before tagging — delete the bad tag (`git push origin :refs/tags/v4.6.5`), bump correctly, re-tag.
- Race condition where someone pushed a hotfix to main between release-please's PR creation and merge — re-run release-please to refresh the PR.

### Two GitHub releases for the same tag

If you see both a published release (from release-please) and a draft release (from `release.yml`'s old behavior), confirm `draft: false` in `release.yml`'s "Upload artifacts" step. The `softprops/action-gh-release` action applies `draft` state on update, so `draft: true` would actively demote the release.

## Future improvements

These are not implemented today; listed here so the roadmap is visible.

- **[cosign keyless signing](https://docs.sigstore.dev/cosign/)** for the `.tar.gz` artifacts and Docker image digests. OIDC-based, no secrets to manage. Would let downstream operators run `cosign verify-blob --certificate-identity-regexp '...lin-snow/Ech0...' --certificate-oidc-issuer https://token.actions.githubusercontent.com ech0-linux-amd64.tar.gz`.
- **[SBOM generation](https://github.com/anchore/syft)** alongside release artifacts — Software Bill of Materials for supply-chain audits.
- **Reproducible builds** via `SOURCE_DATE_EPOCH=$(git log -1 --format=%ct)` instead of `date -u +%Y-%m-%dT%H:%M:%SZ` for `BuildTime`. Two builds of the same commit would then produce byte-identical binaries.
- **GPG-signed tags for the automated path** by provisioning a bot account with a signing key. Requires standing up the bot identity and rotating the key periodically; not justified for a single-maintainer project today.
