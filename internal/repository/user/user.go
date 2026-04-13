package repository

import (
	"context"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type UserRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

var _ userService.Repository = (*UserRepository)(nil)

func NewUserRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) *UserRepository {
	return &UserRepository{
		db:    dbProvider,
		cache: cache,
	}
}

// getDB 从上下文中获取事务
func (userRepository *UserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return userRepository.db()
}

// GetUserByUsername 根据用户名获取用户
func (userRepository *UserRepository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	cacheKey := GetUsernameKey(username)
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			user := model.User{}
			err := userRepository.getDB(ctx).Where("username = ?", username).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
		func() (model.User, error) {
			user := model.User{}
			err := userRepository.db().Where("username = ?", username).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
	)
}

// GetAllUsers 获取所有用户
func (userRepository *UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := userRepository.getDB(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser 创建一个新的用户
func (userRepository *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := userRepository.getDB(ctx).Create(user).Error
	if err != nil {
		return err
	}

	// 加入缓存
	userRepository.cache.Set(GetUserIDKey(user.ID), *user, 1)
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)
	if user.IsOwner {
		userRepository.cache.Set(GetOwnerKey(), *user, 1)
	}

	return nil
}

// GetUserByID 根据用户ID获取用户
func (userRepository *UserRepository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	cacheKey := GetUserIDKey(id)
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			var user model.User
			if err := userRepository.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
				return user, err
			}
			return user, nil
		},
		func() (model.User, error) {
			var user model.User
			if err := userRepository.db().Where("id = ?", id).First(&user).Error; err != nil {
				return user, err
			}
			return user, nil
		})
}

// GetOwner 获取Owner
func (userRepository *UserRepository) GetOwner(ctx context.Context) (model.User, error) {
	cacheKey := GetOwnerKey()
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			user := model.User{}
			err := userRepository.getDB(ctx).Where("is_owner = ?", true).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
		func() (model.User, error) {
			user := model.User{}
			err := userRepository.db().Where("is_owner = ?", true).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		})
}

func (userRepository *UserRepository) IsInitialized(ctx context.Context) (bool, error) {
	var kv commonModel.KeyValue
	err := userRepository.getDB(ctx).Where("key = ?", commonModel.InstallInitializedKey).First(&kv).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return kv.Value == "true", nil
}

func (userRepository *UserRepository) MarkInitialized(ctx context.Context) error {
	result := userRepository.getDB(ctx).
		Model(&commonModel.KeyValue{}).
		Where("key = ?", commonModel.InstallInitializedKey).
		Update("value", "true")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return userRepository.getDB(ctx).Create(&commonModel.KeyValue{
		Key:   commonModel.InstallInitializedKey,
		Value: "true",
	}).Error
}

// UpdateUser 更新用户信息
func (userRepository *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	var existing model.User
	if err := userRepository.getDB(ctx).Where("id = ?", user.ID).First(&existing).Error; err != nil {
		return err
	}

	err := userRepository.getDB(ctx).Save(user).Error
	if err != nil {
		return err
	}

	userRepository.cache.Set(GetUserIDKey(user.ID), *user, 1)
	if existing.Username != "" && existing.Username != user.Username {
		userRepository.cache.Delete(GetUsernameKey(existing.Username))
	}
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)
	if existing.IsAdmin && !user.IsAdmin {
		userRepository.cache.Delete(GetAdminKey(user.ID))
	}
	if user.IsAdmin {
		userRepository.cache.Set(GetAdminKey(user.ID), *user, 1)
	}
	if existing.IsOwner && !user.IsOwner {
		userRepository.cache.Delete(GetOwnerKey())
	}
	if user.IsOwner {
		userRepository.cache.Set(GetOwnerKey(), *user, 1)
	}

	return nil
}

// DeleteUser 删除用户
func (userRepository *UserRepository) DeleteUser(ctx context.Context, id string) error {
	// 先查找待删除的用户
	userToDel, err := userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	err = userRepository.getDB(ctx).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// 清空缓存
	userRepository.cache.Delete(GetUserIDKey(userToDel.ID))
	userRepository.cache.Delete(GetUsernameKey(userToDel.Username))
	if userToDel.IsAdmin {
		userRepository.cache.Delete(GetAdminKey(userToDel.ID))
	}
	if userToDel.IsOwner {
		userRepository.cache.Delete(GetOwnerKey())
	}

	return nil
}

// BindOAuth 绑定 OAuth 或 OIDC 账号
func (userRepository *UserRepository) BindOAuth(
	ctx context.Context,
	userID string,
	provider, oauthID, issuer, authType string,
) error {
	protocol := string(authModel.AuthTypeOAuth2)
	issuerVal := ""
	if authType == string(authModel.AuthTypeOIDC) {
		protocol = string(authModel.AuthTypeOIDC)
		issuerVal = strings.TrimSpace(issuer)
	}

	var identity model.UserExternalIdentity
	err := userRepository.getDB(ctx).
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userID, provider, issuerVal, protocol).
		First(&identity).Error
	if err == nil {
		identity.Subject = oauthID
		return userRepository.getDB(ctx).Save(&identity).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	identity = model.UserExternalIdentity{
		UserID:   userID,
		Provider: provider,
		Subject:  oauthID,
		Issuer:   issuerVal,
		Protocol: protocol,
	}
	return userRepository.getDB(ctx).Create(&identity).Error
}

