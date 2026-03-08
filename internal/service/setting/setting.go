package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type SettingService struct {
	transactor         transaction.Transactor
	commonService      CommonService
	storageManager     *storage.Manager
	keyvalueRepository KeyValueRepository
	settingRepository  SettingRepository
	webhookRepository  WebhookRepository
	publisher          *publisher.Publisher
}

func NewSettingService(
	tx transaction.Transactor,
	commonService CommonService,
	storageManager *storage.Manager,
	keyvalueRepository KeyValueRepository,
	settingRepository SettingRepository,
	webhookRepository WebhookRepository,
	publisher *publisher.Publisher,
) *SettingService {
	return &SettingService{
		transactor:         tx,
		commonService:      commonService,
		storageManager:     storageManager,
		keyvalueRepository: keyvalueRepository,
		webhookRepository:  webhookRepository,
		settingRepository:  settingRepository,
		publisher:          publisher,
	}
}

// GetSetting 获取设置
func (settingService *SettingService) GetSetting(setting *model.SystemSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		systemSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.SystemSettingsKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.SiteTitle = config.Config().Setting.SiteTitle
			setting.ServerLogo = config.Config().Setting.ServerLogo
			setting.ServerName = config.Config().Setting.Servername
			setting.ServerURL = config.Config().Setting.Serverurl
			setting.AllowRegister = config.Config().Setting.AllowRegister
			setting.ICPNumber = config.Config().Setting.Icpnumber
			setting.MetingAPI = config.Config().Setting.MetingAPI
			setting.CustomCSS = config.Config().Setting.CustomCSS
			setting.CustomJS = config.Config().Setting.CustomJS

			// 处理 URL
			setting.ServerURL = httpUtil.TrimURL(setting.ServerURL)
			setting.MetingAPI = httpUtil.TrimURL(setting.MetingAPI)

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.SystemSettingsKey, string(settingToJSON)); err != nil {
				return err
			}

			// 处理 ServerURL
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(systemSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateSetting 更新设置
func (settingService *SettingService) UpdateSetting(
	userid string,
	newSetting *model.SystemSettingDto,
) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		var setting model.SystemSetting
		setting.SiteTitle = newSetting.SiteTitle
		setting.ServerLogo = newSetting.ServerLogo
		setting.ServerName = newSetting.ServerName
		setting.ServerURL = httpUtil.TrimURL(newSetting.ServerURL)
		setting.AllowRegister = newSetting.AllowRegister
		setting.ICPNumber = newSetting.ICPNumber
		setting.MetingAPI = httpUtil.TrimURL(newSetting.MetingAPI)
		setting.CustomCSS = newSetting.CustomCSS
		setting.CustomJS = newSetting.CustomJS

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.SystemSettingsKey, settingToJSONString); err != nil {
			return err
		}

		// 更新 ServerURL
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
			return err
		}

		return nil
	})
}

// GetCommentSetting 获取评论设置
func (settingService *SettingService) GetCommentSetting(setting *model.CommentSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		commentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.CommentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.EnableComment = config.Config().Comment.EnableComment
			setting.Provider = config.Config().Comment.Provider
			setting.CommentAPI = config.Config().Comment.CommentAPI

			// 处理 URL
			setting.CommentAPI = httpUtil.TrimURL(setting.CommentAPI)

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(commentSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateCommentSetting 更新评论设置
func (settingService *SettingService) UpdateCommentSetting(
	userid string,
	newSetting *model.CommentSettingDto,
) error {
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 检查评论服务提供者是否有效
		if newSetting.Provider != string(commonModel.TWIKOO) &&
			newSetting.Provider != string(commonModel.ARTALK) &&
			newSetting.Provider != string(commonModel.WALINE) &&
			newSetting.Provider != string(commonModel.GISCUS) {
			return errors.New(commonModel.NO_SUCH_COMMENT_PROVIDER)
		}

		commentSetting := &model.CommentSetting{
			EnableComment: newSetting.EnableComment,
			Provider:      newSetting.Provider,
			CommentAPI:    httpUtil.TrimURL(newSetting.CommentAPI),
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(commentSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

// GetS3Setting 获取 S3 存储设置
func (settingService *SettingService) GetS3Setting(userid string, setting *model.S3Setting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		s3Setting, err := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.S3SettingKey)
		if err != nil {
			// 数据库缺失时回退到 config 默认值
			cfg := config.Config().Storage
			setting.Enable = cfg.ObjectEnabled || storage.NormalizeStorageMode(cfg.Mode) == storage.StorageModeObject
			setting.Provider = strings.TrimSpace(cfg.Provider)
			setting.Endpoint = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(cfg.Endpoint), "http://"), "https://")
			setting.AccessKey = cfg.AccessKey
			setting.SecretKey = cfg.SecretKey
			setting.BucketName = cfg.BucketName
			setting.Region = strings.TrimSpace(cfg.Region)
			setting.UseSSL = cfg.UseSSL
			setting.CDNURL = strings.TrimRight(strings.TrimSpace(cfg.CDNURL), "/")
			setting.PathPrefix = strings.Trim(strings.TrimSpace(cfg.PathPrefix), "/")
			setting.PublicRead = true

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(s3Setting), setting); err != nil {
			return err
		}

		// 如果用户未登录且不为管理员,则屏蔽 S3 设置的敏感信息
		if userid == "" {
			setting.AccessKey = "******"
			setting.SecretKey = "******"
			setting.BucketName = "******"
			setting.Endpoint = "******"
		} else {
			user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				setting.AccessKey = "******"
				setting.SecretKey = "******"
				setting.BucketName = "******"
				setting.Endpoint = "******"
			}
		}

		return nil
	})
}

