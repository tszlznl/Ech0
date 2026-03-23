# Access Token Scope Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Ech0 的 JWT 鉴权升级为强制 `typ` 分型与 scope 授权模型，并一次性使旧 JWT（含旧 access/session）失效。

**Architecture:** 在 `jwt` 层扩展 claims 契约（`typ/scope/aud/jti`），在 `middleware` 层拆分“解析”与“授权”，并在路由层声明 scope 要求。`session` token 继续走现有角色逻辑，`access` token 强制 scope 检查；缺失/非法 `typ` 一律 `401`。`access token` 持久层补充 scope/audience/jti 元数据，支持审计与最小权限签发。

**Tech Stack:** Go, Gin, GORM (SQLite), golang-jwt/jwt/v5, Go testing (`go test`), Swag 注释（`internal/swagger/docs.go`）

---

## Scope Check

本次仅覆盖“鉴权与令牌模型”单一子系统（claims、middleware、access token 管理、路由授权、测试）。不包含第三方 OAuth2 授权服务器实现，不拆分为多份计划。

## File Structure

**Create**
- `internal/middleware/scope.go` - `RequireScopes` 中间件与 scope 判定辅助函数。
- `internal/model/auth/scope.go` - scope 常量、token 类型常量、audience 白名单常量。
- `internal/service/setting/access_token_service_test.go` - access token scope/audience 校验测试。
- `internal/middleware/auth_test.go` - `JWTAuthMiddleware` 的 `typ` 与状态码测试。

**Modify**
- `internal/model/auth/auth.go` - 扩展 `MyClaims`（`Type`、`Scopes`）并与 `RegisteredClaims` 对齐。
- `internal/util/jwt/jwt.go` - 统一签发/解析新 claims 契约，拒绝无效 `typ`。
- `internal/util/jwt/jwt_test.go` - claims 契约测试、无 `typ` 失败测试。
- `internal/service/user/token_service.go` - 登录 token 强制签发 `typ=session`。
- `internal/model/setting/setting.go` - `AccessTokenSetting` 增加 `TokenType/Scopes/Audience/JTI/LastUsedAt`。
- `internal/model/setting/setting_dto.go` - `AccessTokenSettingDto` 增加 `scopes[]/audience`。
- `internal/database/database.go` - AutoMigrate 模型变更后的迁移断言（如需要辅助函数）。
- `internal/repository/setting/setting.go` - 持久化新字段查询与写入。
- `internal/service/setting/access_token_service.go` - 新 token 签发校验、旧 token 一次性失效策略配套逻辑。
- `pkg/viewer/viewer.go` / `pkg/viewer/user.go` / `pkg/viewer/noop.go` - viewer 扩展 token 元信息读取。
- `internal/middleware/auth.go` - 解析 `typ/scope/aud/jti`，无/非法 `typ` 直接 `401`。
- `internal/router/router.go` - 预留 auth 解析与 scope 鉴权组合顺序。
- `internal/router/setting.go` / `internal/router/file.go` / `internal/router/user.go` / `internal/router/echo.go` - 为管理与写接口挂载 `RequireScopes`。
- `internal/model/common/error.go` / `internal/model/common/i18n_keys.go` - 增加 `403` 细分错误码与 message key 映射。
- `internal/i18n/locales/zh-CN.json` / `en-US.json` / `de-DE.json` - 新增 token scope/audience/transport 对应文案。
- `internal/router/router_test.go` - 路由授权矩阵回归。

**Test/Verification Targets**
- `internal/util/jwt/jwt_test.go`
- `internal/middleware/auth_test.go`
- `internal/service/setting/access_token_service_test.go`
- `internal/router/router_test.go`

---

### Task 1: JWT Claims 契约与签发入口

**Files:**
- Create: `internal/model/auth/scope.go`
- Modify: `internal/model/auth/auth.go`
- Modify: `internal/util/jwt/jwt.go`
- Modify: `internal/service/user/token_service.go`
- Test: `internal/util/jwt/jwt_test.go`

- [ ] **Step 1: 写失败测试（claims 必须带合法 `typ`）**

```go
func TestCreateClaims_WithSessionType(t *testing.T) {
    claims := CreateSessionClaims(user)
    require.Equal(t, authModel.TokenTypeSession, claims.Type)
}

func TestParseToken_RejectsTokenWithoutType(t *testing.T) {
    token := mustSignLegacyTokenWithoutType(t)
    _, err := ParseToken(token)
    require.Error(t, err)
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/util/jwt -run "TestParseToken_RejectsTokenWithoutType|TestCreateClaims_WithSessionType" -v`  
Expected: 至少 1 个 FAIL（当前实现不校验 `typ`）。

