package service

import (
	"context"
	"encoding/json"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

type Service interface {
	Login(loginDto *authModel.LoginDto) (string, error)
	Register(registerDto *authModel.RegisterDto) error
	UpdateUser(ctx context.Context, userdto model.UserInfoDto) error
	UpdateUserAdmin(ctx context.Context, id string) error
	GetAllUsers() ([]model.User, error)
	GetSysAdmin() (model.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByID(userId string) (model.User, error)
	BindOAuth(ctx context.Context, provider string, redirectURI string) (string, error)
	GetOAuthLoginURL(provider string, redirectURI string) (string, error)
	HandleOAuthCallback(provider string, code string, state string) string
	GetOAuthInfo(ctx context.Context, provider string) (model.OAuthInfoDto, error)
	PasskeyRegisterBegin(ctx context.Context, rpID, origin, deviceName string) (authModel.PasskeyRegisterBeginResp, error)
	PasskeyRegisterFinish(ctx context.Context, rpID, origin, nonce string, credential json.RawMessage) error
	PasskeyLoginBegin(rpID, origin string) (authModel.PasskeyLoginBeginResp, error)
	PasskeyLoginFinish(rpID, origin, nonce string, credential json.RawMessage) (string, error)
	ListPasskeys(ctx context.Context) ([]authModel.PasskeyDeviceDto, error)
	DeletePasskey(ctx context.Context, passkeyID string) error
	UpdatePasskeyDeviceName(ctx context.Context, passkeyID string, deviceName string) error
}

type SettingService = settingService.Service

type Repository interface {
	GetUserByID(ctx context.Context, id string) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetAllUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, newUser *model.User) error
	GetSysAdmin(ctx context.Context) (model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	BindOAuth(ctx context.Context, userID string, provider, oauthID, issuer, authType string) error
	GetUserByOAuthID(ctx context.Context, provider, oauthID string) (model.User, error)
	GetUserByOIDC(ctx context.Context, provider, oauthID, issuer string) (model.User, error)
	GetOAuthInfo(userId string, provider string) (model.OAuthBinding, error)
	GetOAuthOIDCInfo(userId string, provider string, issuer string) (model.OAuthBinding, error)
	CreatePasskey(ctx context.Context, passkey *authModel.Passkey) error
	ListPasskeysByUserID(userID string) ([]authModel.Passkey, error)
	GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error)
	UpdatePasskeyUsage(ctx context.Context, passkeyID string, signCount uint32, lastUsedAt time.Time) error
	UpdatePasskeyDeviceName(ctx context.Context, userID, passkeyID string, deviceName string) error
	DeletePasskeyByID(ctx context.Context, userID, passkeyID string) error
	CacheSetPasskeySession(key string, val any, ttl time.Duration)
	CacheGetPasskeySession(key string) (any, error)
	CacheDeletePasskeySession(key string)
}
