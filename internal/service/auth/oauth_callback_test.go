// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/kvstore"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	authmock "github.com/lin-snow/ech0/internal/test/mocks/authmock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// allowedReturnURL 是单测里统一使用的、必落在白名单内的客户端回跳地址。
const allowedReturnURL = "https://app.example.com/auth"

// fakeAdapter 实现未导出接口 oauthProviderAdapter，注入到 AuthService.resolveAdapter，
// 从而让 HandleOAuthCallback 走完整流程而不触发真实 OAuth token/userinfo HTTP。
type fakeAdapter struct {
	identity *oauthIdentity
	err      error
}

func (f *fakeAdapter) ResolveIdentity(
	_ *settingModel.OAuth2Setting,
	_ string,
	_ *authModel.OAuthState,
) (*oauthIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identity, nil
}

// fullOAuth2Setting 构造一份字段齐备、可通过 getOAuthSetting 校验的 OAuth2 设置。
// AuthRedirectAllowedReturnURLs 显式写死，使重定向校验与 ENV/全局 config 解耦、结果确定。
func fullOAuth2Setting(provider string) settingModel.OAuth2Setting {
	return settingModel.OAuth2Setting{
		Enable:       true,
		Provider:     provider,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURI:  "https://app.example.com/oauth/" + provider + "/callback",
		AuthURL:      "https://idp.example.com/authorize",
		TokenURL:     "https://idp.example.com/token",
		UserInfoURL:  "https://idp.example.com/userinfo",
		Scopes:       []string{"read:user"},
		IsOIDC:       false,
		Issuer:       "https://idp.example.com",
		JWKSURL:      "https://idp.example.com/jwks",

		AuthRedirectAllowedReturnURLs: []string{allowedReturnURL},
	}
}

// seedOAuth2KV 用一份给定的 OAuth2 设置预置内存 KV。
func seedOAuth2KV(t *testing.T, setting settingModel.OAuth2Setting) kvstore.Store {
	t.Helper()
	kv := kvstore.NewMemory()
	raw, err := json.Marshal(setting)
	require.NoError(t, err)
	require.NoError(t, kv.Set(context.Background(), commonModel.OAuth2SettingKey, string(raw)))
	return kv
}

// newSvc 组装一个仅依赖 mock 协作者的 AuthService。
func newSvc(
	t *testing.T,
	kv kvstore.Store,
) (*AuthService, *authmock.MockRepository, *authmock.MockAuthRepo, *txmock.MockTransactor) {
	t.Helper()
	repo := authmock.NewMockRepository(t)
	authRepo := authmock.NewMockAuthRepo(t)
	tx := txmock.NewMockTransactor(t)
	svc := NewAuthService(tx, repo, authRepo, kv)
	return svc, repo, authRepo, tx
}

// runsTxInline 让 transactor.Run 真正执行回调（提交语义），用于 bind 事务路径。
func runsTxInline(tx *txmock.MockTransactor) {
	tx.EXPECT().
		Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).
		Once()
}

// ---------------------------------------------------------------------------
// HandleOAuthCallback：provider / state 前置校验 + adapter 解析 + 委派
// ---------------------------------------------------------------------------

