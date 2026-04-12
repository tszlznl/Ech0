# release/

This directory holds **source templates** used to generate the public-facing
[GitHub Pages release site](https://lin-snow.github.io/Ech0/).  
It is **not** where compiled binaries or packaged artifacts are stored.

## Contents

| File | Purpose |
|---|---|
| `index.html.tmpl` | HTML template for the release landing page. Contains a `<!-- RELEASES -->` placeholder that is populated at build time with per-version download links. |
| `README.md` | This file. |

## How It Works

The generation is fully automated inside the
[`release_helm.yml`](../.github/workflows/release_helm.yml) workflow, which
runs on every published GitHub Release (and can also be triggered manually).

The pipeline performs the following steps:

1. **Fetch release metadata** – Uses `gh release list` to retrieve all
   published, non-draft releases from the GitHub API.
2. **Generate release rows** – A Python script filters releases to `v4.4.0+`,
   then produces an `<li>` element for each version containing:
   - Direct download links for `linux/amd64` and `linux/arm64` tarballs.
   - A link to the GitHub Release Notes page.
3. **Render the template** – `awk` replaces the `<!-- RELEASES -->` marker in
   `index.html.tmpl` with the generated rows, producing the final
   `_site/index.html`.
4. **Assemble static assets** – The project logo (`docs/imgs/logo.svg`) and a
   `.nojekyll` marker are copied into the `_site/` staging directory.
5. **Bootstrap Helm repo index** – If the `gh-pages` branch does not yet
   contain an `index.yaml` (Helm repository index), the workflow packages the
   chart from `charts/ech0/` and generates one so that
   `helm repo add ech0 https://lin-snow.github.io/Ech0` works immediately.
6. **Deploy to `gh-pages`** – The rendered page, logo, `.nojekyll`, Helm
   `index.yaml`, and any chart `.tgz` packages are committed and pushed to the
   `gh-pages` branch, which GitHub Pages serves at
   `https://lin-snow.github.io/Ech0/`.

```
release/index.html.tmpl          ─┐
                                   ├─ awk replace ──► _site/index.html ──► gh-pages branch
GitHub Releases API (v4.4.0+)    ─┘                                         │
                                                                            ├─ index.html
docs/imgs/logo.svg ──────────────────────────────────────────────────────── ├─ logo.svg
                                                                            ├─ .nojekyll
charts/ech0/ (Helm chart) ────── helm package + helm repo index ─────────── ├─ index.yaml
                                                                            └─ ech0-*.tgz
```

## Relationship to Other Directories

| Directory / File | Role |
|---|---|
| `charts/ech0/` | Helm chart source. Packaged into `.tgz` and indexed on `gh-pages`. |
| `.github/workflows/release.yml` | Builds binaries, creates the GitHub Release (draft), and pushes Docker images. |
| `.github/workflows/release_helm.yml` | Publishes the Helm chart **and** regenerates the release page from this template. |
| `.github/workflows/release_zigcc.yml` | Experimental Zig-cc cross-compilation builds (artifacts only, not published to releases yet). |

## Editing the Release Page

To change the layout, styles, or static content of the public release page,
edit `index.html.tmpl` directly. The `<!-- RELEASES -->` comment must remain
intact — the CI pipeline uses it as the injection point for dynamic release
entries. After merging your changes, the next published release (or a manual
workflow dispatch) will regenerate and deploy the updated page.