// GetUserByOAuthID 根据 OAuth 提供商和 OAuth ID 获取用户
func (userRepository *UserRepository) GetUserByOAuthID(
	ctx context.Context,
	provider, oauthID string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := userRepository.getDB(ctx).
		Where("provider = ? AND subject = ? AND protocol = ?", provider, oauthID, string(authModel.AuthTypeOAuth2)).
		First(&binding).
		Error
	if err != nil {
		return model.User{}, err
	}

	return userRepository.GetUserByID(ctx, binding.UserID)
}

// GetUserByOIDC 根据 OIDC 提供商、issuer 与 sub 获取用户
func (userRepository *UserRepository) GetUserByOIDC(
	ctx context.Context,
	provider, oauthID, issuer string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := userRepository.getDB(ctx).
		Where(
			"provider = ? AND subject = ? AND issuer = ? AND protocol = ?",
			provider,
			oauthID,
			issuer,
			string(authModel.AuthTypeOIDC),
		).
		First(&binding).Error
	if err != nil {
		return model.User{}, err
	}

	return userRepository.GetUserByID(ctx, binding.UserID)
}

// GetOAuthInfo 获取 OAuth2 信息
func (userRepository *UserRepository) GetOAuthInfo(
	userId string,
	provider string,
) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := userRepository.db().
		Where("user_id = ? AND provider = ? AND protocol = ?", userId, provider, string(authModel.AuthTypeOAuth2)).
		First(&identity).Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

// GetOAuthOIDCInfo 获取 OIDC 信息
func (userRepository *UserRepository) GetOAuthOIDCInfo(
	userId string,
	provider string,
	issuer string,
) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := userRepository.db().
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userId, provider, issuer, string(authModel.AuthTypeOIDC)).
		First(&identity).
		Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

// -----------------------
// Passkey / WebAuthn
// -----------------------

func (userRepository *UserRepository) CreatePasskey(
	ctx context.Context,
	passkey *authModel.Passkey,
) error {
	return userRepository.getDB(ctx).Create(&model.WebAuthnCredential{
		ID:             passkey.ID,
		UserID:         passkey.UserID,
		CredentialID:   passkey.CredentialID,
		CredentialJSON: passkey.CredentialJSON,
		PublicKey:      passkey.PublicKey,
		SignCount:      passkey.SignCount,
		LastUsedAt:     passkey.LastUsedAt,
		DeviceName:     passkey.DeviceName,
		AAGUID:         passkey.AAGUID,
		CreatedAt:      passkey.CreatedAt,
		UpdatedAt:      passkey.UpdatedAt,
	}).Error
}

func (userRepository *UserRepository) ListPasskeysByUserID(
	userID string,
) ([]authModel.Passkey, error) {
	var rows []model.WebAuthnCredential
	if err := userRepository.db().
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	passkeys := make([]authModel.Passkey, 0, len(rows))
	for _, row := range rows {
		passkeys = append(passkeys, authModel.Passkey{
			ID:             row.ID,
			UserID:         row.UserID,
			CredentialID:   row.CredentialID,
			CredentialJSON: row.CredentialJSON,
			PublicKey:      row.PublicKey,
			SignCount:      row.SignCount,
			LastUsedAt:     row.LastUsedAt,
			DeviceName:     row.DeviceName,
			AAGUID:         row.AAGUID,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return passkeys, nil
}

func (userRepository *UserRepository) GetPasskeyByCredentialID(
	credentialID string,
) (authModel.Passkey, error) {
	var row model.WebAuthnCredential
	if err := userRepository.db().
		Where("credential_id = ?", credentialID).
		First(&row).Error; err != nil {
		return authModel.Passkey{}, err
	}
	return authModel.Passkey{
		ID:             row.ID,
		UserID:         row.UserID,
		CredentialID:   row.CredentialID,
		CredentialJSON: row.CredentialJSON,
		PublicKey:      row.PublicKey,
		SignCount:      row.SignCount,
		LastUsedAt:     row.LastUsedAt,
		DeviceName:     row.DeviceName,
		AAGUID:         row.AAGUID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func (userRepository *UserRepository) UpdatePasskeyUsage(
	ctx context.Context,
	passkeyID string,
	signCount uint32,
	lastUsedAt int64,
) error {
	return userRepository.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ?", passkeyID).
		Updates(map[string]any{
			"sign_count":   signCount,
			"last_used_at": lastUsedAt,
		}).Error
}

func (userRepository *UserRepository) UpdatePasskeyDeviceName(
	ctx context.Context,
	userID, passkeyID string,
	deviceName string,
) error {
	return userRepository.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Update("device_name", deviceName).Error
}

func (userRepository *UserRepository) DeletePasskeyByID(
	ctx context.Context,
	userID, passkeyID string,
) error {
	return userRepository.getDB(ctx).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Delete(&model.WebAuthnCredential{}).Error
}

func (userRepository *UserRepository) CacheSetPasskeySession(
	key string,
	val any,
	ttl time.Duration,
) {
	_ = userRepository.cache.SetWithTTL(key, val, 1, ttl)
}

func (userRepository *UserRepository) CacheGetPasskeySession(key string) (any, error) {
	value, found, err := userRepository.cache.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (userRepository *UserRepository) CacheDeletePasskeySession(key string) {
	userRepository.cache.Delete(key)
}