func TestHandleOAuthCallback_ValidationErrors(t *testing.T) {
	helpers.SetJWTSecret(t, "callback-validation-secret")

	// 一份合法的、provider=github 的 state，供「设置侧」错误用例复用。
	validGithubState, _, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin), "", allowedReturnURL, string(commonModel.OAuth2GITHUB),
	)
	require.NoError(t, err)

	// state 内 provider=google，但回调 provider=github → 进入 provider 不一致分支。
	mismatchProviderState, _, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin), "", allowedReturnURL, string(commonModel.OAuth2GOOGLE),
	)
	require.NoError(t, err)

	cases := []struct {
		name     string
		provider string
		setting  settingModel.OAuth2Setting
		state    string
		wantErr  string
	}{
		{
			name:     "provider not matching configured provider",
			provider: string(commonModel.OAuth2GOOGLE),
			setting:  fullOAuth2Setting(string(commonModel.OAuth2GITHUB)),
			state:    validGithubState,
			wantErr:  commonModel.OAUTH2_NOT_CONFIGURED,
		},
		{
			name:     "oauth2 disabled",
			provider: string(commonModel.OAuth2GITHUB),
			setting: func() settingModel.OAuth2Setting {
				s := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
				s.Enable = false
				return s
			}(),
			state:   validGithubState,
			wantErr: commonModel.OAUTH2_NOT_ENABLED,
		},
		{
			name:     "missing required config field",
			provider: string(commonModel.OAuth2GITHUB),
			setting: func() settingModel.OAuth2Setting {
				s := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
				s.ClientSecret = ""
				return s
			}(),
			state:   validGithubState,
			wantErr: commonModel.OAUTH2_NOT_CONFIGURED,
		},
		{
			name:     "state provider mismatch",
			provider: string(commonModel.OAuth2GITHUB),
			setting:  fullOAuth2Setting(string(commonModel.OAuth2GITHUB)),
			state:    mismatchProviderState,
			wantErr:  commonModel.INVALID_PARAMS,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _, _ := newSvc(t, seedOAuth2KV(t, tc.setting))
			// 这些用例都应在触达 adapter 前返回，注入会 panic 的 resolver 以反证未被调用。
			svc.resolveAdapter = func(string) (oauthProviderAdapter, error) {
				t.Fatalf("resolveAdapter must not be called on validation failure")
				return nil, nil
			}

			out, err := svc.HandleOAuthCallback(tc.provider, "code-123", tc.state)
			require.Error(t, err)
			require.EqualError(t, err, tc.wantErr)
			assert.Empty(t, out)
		})
	}
}

func TestHandleOAuthCallback_InvalidState(t *testing.T) {
	helpers.SetJWTSecret(t, "callback-invalid-state-secret")

	svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
	svc.resolveAdapter = func(string) (oauthProviderAdapter, error) {
		t.Fatalf("resolveAdapter must not be called when state is unparseable")
		return nil, nil
	}

	out, err := svc.HandleOAuthCallback(string(commonModel.OAuth2GITHUB), "code-123", "not-a-jwt")
	require.Error(t, err) // ParseOAuthState 失败：非空、非业务常量错误
	assert.Empty(t, out)
}

func TestHandleOAuthCallback_AdapterErrors(t *testing.T) {
	helpers.SetJWTSecret(t, "callback-adapter-secret")

	state, _, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin), "", allowedReturnURL, string(commonModel.OAuth2GITHUB),
	)
	require.NoError(t, err)

	t.Run("resolveAdapter returns error", func(t *testing.T) {
		svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
		sentinel := errors.New("adapter unavailable")
		svc.resolveAdapter = func(string) (oauthProviderAdapter, error) { return nil, sentinel }

		out, err := svc.HandleOAuthCallback(string(commonModel.OAuth2GITHUB), "code-123", state)
		require.ErrorIs(t, err, sentinel)
		assert.Empty(t, out)
	})

	t.Run("ResolveIdentity returns error", func(t *testing.T) {
		svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
		sentinel := errors.New("token exchange failed")
		svc.resolveAdapter = func(string) (oauthProviderAdapter, error) {
			return &fakeAdapter{err: sentinel}, nil
		}

		out, err := svc.HandleOAuthCallback(string(commonModel.OAuth2GITHUB), "code-123", state)
		require.ErrorIs(t, err, sentinel)
		assert.Empty(t, out)
	})
}

// HandleOAuthCallback 成功路径（OAuth login）：贯通 setting/state 校验 → adapter 解析
// → resolveOAuthCallback 登录签发 → 重定向追加一次性 code。
func TestHandleOAuthCallback_LoginSuccess(t *testing.T) {
	helpers.SetJWTSecret(t, "callback-login-success-secret")

	state, _, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin), "", allowedReturnURL, string(commonModel.OAuth2GITHUB),
	)
	require.NoError(t, err)

	svc, repo, authRepo, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
	svc.resolveAdapter = func(string) (oauthProviderAdapter, error) {
		return &fakeAdapter{identity: &oauthIdentity{
			ExternalID: "ext-1",
			AuthType:   string(authModel.AuthTypeOAuth2),
		}}, nil
	}

	user := userModel.User{ID: "u-1", Username: "alice"}
	repo.EXPECT().
		GetUserByOAuthID(mock.Anything, string(commonModel.OAuth2GITHUB), "ext-1").
		Return(user, nil).
		Once()

	var storedCode string
	authRepo.EXPECT().
		StoreOAuthCode(mock.Anything, mock.Anything, 60*time.Second).
		Run(func(code string, _ *authModel.TokenPair, _ time.Duration) { storedCode = code }).
		Once()

	out, err := svc.HandleOAuthCallback(string(commonModel.OAuth2GITHUB), "code-123", state)
	require.NoError(t, err)

	parsed, perr := url.Parse(out)
	require.NoError(t, perr)
	assert.Equal(t, "app.example.com", parsed.Host)
	assert.NotEmpty(t, storedCode)
	assert.Equal(t, storedCode, parsed.Query().Get("code"))
}

