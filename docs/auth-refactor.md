# OAuth2/OIDC/Passkey 重构落地说明

## 模块职责边界

- `internal/handler/user`：只处理请求参数解析、HTTP 响应与重定向，不再吞掉回调错误。
- `internal/handler/user/oauth_handler.go`：统一 provider 路由（`/oauth/:provider/*`）。
- `internal/handler/user/passkey_handler.go`：独立 Passkey API 入口和 RP 解析。
- `internal/service/user/user.go`：作为认证编排层，负责 state 校验、回调路由、账号绑定与 token 下发。
- `internal/service/user/oauth_adapter.go`：Provider 适配层，统一 GitHub/Google/QQ/Custom 的身份解析逻辑。
- `internal/service/user/token_service.go`：统一 token 签发实现。
- `internal/util/jwt/jwt.go`：统一 JWT/OIDC token 的安全校验入口。
- `internal/model/user/auth_identity.go`：认证领域新模型（外部身份、本地认证、Passkey 凭证）。

## 成熟库引入边界

- `golang.org/x/oauth2`：用于 OAuth2 标准授权码流程（auth URL 构建、code exchange）。
- `github.com/coreos/go-oidc/v3/oidc`：用于 OIDC `id_token` 标准化验证（签名、iss、aud、exp、nonce）。
- `github.com/go-webauthn/webauthn`：继续用于 Passkey 流程（注册、认证、signCount）。
- `ory/fosite`：当前不引入，仅在需要自建 OAuth2/OIDC Provider（`/authorize`、`/token`、`/jwks`）时再评估。

## 当前安全基线

- OAuth 回调与绑定路径全部返回显式错误，不再“空字符串成功”。
- OIDC 增加 nonce 强校验，防回放攻击。
- 登录与绑定回跳 URL 走 `auth.redirect.allowed_return_urls` allowlist 校验。
- Passkey 使用 `auth.webauthn.rp_id` + `auth.webauthn.origins` 作为边界配置。
- CORS 使用 `web.cors.allowed_origins`，不再复用 `Serverurl/ServerURL`。

## 新配置项（ENV）

- `ECH0_AUTH_REDIRECT_ALLOWED_RETURN_URLS`：逗号分隔的 OAuth 回跳白名单。
- `ECH0_AUTH_WEBAUTHN_RP_ID`：Passkey RPID。
- `ECH0_AUTH_WEBAUTHN_ORIGINS`：逗号分隔的 Passkey 可用来源。
- `ECH0_WEB_CORS_ALLOWED_ORIGINS`：逗号分隔的 CORS 白名单。

## 发布顺序与回滚策略

### Phase 1

- 发布前校验 `Setting.Serverurl` 配置正确（协议、域名、端口）。
- 灰度启用，优先观察 OAuth 登录与 Passkey Begin/Finish 成功率。

### Phase 2

- 重点监控错误日志：`resolve oauth identity failed`、`id_token nonce 不匹配`、`INVALID_PARAMS`。
- 若跨域或回跳误拦截，先修配置再回滚；仅在大面积登录失败时回滚到上一个版本。

### 回滚步骤

- 回滚应用版本到上一个稳定镜像。
- 清理本次发布新增的环境差异（尤其是 `Serverurl` 配置）。
- 复核 OAuth/OIDC/Passkey 冒烟后再恢复流量。

## 验收标准

- OAuth2/OIDC 回调失败可观测（有明确错误与日志）。
- OIDC nonce mismatch 会被拒绝且可被定位。
- Passkey 登录 fallback 逻辑无错误变量返回问题。
- `go test ./...` 全量通过，且 `internal/util/jwt` 含 OIDC nonce 正反用例测试。
