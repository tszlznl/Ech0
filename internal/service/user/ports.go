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
	UpdateUser(userid uint, userdto model.UserInfoDto) error
	UpdateUserAdmin(userid uint, id uint) error
	GetAllUsers() ([]model.User, error)
	GetSysAdmin() (model.User, error)
	DeleteUser(userid, id uint) error
	GetUserByID(userId int) (model.User, error)
	BindOAuth(userID uint, provider string, redirectURI string) (string, error)
	GetOAuthLoginURL(provider string, redirectURI string) (string, error)
	HandleOAuthCallback(provider string, code string, state string) string
	GetOAuthInfo(userId uint, provider string) (model.OAuthInfoDto, error)
	PasskeyRegisterBegin(userID uint, rpID, origin, deviceName string) (authModel.PasskeyRegisterBeginResp, error)
	PasskeyRegisterFinish(userID uint, rpID, origin, nonce string, credential json.RawMessage) error
	PasskeyLoginBegin(rpID, origin string) (authModel.PasskeyLoginBeginResp, error)
	PasskeyLoginFinish(rpID, origin, nonce string, credential json.RawMessage) (string, error)
	ListPasskeys(userID uint) ([]authModel.PasskeyDeviceDto, error)
	DeletePasskey(userID, passkeyID uint) error
	UpdatePasskeyDeviceName(userID, passkeyID uint, deviceName string) error
}

type SettingService = settingService.Service

type Repository interface {
	GetUserByID(ctx context.Context, id int) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetAllUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, newUser *model.User) error
	GetSysAdmin(ctx context.Context) (model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uint) error
	BindOAuth(ctx context.Context, userID uint, provider, oauthID, issuer, authType string) error
	GetUserByOAuthID(ctx context.Context, provider, oauthID string) (model.User, error)
	GetUserByOIDC(ctx context.Context, provider, oauthID, issuer string) (model.User, error)
	GetOAuthInfo(userId uint, provider string) (model.OAuthBinding, error)
	GetOAuthOIDCInfo(userId uint, provider string, issuer string) (model.OAuthBinding, error)
	CreatePasskey(ctx context.Context, passkey *authModel.Passkey) error
	ListPasskeysByUserID(userID uint) ([]authModel.Passkey, error)
	GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error)
	UpdatePasskeyUsage(ctx context.Context, passkeyID uint, signCount uint32, lastUsedAt time.Time) error
	UpdatePasskeyDeviceName(ctx context.Context, userID, passkeyID uint, deviceName string) error
	DeletePasskeyByID(ctx context.Context, userID, passkeyID uint) error
	CacheSetPasskeySession(key string, val any, ttl time.Duration)
	CacheGetPasskeySession(key string) (any, error)
	CacheDeletePasskeySession(key string)
}