// ---------------------------------------------------------------------------
// resolveOAuthCallback：login vs bind、OIDC vs OAuth 查找、token 签发、
// 重定向校验、一次性码存储、bind 事务 success/fail。
// ---------------------------------------------------------------------------

func loginState(redirect string) *authModel.OAuthState {
	return &authModel.OAuthState{
		Action:   string(authModel.OAuth2ActionLogin),
		Redirect: redirect,
		Provider: string(commonModel.OAuth2GITHUB),
		Nonce:    "nonce-x",
		Exp:      time.Now().Add(5 * time.Minute).Unix(),
	}
}

func bindState(userID, redirect string) *authModel.OAuthState {
	return &authModel.OAuthState{
		Action:   string(authModel.OAuth2ActionBind),
		UserID:   userID,
		Redirect: redirect,
		Provider: string(commonModel.OAuth2GITHUB),
		Nonce:    "nonce-x",
		Exp:      time.Now().Add(5 * time.Minute).Unix(),
	}
}

func TestResolveOAuthCallback_PureValidation(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-validation-secret")

	cases := []struct {
		name  string
		state *authModel.OAuthState
	}{
		{
			name: "login state must not carry user_id",
			state: &authModel.OAuthState{
				Action:   string(authModel.OAuth2ActionLogin),
				UserID:   "u-should-not-be-here",
				Redirect: allowedReturnURL,
				Provider: string(commonModel.OAuth2GITHUB),
			},
		},
		{
			name: "bind state must carry user_id",
			state: &authModel.OAuthState{
				Action:   string(authModel.OAuth2ActionBind),
				UserID:   "",
				Redirect: allowedReturnURL,
				Provider: string(commonModel.OAuth2GITHUB),
			},
		},
		{
			name: "unknown action",
			state: &authModel.OAuthState{
				Action:   "frobnicate",
				Redirect: allowedReturnURL,
				Provider: string(commonModel.OAuth2GITHUB),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 这些分支在触达任何协作者前返回；mock 无期望即反证未被调用。
			svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

			out, err := svc.resolveOAuthCallback(
				tc.state, string(commonModel.OAuth2GITHUB), "ext-1", "", string(authModel.AuthTypeOAuth2),
			)
			require.EqualError(t, err, commonModel.INVALID_PARAMS)
			assert.Empty(t, out)
		})
	}
}

func TestResolveOAuthCallback_LoginOAuthSuccess(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-login-oauth-secret")

	svc, repo, authRepo, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	user := userModel.User{ID: "u-1", Username: "alice"}
	// OAuth（非 OIDC）走 GetUserByOAuthID，不带 issuer。
	repo.EXPECT().
		GetUserByOAuthID(mock.Anything, string(commonModel.OAuth2GITHUB), "ext-oauth").
		Return(user, nil).
		Once()

	var storedCode string
	authRepo.EXPECT().
		StoreOAuthCode(mock.Anything, mock.Anything, 60*time.Second).
		Run(func(code string, pair *authModel.TokenPair, _ time.Duration) {
			storedCode = code
			require.NotNil(t, pair)
			assert.NotEmpty(t, pair.AccessToken)
		}).
		Once()

	out, err := svc.resolveOAuthCallback(
		loginState(allowedReturnURL),
		string(commonModel.OAuth2GITHUB), "ext-oauth", "", string(authModel.AuthTypeOAuth2),
	)
	require.NoError(t, err)

	parsed, perr := url.Parse(out)
	require.NoError(t, perr)
	assert.Equal(t, storedCode, parsed.Query().Get("code"))
}