- [ ] **Step 3: 最小实现 claims 新契约**

```go
type MyClaims struct {
    Userid   string   `json:"user_id"`
    Username string   `json:"username"`
    Type     string   `json:"typ"`
    Scopes   []string `json:"scope,omitempty"`
    jwt.RegisteredClaims
}
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/util/jwt -run "TestParseToken_RejectsTokenWithoutType|TestCreateClaims_WithSessionType" -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/model/auth/scope.go internal/model/auth/auth.go internal/util/jwt/jwt.go internal/util/jwt/jwt_test.go internal/service/user/token_service.go
git commit -m "feat(auth): enforce typed jwt claims for session and access tokens"
```

---

### Task 2: Access Token 数据模型与签发输入校验

**Files:**
- Modify: `internal/model/setting/setting.go`
- Modify: `internal/model/setting/setting_dto.go`
- Modify: `internal/repository/setting/setting.go`
- Modify: `internal/service/setting/access_token_service.go`
- Test: `internal/service/setting/access_token_service_test.go`

- [ ] **Step 1: 写失败测试（非法 scope/audience 应拒绝）**

```go
func TestCreateAccessToken_RejectsUnknownScope(t *testing.T) {
    dto := &model.AccessTokenSettingDto{Name: "bad", Scopes: []string{"admin:root"}, Audience: "public-client"}
    _, err := svc.CreateAccessToken(ctx, dto)
    require.Error(t, err)
}

func TestCreateAccessToken_RejectsUnknownAudience(t *testing.T) {
    dto := &model.AccessTokenSettingDto{Name: "bad", Scopes: []string{"echo:read"}, Audience: "unknown-client"}
    _, err := svc.CreateAccessToken(ctx, dto)
    require.Error(t, err)
}

func TestCreateAccessToken_RejectsAdminScopeForNonAdminUser(t *testing.T) {
    ctx := withViewerUser(nonAdminUserID)
    dto := &model.AccessTokenSettingDto{Name: "bad", Scopes: []string{"admin:settings"}, Audience: "public-client"}
    _, err := svc.CreateAccessToken(ctx, dto)
    require.Error(t, err)
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/service/setting -run "TestCreateAccessToken_RejectsUnknownScope|TestCreateAccessToken_RejectsUnknownAudience" -v`  
Expected: FAIL（当前 DTO/服务层未校验）。

- [ ] **Step 3: 最小实现模型与服务校验**

```go
if !authModel.IsValidScope(scope) { return "", errors.New(commonModel.INVALID_PARAMS_BODY) }
if !authModel.IsValidAudience(dto.Audience) { return "", errors.New(commonModel.INVALID_PARAMS_BODY) }
if authModel.HasAdminScope(dto.Scopes) && !user.IsAdmin { return "", errors.New(commonModel.NO_PERMISSION_DENIED) }
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/service/setting -run "TestCreateAccessToken_" -v`  
Expected: PASS。

测试实现约束：

- 若 `CreateAccessToken` 对 `viewer.MustFromContext` 依赖过强，先抽取 `validateAccessTokenRequest(user, dto)` 纯函数并优先测试该函数。
- context 注入统一使用测试辅助函数（例如 `withViewerUser(...)`），避免在每个用例重复拼装 request context。

- [ ] **Step 5: 提交**

```bash
git add internal/model/setting/setting.go internal/model/setting/setting_dto.go internal/repository/setting/setting.go internal/service/setting/access_token_service.go internal/service/setting/access_token_service_test.go
git commit -m "feat(setting): add scoped access token schema and validation"
```

---

### Task 3: Viewer 扩展与 JWT 解析中间件（401 规则）

**Files:**
- Modify: `pkg/viewer/viewer.go`
- Modify: `pkg/viewer/user.go`
- Modify: `pkg/viewer/noop.go`
- Modify: `internal/middleware/auth.go`
- Test: `internal/middleware/auth_test.go`

- [ ] **Step 1: 写失败测试（无 `typ` 返回 401）**

```go
func TestJWTAuthMiddleware_RejectsTokenWithoutType(t *testing.T) {
    // 请求受保护路径，附加 legacy token
    // 断言状态码 401
}
```

