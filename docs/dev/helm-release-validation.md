# Helm Release Validation Checklist

After creating a `v*` tag and publishing the Draft Release, `.github/workflows/release_helm.yml` runs on `release.published` (or can be triggered manually via `workflow_dispatch`). Verify the GitHub Pages site and Helm repository are usable:

1. Check GitHub Pages source is set to branch `gh-pages` and folder `/(root)`.
2. Check `https://lin-snow.github.io/Ech0/` shows the release index page and lists versions `v4.4.0+`.
3. Check `gh-pages` branch contains `index.html`, `.nojekyll`, and `index.yaml`.
4. Check release assets include `ech0-<chart-version>.tgz`.
5. Verify Helm install from repository:

```bash
helm repo add ech0 https://lin-snow.github.io/Ech0
helm repo update
helm install ech0 ech0/ech0
```

6. Verify Helm upgrade from repository:

```bash
helm upgrade ech0 ech0/ech0
```

Notes:
- Before creating a new release tag, bump `charts/ech0/Chart.yaml` `version`.
- `release_helm.yml` uses `charts_dir: ./charts` because chart-releaser expects the parent folder of chart directories.
- `release_helm.yml` now triggers on `release.published` so release-page generation can read the latest published release metadata.
- `release_helm.yml` also supports manual rerun from the Actions UI via `workflow_dispatch`.
- If `version` is unchanged, chart-releaser will not produce a new package version.
