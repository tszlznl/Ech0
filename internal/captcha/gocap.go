package captcha

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/gocap/cap"
)

const defaultSiteKey = "ech0-comment"

func SiteKey() string {
	siteKey := strings.TrimSpace(config.Config().Comment.CaptchaSiteKey)
	if siteKey == "" {
		return defaultSiteKey
	}
	return siteKey
}

func Secret() string {
	secret := strings.TrimSpace(config.Config().Comment.CaptchaSecret)
	if secret != "" {
		return secret
	}
	sum := sha256.Sum256(config.Config().Security.JWTSecret)
	return hex.EncodeToString(sum[:])
}

func APIEndpoint() string {
	return "/api/cap/" + SiteKey() + "/"
}

func APIEndpointWithBase(baseURL string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return APIEndpoint()
	}
	return strings.TrimRight(base, "/") + APIEndpoint()
}

func SiteVerifyURL() string {
	siteKey := url.PathEscape(SiteKey())
	return fmt.Sprintf("http://127.0.0.1:%s/api/cap/%s/siteverify", config.Config().Server.Port, siteKey)
}

func NewEngine() (*cap.Engine, error) {
	cfg := config.Config().Comment
	engine, err := cap.New(
		cap.WithInMemoryStore(),
		cap.WithEnableCORS(true),
	)
	if err != nil {
		return nil, err
	}

	if err := engine.RegisterSite(cap.SiteRegistration{
		SiteKey:        SiteKey(),
		Secret:         Secret(),
		Difficulty:     cfg.CaptchaDifficulty,
		ChallengeCount: cfg.CaptchaChallengeCount,
		SaltSize:       cfg.CaptchaSaltSize,
	}); err != nil {
		_ = engine.Close()
		return nil, err
	}
	return engine, nil
}