- [ ] **Step 1.1: 写失败测试（admin scope + query token 返回 403）**

```go
func TestJWTAuthMiddleware_RejectsAdminScopeTokenFromQuery(t *testing.T) {
    // /api/settings?token=...
    // token 为 typ=access 且 scope 包含 admin:settings
    // 断言 403 + ErrCodeTokenTransportForbidden
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/middleware -run "TestJWTAuthMiddleware_RejectsTokenWithoutType|TestJWTAuthMiddleware_RejectsAdminScopeTokenFromQuery|TestJWTAuthMiddleware_AllowsSessionType" -v`  
Expected: FAIL。

- [ ] **Step 3: 最小实现 viewer token 元信息与 middleware 判定**

```go
if mc.Type != authModel.TokenTypeSession && mc.Type != authModel.TokenTypeAccess {
    unauthorized(ctx, commonModel.ErrCodeTokenInvalid, commonModel.MsgKeyAuthTokenInvalid)
    return
}
if tokenFromQuery && authModel.HasAdminScope(mc.Scopes) {
    forbidden(ctx, commonModel.ErrCodeTokenTransportForbidden, commonModel.MsgKeyAuthTokenTransportForbidden)
    return
}
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/middleware -run "TestJWTAuthMiddleware_" -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add pkg/viewer/viewer.go pkg/viewer/user.go pkg/viewer/noop.go internal/middleware/auth.go internal/middleware/auth_test.go
git commit -m "feat(auth): reject untyped jwt and attach token metadata to viewer"
```

---

### Task 4: Scope 授权中间件与路由绑定

**Files:**
- Create: `internal/middleware/scope.go`
- Modify: `internal/router/router.go`
- Modify: `internal/router/setting.go`
- Modify: `internal/router/file.go`
- Modify: `internal/router/user.go`
- Modify: `internal/router/echo.go`
- Test: `internal/router/router_test.go`

- [ ] **Step 1: 写失败测试（access token scope 不足返回 403）**

```go
func TestSetupRouter_AccessTokenWithoutRequiredScopeGetsForbidden(t *testing.T) {
    // 对 /api/settings (PUT) 发送仅 echo:read 的 access token
    // 断言 403
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/router -run "TestSetupRouter_AccessTokenWithoutRequiredScopeGetsForbidden|TestSetupRouter_AccessTokenWithScopePasses" -v`  
Expected: FAIL（当前路由未挂 scope 中间件）。

- [ ] **Step 3: 最小实现 `RequireScopes` 并在关键写接口挂载**

```go
func RequireScopes(scopes ...string) gin.HandlerFunc {
    return func(c *gin.Context) { /* session 放行，access 校验 scope */ }
}
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/router -run "TestSetupRouter_AccessToken" -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/middleware/scope.go internal/router/router.go internal/router/setting.go internal/router/file.go internal/router/user.go internal/router/echo.go internal/router/router_test.go
git commit -m "feat(authz): enforce route-level scope checks for access tokens"
```

---

### Task 5: 错误码、i18n 与响应语义统一

**Files:**
- Modify: `internal/model/common/error.go`
- Modify: `internal/model/common/i18n_keys.go`
- Modify: `internal/i18n/locales/zh-CN.json`
- Modify: `internal/i18n/locales/en-US.json`
- Modify: `internal/i18n/locales/de-DE.json`
- Modify: `internal/middleware/auth.go`
- Modify: `internal/middleware/scope.go`

- [ ] **Step 1: 写失败测试（403 场景返回明确错误码）**

```go
func TestRequireScopes_ReturnsScopeForbiddenCode(t *testing.T) {
    // scope 不足时断言 code == ErrCodeScopeForbidden
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/middleware -run "TestRequireScopes_ReturnsScopeForbiddenCode|TestRequireScopes_ReturnsAudienceForbiddenCode" -v`  
Expected: FAIL。

- [ ] **Step 3: 最小实现错误码与本地化 key**

```go
const ErrCodeScopeForbidden = "SCOPE_FORBIDDEN"
const MsgKeyAuthScopeForbidden = "auth.scope_forbidden"
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/middleware -run "TestRequireScopes_" -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/model/common/error.go internal/model/common/i18n_keys.go internal/i18n/locales/zh-CN.json internal/i18n/locales/en-US.json internal/i18n/locales/de-DE.json internal/middleware/auth.go internal/middleware/scope.go
git commit -m "feat(i18n): add explicit authz error codes for scope and audience failures"
```

