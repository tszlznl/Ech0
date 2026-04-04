---
title: 统一登录
description: OAuth2、OIDC 与 Passkey 通行密钥
---

在 **系统设置 → SSO** 中可配置「用第三方账号登录」以及「用系统指纹 / 面容 / 安全密钥登录（Passkey）」。二者可同时开启，用户可选用其一。

---

## OAuth2 / OIDC（GitHub、Google、QQ 等）

适用于：希望用 GitHub、Google、QQ 或**兼容 OAuth2/OIDC 的厂商**登录，而不单独为 Ech0 记一套密码。

### 大致步骤

1. 在对应平台创建 OAuth 应用，拿到 **Client ID** 与 **Client Secret**。  
2. 打开 **系统设置 → SSO → OAuth2**，启用开关，选择模板（GitHub / Google / QQ / 自定义）。  
3. 把页面上自动生成的 **回调地址（Callback / Redirect URI）** 原样复制到第三方平台，**不要手改路径**。  
4. 填写 Client ID / Secret；权限（Scope）按最少够用填写（例如 GitHub `read:user`，Google 常填 `openid`）。  
5. 若提供商支持 OIDC，可在界面补全 Issuer、JWKS 等；保存后系统会对返回的 `id_token` 做校验。

### 常见问题

- **Redirect URI mismatch**：本系统显示的回调地址与第三方填写的不一致（多一个斜杠、协议不同都会失败）。  
- **Invalid client**：ID/Secret 复制错误，或应用未启用。  
- 若部署环境对回跳域名有限制，需保证回调 URL 落在允许列表内（见部署说明或环境变量）。

已登录用户可在**个人设置**里把 OAuth 身份**绑定**到现有账号。

---

## Passkey（通行密钥）

适用于：希望用**指纹、面容或硬件安全密钥**登录，无需密码。

### 配置

1. 打开 **系统设置 → SSO → Passkey**。  
2. 填写 **WebAuthn RP ID**（一般为你的**域名**，不含协议与路径）和 **WebAuthn Origins**（完整站点 URL，如 `https://你的域名`）。  
3. 保存后页面应显示 **Passkey 就绪** 之类提示。  
4. 点击注册 Passkey，按浏览器与系统提示完成绑定。

**RP ID 与 Origins 必须和浏览器地址栏里访问的域名一致**，否则会注册失败或无法登录。

同一页可管理已注册设备（重命名、删除等）。需要对接 HTTP API 时，见部署实例上的 **`/swagger/index.html`** 中 Passkey 相关路径。

---

## 同时使用 OAuth 与 Passkey

可同时启用：用户既可用第三方账号登录，也可在已登录状态下绑定 Passkey，下次用通行密钥登录。具体以当前版本登录页与设置页为准。
