package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bytedance/sonic"
	"github.com/lin-snow/ech0/internal/config"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const (
	minSubmitMS       int64 = 2000
	maxFormTokenHours int64 = 24
	maxCommentRunes         = 200
)

type CommentService struct {
	commonService      CommonService
	repo               Repository
	keyvalueRepository KeyValueRepository
}

func NewCommentService(
	commonService CommonService,
	repo Repository,
	keyvalueRepository KeyValueRepository,
) *CommentService {
	return &CommentService{
		commonService:      commonService,
		repo:               repo,
		keyvalueRepository: keyvalueRepository,
	}
}

func (s *CommentService) GetFormMeta(ctx context.Context, clientIP string) (model.FormMeta, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return model.FormMeta{}, err
	}
	issuedAt := time.Now().UnixMilli()
	token := s.signFormToken(clientIP, issuedAt)
	return model.FormMeta{
		FormToken:      token,
		MinSubmitMs:    minSubmitMS,
		CaptchaEnabled: setting.CaptchaEnabled && strings.TrimSpace(setting.CaptchaVerify) != "",
		EnableComment:  setting.EnableComment,
	}, nil
}

func (s *CommentService) CreateComment(
	ctx context.Context,
	clientIP string,
	userAgent string,
	dto *model.CreateCommentDto,
) error {
	if strings.TrimSpace(dto.HoneypotField) != "" {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "提交被拒绝")
	}
	if err := s.verifyFormToken(clientIP, dto.FormToken); err != nil {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "提交过快或表单已失效")
	}

	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return err
	}
	if !setting.EnableComment {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论功能未启用")
	}

	if setting.CaptchaEnabled && strings.TrimSpace(setting.CaptchaVerify) != "" {
		if err := s.verifyCaptcha(dto.CaptchaToken, clientIP, setting); err != nil {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "验证码验证失败")
		}
	}

	user, validUser, err := s.resolveRequestUser(ctx)
	if err != nil {
		return err
	}

	comment := model.Comment{
		EchoID:    strings.TrimSpace(dto.EchoID),
		Content:   strings.TrimSpace(dto.Content),
		IPHash:    hashClientIP(clientIP),
		UserAgent: strings.TrimSpace(userAgent),
		Status:    model.StatusPending,
		Source:    model.SourceGuest,
	}
	if comment.EchoID == "" || comment.Content == "" {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能为空")
	}
	if utf8.RuneCountInString(comment.Content) > maxCommentRunes {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能超过200字")
	}

	if validUser && (user.IsAdmin || user.IsOwner) {
		comment.Source = model.SourceSystem
		comment.Nickname = user.Username
		// 内部成员评论允许邮箱为空，不再自动填充占位邮箱。
		comment.Email = ""
		comment.AvatarURL = strings.TrimSpace(user.Avatar)
		if comment.AvatarURL == "" {
			comment.AvatarURL = buildDiceBearURL(user.Username)
		}
		comment.UserID = &user.ID
		comment.Status = model.StatusApproved
	} else {
		nickname := strings.TrimSpace(dto.Nickname)
		email := strings.TrimSpace(dto.Email)
		website := strings.TrimSpace(dto.Website)
		if nickname == "" || email == "" {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "昵称和邮箱不能为空")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "邮箱格式无效")
		}
		if website != "" {
			parsed, err := url.ParseRequestURI(website)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "网址格式无效")
			}
		}
		comment.Nickname = nickname
		comment.Email = email
		comment.Website = website
		comment.AvatarURL = buildDiceBearURL(nickname)

		if !setting.RequireApproval {
			comment.Status = model.StatusApproved
		}

		if err := s.checkRateLimit(ctx, comment.IPHash, email, ""); err != nil {
			return err
		}
	}

	return s.repo.CreateComment(ctx, &comment)
}

func (s *CommentService) ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.EnableComment {
		return []model.Comment{}, nil
	}
	return s.repo.ListPublicByEchoID(ctx, strings.TrimSpace(echoID))
}

func (s *CommentService) ListPanelComments(
	ctx context.Context,
	query model.ListCommentQuery,
) (model.PageResult[model.Comment], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return model.PageResult[model.Comment]{}, err
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.repo.ListComments(ctx, query)
}

func (s *CommentService) GetCommentByID(ctx context.Context, id string) (model.Comment, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return model.Comment{}, err
	}
	return s.repo.GetCommentByID(ctx, id)
}

func (s *CommentService) UpdateCommentStatus(ctx context.Context, id string, status model.Status) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if status != model.StatusPending && status != model.StatusApproved && status != model.StatusRejected {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "无效的评论状态")
	}
	return s.repo.UpdateCommentStatus(ctx, id, status)
}

func (s *CommentService) DeleteComment(ctx context.Context, id string) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return s.repo.DeleteComment(ctx, id)
}

func (s *CommentService) BatchAction(ctx context.Context, action string, ids []string) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	switch action {
	case "approve":
		return s.repo.BatchUpdateStatus(ctx, ids, model.StatusApproved)
	case "reject":
		return s.repo.BatchUpdateStatus(ctx, ids, model.StatusRejected)
	case "delete":
		return s.repo.BatchDelete(ctx, ids)
	default:
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "无效的批量动作")
	}
}

