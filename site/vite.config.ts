// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import type { Plugin } from "vite";
import { defineConfig } from "vite";

/**
 * React Router 的 dev 中间件会接在 Vite transform 之后，对所有未结束的请求做「文档路由」匹配。
 * 浏览器或旧 Service Worker 若以页面导航方式请求 /@vite-plugin-pwa/*、/sw.js、/.well-known/*、
 * /src/main.ts 等路径，Vite 无法当作模块处理时请求会落到 RR，触发终端里刷屏的 getInternalRouterError。
 * 必须在 RR 之前短路：不要用 configureServer 的「返回函数」注册中间件，否则会插在栈尾，请求已先被 RR 处理。
 */
function reactRouterDevSpuriousRequestFilter(): Plugin {
  return {
    name: "react-router-dev-spurious-request-filter",
    enforce: "pre",
    apply: "serve",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const pathname = req.url?.split("?")[0] ?? "";
        if (
          pathname.startsWith("/@vite-plugin-pwa/") ||
          pathname.startsWith("/.well-known/") ||
          pathname === "/sw.js" ||
          /^\/src\/.*\.(ts|tsx|js|jsx|mjs|cjs)$/.test(pathname)
        ) {
          res.statusCode = 404;
          res.end();
          return;
        }
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [
    tailwindcss(),
    reactRouterDevSpuriousRequestFilter(),
    reactRouter(),
  ],
  resolve: {
    tsconfigPaths: true,
  },
});