func TestResolveOAuthCallback_LoginOIDCSuccess(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-login-oidc-secret")

	svc, repo, authRepo, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	user := userModel.User{ID: "u-9", Username: "oidcuser"}
	// OIDC 走 GetUserByOIDC，按 (provider, externalID, issuer) 三元组查找。
	repo.EXPECT().
		GetUserByOIDC(mock.Anything, string(commonModel.OAuth2GITHUB), "sub-123", "https://idp.example.com").
		Return(user, nil).
		Once()
	authRepo.EXPECT().
		StoreOAuthCode(mock.Anything, mock.Anything, 60*time.Second).
		Return().
		Once()

	out, err := svc.resolveOAuthCallback(
		loginState(allowedReturnURL),
		string(commonModel.OAuth2GITHUB), "sub-123", "https://idp.example.com", string(authModel.AuthTypeOIDC),
	)
	require.NoError(t, err)
	assert.Contains(t, out, "code=")
}

func TestResolveOAuthCallback_LoginLookupFailure(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-login-lookup-fail-secret")

	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	notBound := errors.New("identity not bound")
	repo.EXPECT().
		GetUserByOAuthID(mock.Anything, string(commonModel.OAuth2GITHUB), "ext-unbound").
		Return(userModel.User{}, notBound).
		Once()
	// 查找失败时不应签发 code（StoreOAuthCode 无期望即反证）。

	out, err := svc.resolveOAuthCallback(
		loginState(allowedReturnURL),
		string(commonModel.OAuth2GITHUB), "ext-unbound", "", string(authModel.AuthTypeOAuth2),
	)
	require.ErrorIs(t, err, notBound)
	assert.Empty(t, out)
}

// 登录成功签发 token 后，若 redirect 不在白名单内则拒绝，且一次性 code 不得落库。
func TestResolveOAuthCallback_LoginRedirectRejected(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-login-redirect-reject-secret")

	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	repo.EXPECT().
		GetUserByOAuthID(mock.Anything, string(commonModel.OAuth2GITHUB), "ext-1").
		Return(userModel.User{ID: "u-1", Username: "alice"}, nil).
		Once()
	// StoreOAuthCode 必须在重定向校验之后；校验失败时它不应被调用。

	out, err := svc.resolveOAuthCallback(
		loginState("https://evil.example.net/auth"),
		string(commonModel.OAuth2GITHUB), "ext-1", "", string(authModel.AuthTypeOAuth2),
	)
	require.EqualError(t, err, commonModel.INVALID_PARAMS)
	assert.Empty(t, out)
}

func TestResolveOAuthCallback_BindSuccess(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-bind-success-secret")

	svc, repo, _, tx := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	runsTxInline(tx)
	repo.EXPECT().
		BindOAuth(mock.Anything, "u-7", string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2)).
		Return(nil).
		Once()

	out, err := svc.resolveOAuthCallback(
		bindState("u-7", allowedReturnURL),
		string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2),
	)
	require.NoError(t, err)

	parsed, perr := url.Parse(out)
	require.NoError(t, perr)
	assert.Equal(t, "app.example.com", parsed.Host)
	assert.Equal(t, "success", parsed.Query().Get("bind"))
}

func TestResolveOAuthCallback_BindPersistFailure(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-bind-fail-secret")

	svc, repo, _, tx := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	runsTxInline(tx)
	persistErr := errors.New("unique constraint violation")
	repo.EXPECT().
		BindOAuth(mock.Anything, "u-7", string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2)).
		Return(persistErr).
		Once()

	out, err := svc.resolveOAuthCallback(
		bindState("u-7", allowedReturnURL),
		string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2),
	)
	require.ErrorIs(t, err, persistErr)
	assert.Empty(t, out)
}

// bind 持久化成功后 redirect 不合法：事务已提交，但仍返回 INVALID_PARAMS。
func TestResolveOAuthCallback_BindRedirectRejected(t *testing.T) {
	helpers.SetJWTSecret(t, "resolve-bind-redirect-reject-secret")

	svc, repo, _, tx := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))

	runsTxInline(tx)
	repo.EXPECT().
		BindOAuth(mock.Anything, "u-7", string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2)).
		Return(nil).
		Once()

	out, err := svc.resolveOAuthCallback(
		bindState("u-7", "https://evil.example.net/panel"),
		string(commonModel.OAuth2GITHUB), "ext-bind", "", string(authModel.AuthTypeOAuth2),
	)
	require.EqualError(t, err, commonModel.INVALID_PARAMS)
	assert.Empty(t, out)
}