// UpdateS3Setting 更新 S3 存储设置
func (settingService *SettingService) UpdateS3Setting(
	userid string,
	newSetting *model.S3SettingDto,
) error {
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	oldRaw, _ := settingService.keyvalueRepository.GetKeyValue(context.Background(), commonModel.S3SettingKey)
	var appliedSetting *model.S3Setting

	err = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 检查endpoint是否为http(s)动态改变USE SSL
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "https://") {
			newSetting.UseSSL = true
		} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "http://") {
			newSetting.UseSSL = false
		}

		// 去除Endpoint的协议头（http://或https://）
		endpoint := strings.TrimSpace(newSetting.Endpoint)
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		newSetting.Endpoint = endpoint

		cdnURL := strings.TrimSpace(newSetting.CDNURL)
		if cdnURL != "" {
			cdnURL = strings.TrimRight(cdnURL, "/")
		}

		s3Setting := &model.S3Setting{
			Enable:     newSetting.Enable,
			Provider:   newSetting.Provider,
			Endpoint:   httpUtil.TrimURL(newSetting.Endpoint),
			AccessKey:  newSetting.AccessKey,
			SecretKey:  newSetting.SecretKey,
			BucketName: newSetting.BucketName,
			Region:     strings.TrimSpace(newSetting.Region),
			UseSSL:     newSetting.UseSSL,
			CDNURL:     cdnURL,
			PathPrefix: httpUtil.TrimURL(newSetting.PathPrefix),
			PublicRead: newSetting.PublicRead,
		}

		// 配置检查
		switch s3Setting.Provider {
		case string(commonModel.R2):
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
			s3Setting.UseSSL = true
		case string(commonModel.AWS):
			if s3Setting.Region == "" {
				s3Setting.Region = "us-east-1"
			}
		case string(commonModel.MINIO):
			if s3Setting.Region == "" {
				s3Setting.Region = "us-east-1"
			}
		case string(commonModel.OTHER):
			// 其他 S3 兼容厂商（Backblaze、Wasabi、Ceph 等）
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
		default:
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(s3Setting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.AddOrUpdateKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
			return err
		}

		appliedSetting = s3Setting
		return nil
	})
	if err != nil {
		return err
	}

	if settingService.storageManager != nil && appliedSetting != nil {
		if err := settingService.storageManager.ApplyS3Setting(*appliedSetting); err != nil {
			_ = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
				if strings.TrimSpace(oldRaw) == "" {
					return settingService.keyvalueRepository.DeleteKeyValue(ctx, commonModel.S3SettingKey)
				}
				return settingService.keyvalueRepository.AddOrUpdateKeyValue(ctx, commonModel.S3SettingKey, oldRaw)
			})
			_ = settingService.storageManager.ReloadFromConfigAndDB(context.Background())
			return err
		}
	}

	return nil
}

