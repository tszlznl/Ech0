# Helm Release Validation Checklist

After the first `v*` tag triggers `.github/workflows/release_helm.yml`, verify the Helm repository is usable:

1. Check `gh-pages` branch contains `index.yaml`.
2. Check release assets include `ech0-<chart-version>.tgz`.
3. Verify Helm install from repository:

```bash
helm repo add ech0 https://lin-snow.github.io/Ech0
helm repo update
helm install ech0 ech0/ech0
```

4. Verify Helm upgrade from repository:

```bash
helm upgrade ech0 ech0/ech0
```

Notes:
- Before creating a new release tag, bump `charts/ech0/Chart.yaml` `version`.
- If `version` is unchanged, chart-releaser will not produce a new package version.