func (s *CommentService) GetSystemSetting(ctx context.Context) (model.SystemSetting, error) {
	raw, err := s.keyvalueRepository.GetKeyValue(ctx, model.CommentSystemSettingKey)
	if err != nil {
		defaultSetting := model.SystemSetting{
			EnableComment:   true,
			RequireApproval: true,
			CaptchaEnabled:  false,
			CaptchaVerify:   "",
			CaptchaSecret:   "",
		}
		buf, _ := sonic.Marshal(defaultSetting)
		_ = s.keyvalueRepository.AddKeyValue(ctx, model.CommentSystemSettingKey, string(buf))
		return defaultSetting, nil
	}
	var setting model.SystemSetting
	if err := sonic.Unmarshal([]byte(raw), &setting); err != nil {
		return model.SystemSetting{}, err
	}
	return setting, nil
}

func (s *CommentService) UpdateSystemSetting(ctx context.Context, setting model.SystemSetting) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	setting.CaptchaVerify = strings.TrimSpace(setting.CaptchaVerify)
	setting.CaptchaSecret = strings.TrimSpace(setting.CaptchaSecret)
	buf, err := sonic.Marshal(setting)
	if err != nil {
		return err
	}
	return s.keyvalueRepository.AddOrUpdateKeyValue(ctx, model.CommentSystemSettingKey, string(buf))
}

func (s *CommentService) checkRateLimit(ctx context.Context, ipHash, email, userID string) error {
	shortWindow, longWindow := int64(60), int64(3600)
	shortLimit, longLimit := int64(3), int64(20)

	ipShort, err := s.repo.CountByIPWithin(ctx, ipHash, shortWindow)
	if err != nil {
		return err
	}
	ipLong, err := s.repo.CountByIPWithin(ctx, ipHash, longWindow)
	if err != nil {
		return err
	}
	if ipShort >= shortLimit || ipLong >= longLimit {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
	}

	emailShort, err := s.repo.CountByEmailWithin(ctx, email, shortWindow)
	if err != nil {
		return err
	}
	if emailShort >= shortLimit {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
	}

	if userID != "" {
		userShort, err := s.repo.CountByUserWithin(ctx, userID, shortWindow)
		if err != nil {
			return err
		}
		if userShort >= shortLimit {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
		}
	}

	return nil
}

func (s *CommentService) verifyCaptcha(token, clientIP string, setting model.SystemSetting) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("captcha token missing")
	}
	payload := map[string]string{
		"token":     strings.TrimSpace(token),
		"secret":    setting.CaptchaSecret,
		"remote_ip": clientIP,
	}
	body, err := sonic.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		setting.CaptchaVerify,
		strings.NewReader(string(body)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var out map[string]any
	if err := sonic.ConfigFastest.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if ok, _ := out["success"].(bool); ok {
		return nil
	}
	if ok, _ := out["ok"].(bool); ok {
		return nil
	}
	return errors.New("captcha verify failed")
}

func (s *CommentService) verifyFormToken(clientIP, token string) error {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return errors.New("token invalid")
	}
	issuedAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	nowMS := time.Now().UnixMilli()
	if nowMS-issuedAt < minSubmitMS {
		return errors.New("submit too fast")
	}
	if nowMS-issuedAt > maxFormTokenHours*3600*1000 {
		return errors.New("token expired")
	}
	expectSig := s.computeHMAC(fmt.Sprintf("%s:%d", clientIP, issuedAt))
	if !hmac.Equal([]byte(parts[1]), []byte(expectSig)) {
		return errors.New("token sign mismatch")
	}
	return nil
}

func (s *CommentService) signFormToken(clientIP string, issuedAt int64) string {
	sig := s.computeHMAC(fmt.Sprintf("%s:%d", clientIP, issuedAt))
	return fmt.Sprintf("%d.%s", issuedAt, sig)
}

func (s *CommentService) computeHMAC(payload string) string {
	mac := hmac.New(sha256.New, config.Config().Security.JWTSecret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *CommentService) requireAdmin(ctx context.Context) error {
	v := viewer.MustFromContext(ctx)
	if v == nil || strings.TrimSpace(v.UserID()) == "" {
		return commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	}
	user, err := s.commonService.CommonGetUserByUserId(ctx, v.UserID())
	if err != nil {
		return err
	}
	if !user.IsAdmin && !user.IsOwner {
		return commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

func (s *CommentService) resolveRequestUser(ctx context.Context) (user userModel.User, valid bool, err error) {
	v := viewer.MustFromContext(ctx)
	if v == nil || strings.TrimSpace(v.UserID()) == "" {
		return user, false, nil
	}
	u, err := s.commonService.CommonGetUserByUserId(ctx, v.UserID())
	if err != nil {
		return user, false, nil
	}
	return u, true, nil
}

func ParseOptionalUserIDFromAuthHeader(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	claims, err := jwtUtil.ParseToken(strings.TrimSpace(parts[1]))
	if err != nil {
		return ""
	}
	return claims.Userid
}

func hashClientIP(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(ip)))
	return hex.EncodeToString(sum[:])
}

func buildDiceBearURL(seed string) string {
	trimmed := strings.TrimSpace(seed)
	if trimmed == "" {
		trimmed = "guest"
	}
	return "https://api.dicebear.com/9.x/fun-emoji/svg?seed=" + url.QueryEscape(trimmed)
}
