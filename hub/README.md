# Ech0 Hub

Frontend for **[hub.ech0.app](https://hub.ech0.app)**: it loads a static **instance registry JSON**, fetches each Ech0 deployment’s posts API in parallel in the browser, and merges results into a **single unified timeline**.

## Architecture


| Aspect   | Description                                                                                                                                 |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| **CSR**  | Content is rendered after data is fetched in the browser; no server-side templates.                                                           |
| **SPA**  | Vue 3 + `vue-router` single-page app; route changes do not full-page reload.                                                                |
| **PWA**  | `vite-plugin-pwa` provides the Web App Manifest and Service Worker (installable on desktop/home screen, static asset caching per `vite.config.ts`). |


For a detailed task list as the implementation evolves, see the repo root: `docs/superpowers/plans/2026-04-05-ech0-hub-csr-spa-pwa.md`.

## Listing your instance on Hub

Submit via the **“Register on Ech0 Hub”** Issue form in this repository (GitHub sign-in required). The template applies the `hub` label; ensure a label with that name exists (**Settings → Labels**). After submission, Actions parse the issue and **open a PR** that updates `hub/public/hub.json`. It goes live after maintainers merge. Configure **CORS** on your instance to allow `https://hub.ech0.app`.

## Instance registry `public/hub.json`

At runtime Hub requests `/hub.json` from the same origin (in dev, Vite serves it from `public/`). Each instance needs only `id` and `url`:

```json
{
  "instances": [
    { "id": "my-instance", "url": "https://your-ech0-origin.example.com" }
  ]
}
```

- **`id`**: Short identifier for the instance (source labeling and UI).
- **`url`**: API root **without** a trailing slash; aggregated requests use `{url}/api/echo/query` (same as the main project’s `internal/router`). Health checks use **`GET {url}/healthz`** on the same host (see `internal/router/resource.go`; not under `/api`).

## Health checks and version gate

1. For each instance, call **`GET {url}/healthz`**, parse `data.version` when **`code === 1`** (same contract as `Healthz` in `internal/handler/common/common.go`).
2. Only instances whose version is **≥ 4.4.0** (same semantics as `Version` in `internal/model/common/common.go`) participate in post aggregation.
3. If the check fails or the request errors, show the reason on the page and exclude that instance from the timeline.

## Body / images and reuse from `web`

- **Markdown & images**: Vite `resolve.alias` maps `@` to the repo’s `web/src`, reusing `TheMdPreview` (via `MarkdownRenderer`) and `TheImageGallery`; gallery requests pass the instance `baseUrl`, matching the main Hub use case.
- **Global styles & i18n**: The Hub entry imports `web` theme SCSS, `virtual:uno.css`, and `vue-i18n` (currently `zh-CN` strings) to drive those components.
- **Extension**: Aggregated feeds **omit** posts with `extension` (filtered in `src/composables/useHubMergeFeed.ts`); Hub does not show Extension cards or related wrappers.

## Cross-origin (CORS)

When the browser on `hub.ech0.app` calls each instance, **`/api/*` and `/healthz`** (among others) must allow the Hub origin (e.g. `Access-Control-Allow-Origin: https://hub.ech0.app`). If you cannot change the instance, use a **reverse proxy** on the Hub domain (same-origin avoids CORS).

## Development & build

```bash
pnpm install
pnpm dev
pnpm build
pnpm preview
```

- Local dev defaults to `http://localhost:5173`; add that origin to instance CORS for debugging.

## Stack

Vue 3, TypeScript, Vite, `vue-router`, `vite-plugin-pwa`, `vue-i18n`, UnoCSS; shares some UI with `web/` in this repo (see above).

## Aggregation flow (summary)

1. `GET /hub.json` → instance list.
2. `GET {url}/healthz` → candidates with version ≥ 4.4.0.
3. For each candidate, `POST {url}/api/echo/query` with a body matching the main project’s `EchoQueryDto`; success when `code === 1` and `data.items` is the post array (see `internal/model/common/result.go`).
4. Merge results, sort by `created_at` descending; surface failures or partial errors in the UI.

## Repository layout

This directory is the `hub/` package in the monorepo. It shares Git history with the main Ech0 (Go backend) project but builds and deploys independently.