// GetOAuth2Setting 获取 OAuth2 设置
func (settingService *SettingService) GetOAuth2Setting(
	userid string,
	setting *model.OAuth2Setting,
	forInternal bool,
) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		if !forInternal {
			user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				return errors.New(commonModel.NO_PERMISSION_DENIED)
			}
		}

		oauthSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.OAuth2SettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			setting.Provider = string(commonModel.OAuth2GITHUB)
			setting.ClientID = ""
			setting.ClientSecret = ""
			setting.AuthURL = "https://github.com/login/oauth/authorize"
			setting.TokenURL = "https://github.com/login/oauth/access_token"
			setting.UserInfoURL = "https://api.github.com/user"
			setting.RedirectURI = ""
			setting.Scopes = []string{
				"read:user",
			}
			setting.IsOIDC = false
			setting.Issuer = ""
			setting.JWKSURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(oauthSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateOAuth2Setting 更新 OAuth2 设置
func (settingService *SettingService) UpdateOAuth2Setting(
	userid string,
	newSetting *model.OAuth2SettingDto,
) error {
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		oauthSetting := &model.OAuth2Setting{
			Enable:       newSetting.Enable,
			Provider:     newSetting.Provider,
			ClientID:     newSetting.ClientID,
			ClientSecret: newSetting.ClientSecret,
			AuthURL:      httpUtil.TrimURL(newSetting.AuthURL),
			TokenURL:     httpUtil.TrimURL(newSetting.TokenURL),
			UserInfoURL:  httpUtil.TrimURL(newSetting.UserInfoURL),
			RedirectURI:  httpUtil.TrimURL(newSetting.RedirectURI),
			Scopes:       newSetting.Scopes,
			IsOIDC:       newSetting.IsOIDC,
			Issuer:       newSetting.Issuer,
			JWKSURL:      httpUtil.TrimURL(newSetting.JWKSURL),
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(oauthSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

// GetOAuth2Status 获取 OAuth2 状态
func (settingService *SettingService) GetOAuth2Status(status *model.OAuth2Status) error {
	var oauthSetting model.OAuth2Setting
	if err := settingService.GetOAuth2Setting("", &oauthSetting, true); err != nil {
		return err
	}

	status.Enabled = oauthSetting.Enable
	status.Provider = oauthSetting.Provider

	return nil
}

// GetAllWebhooks 获取所有 Webhook
func (settingService *SettingService) GetAllWebhooks(userid string) ([]webhookModel.Webhook, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	webhooks, err := settingService.webhookRepository.GetAllWebhooks(context.Background())
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

// DeleteWebhook 删除 Webhook
func (settingService *SettingService) DeleteWebhook(userid, id string) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return settingService.webhookRepository.DeleteWebhookByID(ctx, id)
	})
}

// UpdateWebhook 更新 Webhook
func (settingService *SettingService) UpdateWebhook(
	userid, id string,
	newWebhook *model.WebhookDto,
) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		ID:       id,
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 先删除再创建，避免部分字段无法更新的问题
		if err := settingService.webhookRepository.DeleteWebhookByID(ctx, webhook.ID); err != nil {
			return err
		}
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})
}

// CreateWebhook 创建 Webhook
func (settingService *SettingService) CreateWebhook(
	userid string,
	newWebhook *model.WebhookDto,
) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})
}

// ListAccessTokens 列出访问令牌
func (settingService *SettingService) ListAccessTokens(
	userid string,
) ([]model.AccessTokenSetting, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	tokens, err := settingService.settingRepository.ListAccessTokens(context.Background(), user.ID)
	if err != nil {
		return []model.AccessTokenSetting{}, nil
	}

	// 处理tokens,过滤并删除过期的token
	var validTokens []model.AccessTokenSetting
	currentTime := time.Now().UTC()

	for _, token := range tokens {
		if token.Expiry == nil || token.Expiry.After(currentTime) {
			// nil 表示永不过期，或者还没过期
			validTokens = append(validTokens, token)
		} else {
			// 删除过期 token
			_ = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
				return settingService.settingRepository.DeleteAccessTokenByID(ctx, token.ID)
			})
		}
	}

	return validTokens, nil
}

// CreateAccessToken 创建访问令牌
func (settingService *SettingService) CreateAccessToken(
	userid string,
	newToken *model.AccessTokenSettingDto,
) (string, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	name := newToken.Name
	expiry := newToken.Expiry
	var expiryDuration time.Duration

	switch expiry {
	case model.EIGHT_HOUR_EXPIRY:
		expiryDuration = 8 * time.Hour
	case model.ONE_MONTH_EXPIRY:
		expiryDuration = 30 * 24 * time.Hour
	case model.NEVER_EXPIRY:
		expiryDuration = 0
	default:
		expiryDuration = 8 * time.Hour
	}

	// 生成jwt令牌
	claims := jwtUtil.CreateClaimsWithExpiry(user, int64(expiryDuration))
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	// 处理数据库存储的 expiry
	var expiryPtr *time.Time
	if expiry == model.NEVER_EXPIRY {
		expiryPtr = nil // 永不过期，用 NULL
	} else {
		t := time.Now().UTC().Add(expiryDuration)
		expiryPtr = &t
	}

	// 保存到数据库
	accessToken := &model.AccessTokenSetting{
		UserID:    user.ID,
		Token:     tokenString,
		Name:      name,
		Expiry:    expiryPtr,
		CreatedAt: time.Now().UTC(),
	}

	if err := settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return settingService.settingRepository.CreateAccessToken(ctx, accessToken)
	}); err != nil {
		return "", err
	}

	return tokenString, nil
}

