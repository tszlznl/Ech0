import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import type { Plugin } from "vite";
import { defineConfig } from "vite";

/**
 * React Router 的 dev 中间件会接在 Vite transform 之后，对所有未结束的请求做「文档路由」匹配。
 * 浏览器或旧 Service Worker 若以页面导航方式请求 /@vite-plugin-pwa/*、/src/main.ts 等路径，
 * Vite 无法当作模块处理时请求会落到 RR，触发终端里刷屏的 getInternalRouterError。
 * 在 RR 之前短路这些路径，返回 404 即可。
 */
function reactRouterDevSpuriousRequestFilter(): Plugin {
  return {
    name: "react-router-dev-spurious-request-filter",
    apply: "serve",
    configureServer(server) {
      return () => {
        server.middlewares.use((req, res, next) => {
          const pathname = req.url?.split("?")[0] ?? "";
          if (
            pathname.startsWith("/@vite-plugin-pwa/") ||
            /^\/src\/.*\.(ts|tsx|js|jsx|mjs|cjs)$/.test(pathname)
          ) {
            res.statusCode = 404;
            res.end();
            return;
          }
          next();
        });
      };
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
