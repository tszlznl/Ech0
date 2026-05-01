// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

type CommonHandler struct {
	commonService service.Service
}

type siteMapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []siteMapURL `xml:"url"`
}

type siteMapURL struct {
	Loc        string `xml:"loc"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func NewCommonHandler(commonService service.Service) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
	}
}

func (commonHandler *CommonHandler) GetHeatMap() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		heatMap, err := commonHandler.commonService.GetHeatMap(timezone)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: heatMap,
			Msg:  commonModel.GET_HEATMAP_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetRss(ctx *gin.Context) {
	atom, err := commonHandler.commonService.GenerateRSS(ctx)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(atom))
}

func (commonHandler *CommonHandler) HelloEch0() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// Embeds versionPkg.Info so version / commit / build_time / license /
		// author / repo_url are all flattened into the JSON top level — the
		// frontend About page reads them directly as the single source of truth.
		hello := struct {
			Hello     string `json:"hello"`
			Copyright string `json:"copyright"`
			versionPkg.Info
		}{
			Hello:     "Hello, Ech0! 👋",
			Copyright: versionPkg.Copyright(),
			Info:      versionPkg.Get(),
		}

		return res.Response{
			Msg:  commonModel.GET_HELLO_SUCCESS,
			Data: hello,
		}
	})
}

func (commonHandler *CommonHandler) Healthz() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Msg: commonModel.GET_HEALTHZ_SUCCESS,
			Data: struct {
				Status  string `json:"status"`
				Version string `json:"version"`
			}{
				Status:  "ok",
				Version: versionPkg.Version,
			},
		}
	})
}

func (commonHandler *CommonHandler) GetWebsiteTitle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.GetWebsiteTitleDto
		if err := ctx.ShouldBindQuery(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}
		title, err := commonHandler.commonService.GetWebsiteTitle(dto.WebSiteURL)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}
		return res.Response{
			Data: title,
			Msg:  commonModel.GET_WEBSITE_TITLE_SUCCESS,
		}
	})
}

func resolveBaseURL(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}

	if forwardedProto := strings.TrimSpace(ctx.GetHeader("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = strings.TrimSpace(strings.Split(forwardedProto, ",")[0])
	}

	host := strings.TrimSpace(ctx.Request.Host)
	if forwardedHost := strings.TrimSpace(ctx.GetHeader("X-Forwarded-Host")); forwardedHost != "" {
		host = strings.TrimSpace(strings.Split(forwardedHost, ",")[0])
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

func (commonHandler *CommonHandler) GetRobotsTxt(ctx *gin.Context) {
	baseURL := resolveBaseURL(ctx)
	content := strings.Join([]string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /api/",
		"Disallow: /auth",
		"Disallow: /panel",
		"Disallow: /init",
		"Disallow: /swagger/",
		"Disallow: /healthz",
		fmt.Sprintf("Sitemap: %s/sitemap.xml", strings.TrimRight(baseURL, "/")),
		"",
	}, "\n")

	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(content))
}

func (commonHandler *CommonHandler) GetSitemap(ctx *gin.Context) {
	baseURL := strings.TrimRight(resolveBaseURL(ctx), "/")
	urlSet := siteMapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []siteMapURL{
			{Loc: baseURL + "/", ChangeFreq: "daily", Priority: "1.0"},
			{Loc: baseURL + "/hub", ChangeFreq: "daily", Priority: "0.8"},
			{Loc: baseURL + "/rss", ChangeFreq: "hourly", Priority: "0.6"},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/xml; charset=utf-8", buf.Bytes())
}
