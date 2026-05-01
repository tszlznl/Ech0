// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("privacy", "routes/privacy.tsx"),
  route("docs", "routes/docs/layout.tsx", [
    index("routes/docs/index.tsx"),
    route("*", "routes/docs/doc.tsx"),
  ]),
] satisfies RouteConfig;
