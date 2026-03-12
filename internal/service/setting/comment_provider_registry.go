package service

import (
	"errors"
	"fmt"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

type commentProviderSpec struct {
	label   string
	fields  []model.CommentProviderFieldMeta
	def     model.CommentProviderSetting
	check   func(cfg map[string]interface{}) error
	trimURL map[string]bool
}

type commentProviderRegistry map[string]commentProviderSpec

func newCommentProviderRegistry() commentProviderRegistry {
	return commentProviderRegistry{
		string(commonModel.TWIKOO): {
			label: "Twikoo",
			fields: []model.CommentProviderFieldMeta{
				{Key: "envId", Label: "Env ID", Required: true, Placeholder: "Twikoo envId"},
			},
			def: model.CommentProviderSetting{
				ScriptURL: "https://cdn.jsdelivr.net/npm/twikoo@1.7.2/dist/twikoo.all.min.js",
				Config: map[string]interface{}{
					"envId": "",
				},
			},
			check: func(cfg map[string]interface{}) error {
				if strVal(cfg["envId"]) == "" {
					return errors.New("twikoo.envId 不能为空")
				}
				return nil
			},
		},
		string(commonModel.WALINE): {
			label: "Waline",
			fields: []model.CommentProviderFieldMeta{
				{Key: "serverURL", Label: "Server URL", Required: true, Placeholder: "https://your-waline-server"},
				{Key: "path", Label: "Path", Required: false, Placeholder: "/"},
			},
			def: model.CommentProviderSetting{
				ScriptURL: "https://unpkg.com/@waline/client@3.13.0/dist/waline.umd.js",
				CSSURL:    "https://unpkg.com/@waline/client@3.13.0/dist/waline.css",
				Config: map[string]interface{}{
					"serverURL": "",
					"path":      "",
				},
			},
			trimURL: map[string]bool{"serverURL": true},
			check: func(cfg map[string]interface{}) error {
				if strVal(cfg["serverURL"]) == "" {
					return errors.New("waline.serverURL 不能为空")
				}
				return nil
			},
		},
		string(commonModel.ARTALK): {
			label: "Artalk",
			fields: []model.CommentProviderFieldMeta{
				{Key: "server", Label: "Server URL", Required: true, Placeholder: "https://artalk.example.com"},
				{Key: "site", Label: "Site Name", Required: true, Placeholder: "Ech0"},
				{Key: "pageKey", Label: "Page Key", Required: false, Placeholder: "留空时使用当前路径"},
			},
			def: model.CommentProviderSetting{
				ScriptURL: "https://unpkg.com/artalk@2.9.1/dist/Artalk.js",
				CSSURL:    "https://unpkg.com/artalk@2.9.1/dist/Artalk.css",
				Config: map[string]interface{}{
					"server":  "",
					"site":    "",
					"pageKey": "",
				},
			},
			trimURL: map[string]bool{"server": true},
			check: func(cfg map[string]interface{}) error {
				if strVal(cfg["server"]) == "" {
					return errors.New("artalk.server 不能为空")
				}
				if strVal(cfg["site"]) == "" {
					return errors.New("artalk.site 不能为空")
				}
				return nil
			},
		},
		string(commonModel.GISCUS): {
			label: "Giscus",
			fields: []model.CommentProviderFieldMeta{
				{Key: "repo", Label: "Repo", Required: true, Placeholder: "owner/repo"},
				{Key: "repoId", Label: "Repo ID", Required: true},
				{Key: "category", Label: "Category", Required: true},
				{Key: "categoryId", Label: "Category ID", Required: true},
				{Key: "mapping", Label: "Mapping", Required: false},
				{Key: "strict", Label: "Strict", Required: false},
				{Key: "reactionsEnabled", Label: "Reactions", Required: false},
				{Key: "inputPosition", Label: "Input Position", Required: false},
				{Key: "lang", Label: "Language", Required: false},
				{Key: "theme", Label: "Theme", Required: false},
			},
			def: model.CommentProviderSetting{
				ScriptURL: "https://giscus.app/client.js",
				Config: map[string]interface{}{
					"repo":             "",
					"repoId":           "",
					"category":         "",
					"categoryId":       "",
					"mapping":          "pathname",
					"strict":           "0",
					"reactionsEnabled": "1",
					"inputPosition":    "top",
					"lang":             "zh-CN",
					"theme":            "preferred_color_scheme",
				},
			},
			check: func(cfg map[string]interface{}) error {
				required := []string{"repo", "repoId", "category", "categoryId"}
				for _, key := range required {
					if strVal(cfg[key]) == "" {
						return fmt.Errorf("giscus.%s 不能为空", key)
					}
				}
				return nil
			},
		},
	}
}

func (r commentProviderRegistry) defaultCommentSetting() model.CommentSetting {
	providers := map[string]model.CommentProviderSetting{}
	for key, spec := range r {
		cfgCopy := map[string]interface{}{}
		for k, v := range spec.def.Config {
			cfgCopy[k] = v
		}
		providers[key] = model.CommentProviderSetting{
			ScriptURL: spec.def.ScriptURL,
			CSSURL:    spec.def.CSSURL,
			Config:    cfgCopy,
		}
	}

	return model.CommentSetting{
		EnableComment: false,
		Provider:      string(commonModel.TWIKOO),
		Providers:     providers,
	}
}

func (r commentProviderRegistry) validateProvider(provider string) error {
	if _, ok := r[provider]; !ok {
		return errors.New(commonModel.NO_SUCH_COMMENT_PROVIDER)
	}
	return nil
}

func (r commentProviderRegistry) normalizeAndValidate(setting *model.CommentSettingDto) error {
	if err := r.validateProvider(setting.Provider); err != nil {
		return err
	}
	if setting.Providers == nil {
		setting.Providers = map[string]model.CommentProviderSetting{}
	}

	for provider, spec := range r {
		existing, ok := setting.Providers[provider]
		if !ok {
			setting.Providers[provider] = model.CommentProviderSetting{
				ScriptURL: spec.def.ScriptURL,
				CSSURL:    spec.def.CSSURL,
				Config:    cloneMap(spec.def.Config),
			}
			continue
		}

		if strings.TrimSpace(existing.ScriptURL) == "" {
			existing.ScriptURL = spec.def.ScriptURL
		} else {
			existing.ScriptURL = httpUtil.TrimURL(existing.ScriptURL)
		}
		if strings.TrimSpace(existing.CSSURL) == "" {
			existing.CSSURL = spec.def.CSSURL
		} else {
			existing.CSSURL = httpUtil.TrimURL(existing.CSSURL)
		}
		if existing.Config == nil {
			existing.Config = cloneMap(spec.def.Config)
		}
		for key, val := range spec.def.Config {
			if _, found := existing.Config[key]; !found {
				existing.Config[key] = val
			}
		}
		for urlKey := range spec.trimURL {
			if raw, found := existing.Config[urlKey]; found {
				if value := strVal(raw); value != "" {
					existing.Config[urlKey] = httpUtil.TrimURL(value)
				}
			}
		}
		setting.Providers[provider] = existing
	}

	activeConfig := setting.Providers[setting.Provider]
	if activeConfig.Config == nil {
		activeConfig.Config = map[string]interface{}{}
	}
	if err := r[setting.Provider].check(activeConfig.Config); err != nil {
		return err
	}
	return nil
}

func (r commentProviderRegistry) providerMeta() model.CommentProviderMetaResponse {
	order := []string{
		string(commonModel.TWIKOO),
		string(commonModel.WALINE),
		string(commonModel.ARTALK),
		string(commonModel.GISCUS),
	}
	result := model.CommentProviderMetaResponse{
		Providers: make([]model.CommentProviderMeta, 0, len(order)),
	}
	for _, provider := range order {
		spec, ok := r[provider]
		if !ok {
			continue
		}
		result.Providers = append(result.Providers, model.CommentProviderMeta{
			Provider: provider,
			Label:    spec.label,
			Fields:   spec.fields,
		})
	}
	return result
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func strVal(input interface{}) string {
	str, ok := input.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}
