# Helm Release Validation Checklist

After creating a `v*` tag and publishing the Draft Release, `.github/workflows/release_helm.yml` runs on `release.published` (or can be triggered manually via `workflow_dispatch`). Verify the GitHub Pages site and Helm repository are usable:

1. Check GitHub Pages source is set to branch `gh-pages` and folder `/(root)`.
2. Check `https://lin-snow.github.io/Ech0/` shows the release index page and lists versions `v4.4.0+`.
3. Check `gh-pages` branch contains `index.html`, `.nojekyll`, and `index.yaml`.
4. Check release assets include `ech0-<version>.tgz` (chart `version` matches the release tag without `v`).
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
- `release_helm.yml` sets both `version` and `appVersion` in `charts/ech0/Chart.yaml` from the release tag (e.g. `v4.4.5` → chart `4.4.5`), so each app release produces a new `ech0-4.4.5.tgz` for `helm upgrade`.
- Committed values in `Chart.yaml` are defaults for local `helm install ./charts/ech0`; published chart versions follow tags.
- For manual runs (`workflow_dispatch`), you can pass `release_tag`; if omitted, the workflow uses the latest release tag.
- `release_helm.yml` uses `charts_dir: ./charts` because chart-releaser expects the parent folder of chart directories.
- `release_helm.yml` enables `skip_existing` in chart-releaser to avoid failing when re-running for an already-published chart tag (e.g. `ech0-4.4.5`).
- `release_helm.yml` now triggers on `release.published` so release-page generation can read the latest published release metadata.
- `release_helm.yml` also supports manual rerun from the Actions UI via `workflow_dispatch`.
