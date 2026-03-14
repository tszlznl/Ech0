package i18n

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	ContextLocaleKey    = "locale"
	ContextLocalizerKey = "localizer"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	supportedLocales = []language.Tag{
		language.MustParse("zh-CN"),
		language.MustParse("en-US"),
	}
	matcher = language.NewMatcher(supportedLocales)
	bundle  = newBundle()
)

func newBundle() *goi18n.Bundle {
	b := goi18n.NewBundle(language.MustParse("zh-CN"))
	b.RegisterUnmarshalFunc("json", unmarshalJSON)
	for _, path := range []string{"locales/zh-CN.json", "locales/en-US.json"} {
		_, err := b.LoadMessageFileFS(localeFS, path)
		if err != nil {
			panic(fmt.Sprintf("load i18n message file failed: %s: %v", path, err))
		}
	}
	return b
}

func ResolveLocale(raw ...string) string {
	parts := make([]string, 0, len(raw))
	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return "zh-CN"
	}
	tag, _, _ := language.ParseAcceptLanguage(strings.Join(parts, ","))
	best, _, _ := matcher.Match(tag...)
	base, _ := best.Base()
	region, _ := best.Region()
	if region.IsCountry() {
		return fmt.Sprintf("%s-%s", strings.ToLower(base.String()), strings.ToUpper(region.String()))
	}
	return best.String()
}

func NewLocalizer(locale, acceptLanguage string) *goi18n.Localizer {
	return goi18n.NewLocalizer(bundle, locale, acceptLanguage)
}

func Localize(localizer *goi18n.Localizer, messageID string, defaultText string, templateData map[string]any) string {
	if localizer == nil || strings.TrimSpace(messageID) == "" {
		return defaultText
	}
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &goi18n.Message{
			ID:    messageID,
			Other: defaultText,
		},
		TemplateData: templateData,
	})
	if err != nil {
		return defaultText
	}
	return msg
}

func LocalizerFromGin(ctx *gin.Context) *goi18n.Localizer {
	if ctx == nil {
		return nil
	}
	if v, ok := ctx.Get(ContextLocalizerKey); ok {
		if localizer, ok := v.(*goi18n.Localizer); ok {
			return localizer
		}
	}
	return nil
}

func LocaleFromGin(ctx *gin.Context) string {
	if ctx == nil {
		return "zh-CN"
	}
	if v, ok := ctx.Get(ContextLocaleKey); ok {
		if locale, ok := v.(string); ok && locale != "" {
			return locale
		}
	}
	return "zh-CN"
}

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		explicit := strings.TrimSpace(ctx.Query("lang"))
		if explicit == "" {
			explicit = strings.TrimSpace(ctx.GetHeader("X-Locale"))
		}
		acceptLanguage := strings.TrimSpace(ctx.GetHeader("Accept-Language"))
		locale := ResolveLocale(explicit, acceptLanguage)
		localizer := NewLocalizer(locale, acceptLanguage)
		ctx.Set(ContextLocaleKey, locale)
		ctx.Set(ContextLocalizerKey, localizer)
		ctx.Header("Content-Language", locale)
		ctx.Next()
	}
}

func HeaderLocale(req *http.Request) string {
	if req == nil {
		return "zh-CN"
	}
	return ResolveLocale(req.Header.Get("X-Locale"), req.Header.Get("Accept-Language"))
}
