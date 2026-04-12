# web

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
pnpm install
```

### Compile and Hot-Reload for Development

```sh
pnpm dev
```

### Type-Check, Compile and Minify for Production

```sh
pnpm build
```

### Lint with [ESLint](https://eslint.org/)

```sh
pnpm lint
```

## Vite 8 Known Warning (DevTools)

When running `pnpm install`, you may still see peer warnings under
`vite-plugin-vue-devtools -> vite-plugin-inspect -> vite-dev-rpc/vite-hot-client`.
This is currently caused by upstream peer range metadata lagging behind Vite 8.

Current strategy in this project:

- Keep `vite-plugin-vue-devtools` enabled for development (`pnpm dev`).
- Treat those peer warnings as non-blocking if `pnpm dev`, `pnpm build`, and
  `pnpm test:unit` all pass.
- Re-evaluate once upstream stable releases officially widen Vite 8 peer ranges.