// DeleteAccessToken 删除访问令牌
func (settingService *SettingService) DeleteAccessToken(userid, id string) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return settingService.settingRepository.DeleteAccessTokenByID(ctx, id)
	})
}

// GetBackupScheduleSetting 获取备份计划
func (settingService *SettingService) GetBackupScheduleSetting(
	setting *model.BackupSchedule,
) error {
	// 鉴权
	// user, err := settingService.commonService.CommonGetUserByUserId(userid)
	// if err != nil {
	// 	return err
	// }
	// if !user.IsAdmin {
	// 	return errors.New(commonModel.NO_PERMISSION_DENIED)
	// }

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		backupSchedule, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.BackupScheduleKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			// 默认每周日凌晨2点备份
			setting.CronExpression = "0 2 * * 0"

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.BackupScheduleKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(backupSchedule), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateBackupScheduleSetting 更新备份计划
func (settingService *SettingService) UpdateBackupScheduleSetting(
	userid string,
	newSetting *model.BackupScheduleDto,
) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	var updated model.BackupSchedule
	err = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		updated.Enable = newSetting.Enable
		updated.CronExpression = newSetting.CronExpression

		// 验证 Cron 表达式是否合法
		if err := fmtUtil.ValidateCrontabExpression(updated.CronExpression); err != nil {
			return errors.New(commonModel.INVALID_CRON_EXPRESSION)
		}

		settingToJSON, err := jsonUtil.JSONMarshal(updated)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.BackupScheduleKey, settingToJSONString); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 在事务提交后再发布事件，避免回滚时出现幽灵事件。
	if err := settingService.publisher.BackupScheduleUpdated(
		context.Background(),
		contracts.UpdateBackupScheduleEvent{Schedule: updated},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish update backup schedule event", zap.String("error", err.Error()))
	}
	return nil
}

// GetAgentInfo 获取 Agent 信息
func (settingService *SettingService) GetAgentInfo(setting *model.AgentSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		agentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.AgentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，返回默认值
			setting.Enable = false
			setting.Provider = string(commonModel.OpenAI)
			setting.Model = ""
			setting.ApiKey = ""
			setting.Prompt = ""
			setting.BaseURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.AgentSettingKey, string(settingToJSON)); err != nil {
				return err
			}
			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(agentSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// GetAgentSettings 获取 Agent 设置
func (settingService *SettingService) GetAgentSettings(
	userid string,
	setting *model.AgentSetting,
) error {
	// 检查用户权限
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		agentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.AgentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，返回默认值
			setting.Enable = false
			setting.Provider = string(commonModel.OpenAI)
			setting.Model = ""
			setting.ApiKey = ""
			setting.Prompt = ""
			setting.BaseURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.AgentSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(agentSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateAgentSettings 更新 Agent 设置
func (settingService *SettingService) UpdateAgentSettings(
	userid string,
	newSetting *model.AgentSettingDto,
) error {
	// 检查用户权限
	user, err := settingService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if newSetting.Provider != string(commonModel.OpenAI) &&
		newSetting.Provider != string(commonModel.DeepSeek) &&
		newSetting.Provider != string(commonModel.Anthropic) &&
		newSetting.Provider != string(commonModel.Gemini) &&
		newSetting.Provider != string(commonModel.Qwen) &&
		newSetting.Provider != string(commonModel.Ollama) &&
		newSetting.Provider != string(commonModel.Custom) {
		newSetting.Provider = string(commonModel.Custom) // 如果提供商不在列表中，默认为 Custom
	}

	setting := &model.AgentSetting{
		Enable:   newSetting.Enable,
		Provider: newSetting.Provider,
		Model:    newSetting.Model,
		ApiKey:   newSetting.ApiKey,
		Prompt:   newSetting.Prompt,
		BaseURL:  httpUtil.TrimURL(newSetting.BaseURL),
	}

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}
		settingToJSONString := string(settingToJSON)

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.AgentSettingKey, settingToJSONString); err != nil {
			return err
		}

		return nil
	})
}