---

### Task 6: 数据库迁移与字段落库验证

**Files:**
- Modify: `internal/model/setting/setting.go`
- Modify: `internal/database/database.go`
- Modify: `internal/repository/setting/setting.go`
- Test: `internal/router/router_test.go`（复用内存 DB 场景）或新增 `internal/database/database_test.go`

- [ ] **Step 1: 写失败测试（AccessTokenSetting 新字段可迁移并可读写）**

```go
func TestMigrateDB_AccessTokenSettingIncludesScopeFields(t *testing.T) {
    // AutoMigrate 后插入带 scopes/audience/jti 的记录并回读
}
```

- [ ] **Step 2: 运行单测确认失败**

Run: `go test ./internal/database -run TestMigrateDB_AccessTokenSettingIncludesScopeFields -v`  
Expected: FAIL（字段不存在或无法写入）。

- [ ] **Step 3: 最小实现模型标签与仓储映射**

```go
Scopes   string `gorm:"type:text" json:"scopes"`
Audience string `gorm:"size:64" json:"audience"`
JTI      string `gorm:"size:64;uniqueIndex" json:"jti"`
```

- [ ] **Step 4: 运行单测确认通过**

Run: `go test ./internal/database -run TestMigrateDB_AccessTokenSettingIncludesScopeFields -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/model/setting/setting.go internal/database/database.go internal/repository/setting/setting.go internal/database/database_test.go
git commit -m "feat(db): migrate access token table for scope audience and jti metadata"
```

---

### Task 7: 全量回归与文档同步

**Files:**
- Modify: `README.zh.md`
- Modify: `README.md`
- Modify: `internal/swagger/docs.go`
- Modify: `docs/superpowers/specs/2026-03-23-access-token-scope-design.md`（仅当实现偏离需回写）
- Test: `internal/util/jwt/jwt_test.go`
- Test: `internal/middleware/auth_test.go`
- Test: `internal/service/setting/access_token_service_test.go`
- Test: `internal/router/router_test.go`

- [ ] **Step 1: 先运行定向测试套件**

Run: `go test ./internal/util/jwt ./internal/middleware ./internal/service/setting ./internal/router -v`  
Expected: 全部 PASS。

- [ ] **Step 2: 运行更大范围回归**

Run: `go test ./internal/... -count=1`  
Expected: PASS（允许无关 flaky 需单独记录）。

- [ ] **Step 3: 更新用户文档（破坏性变更公告）**

```markdown
- 旧 JWT（含旧登录态与旧 access token）在升级后会失效，需重新登录/重建 token。
```

- [ ] **Step 3.1: 同步 API 文档定义**

Run: `swag init -g internal/server/server.go -o internal/swagger`  
Expected: `internal/swagger/docs.go`、`internal/swagger/swagger.json`、`internal/swagger/swagger.yaml` 更新。

Run: `go test ./internal/handler/setting -run TestNonExistent -count=1`  
Expected: 编译通过（无测试执行），确认 DTO/handler 依赖一致。

- [ ] **Step 4: 再次验证关键测试**

Run: `go test ./internal/middleware ./internal/router -v`  
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add README.md README.zh.md internal/swagger/docs.go internal/swagger/swagger.json internal/swagger/swagger.yaml docs/superpowers/specs/2026-03-23-access-token-scope-design.md
git commit -m "docs(auth): document jwt type enforcement and token invalidation behavior"
```

---

## Implementation Notes

- 强制遵循 @superpowers:test-driven-development：每个任务先写失败测试再实现。
- 强制遵循 DRY/YAGNI：首版只覆盖已定义 scope 词表，不引入动态策略引擎。
- 强制频繁小提交：每个任务 1 次提交，必要时拆到 2 次。
- 若 route 范围过大，优先保证高风险写接口（设置、用户管理、文件写入、内容写入）先全部落地，再补低风险读接口。
- 架构选择说明：首版维持单 `JWTAuthMiddleware` 负责“解析+注入”，新增 `RequireScopes` 负责授权判定，不额外拆 `ParseTokenMiddleware`，以减少改动面并保持与现有路由结构兼容。

## Execution Risk Checklist

- 升级后旧 JWT 全量失效，需在发布说明中明确“会触发重新登录”。
- `session` 与 `access` 分支必须在中间件中可观测（日志包含 `typ`、`jti`）。
- 任意 `admin:*` scope 不允许通过 query token 传递。

