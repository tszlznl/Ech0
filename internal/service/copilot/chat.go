// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// 检索默认返回的命中条数（固定，不暴露给模型，行为可预测）
const defaultTopK = 6

// chatTemperature 是 Chat 生成温度
var chatTemperature float32 = 0.4

func (s *CopilotService) agentSetting(ctx context.Context) (settingModel.AgentSetting, error) {
	var setting settingModel.AgentSetting
	raw, err := s.kvRepository.GetKeyValue(ctx, commonModel.AgentSettingKey)
	if err != nil {
		return setting, errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return setting, err
	}
	return setting, nil
}

// AskStream 以 Agent（function calling）形态执行一轮问答：模型在一次对话内自主决定
// 是否检索、检索几次（search_echos 工具），全过程以 SSE 写入 w。
//
// 设计上：尽早写出 SSE 头，之后所有错误都以 SSE "error" 事件回传，而非 HTTP 状态码。
// SSE 事件：searching（模型决定检索）/ sources（命中来源，可多次）/ delta（文本增量）/
// done（收尾）/ error（中止）。
func (s *CopilotService) AskStream(ctx context.Context, question string, locale string, w http.ResponseWriter) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	question = strings.TrimSpace(question)
	if question == "" {
		writeSSE(w, flusher, "error", map[string]string{"message": "empty question"})
		return nil
	}

	// 登录用户 ID：用于本轮问答正常收尾后把对话追加进该用户的持久化会话（仅展示恢复，模型不读历史）。
	userID := viewer.MustFromContext(ctx).UserID()
	// 收集本轮 assistant 文本与命中来源，正常收尾时一并持久化。
	var assistantBuf strings.Builder
	var collectedSources []embeddingModel.SearchResult

	agentSetting, err := s.agentSetting(ctx)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	// 一次性取标签：既注入 system prompt 供模型挑选，又供工具把标签名解析成 ID。
	allTags, _ := s.echoService.GetAllTags()
	today := time.Now().UTC().Format("2006-01-02")

	// 多轮记忆：加载已持久化的会话并投影成模型历史（在 persistTurn 之前加载，本轮 question
	// 不在其中，由 buildChatMessages 单独追加，不重复计入）。
	history := historyForModel(s.loadSession(ctx, userID), locale, maxHistoryTokens)

	stream, err := agent.Run(ctx, agent.RunRequest{
		Setting:  agentSetting,
		Messages: buildChatMessages(history, question, locale, today, tagNamesForPrompt(allTags)),
		Tools:    []agent.Tool{s.searchEchosTool(allTags, agentSetting.Multimodal)},
		Temp:     &chatTemperature,
	})
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case ev, ok := <-stream:
			if !ok {
				// Run 正常会在关闭前发 done/error；兜底以 done 收尾
				s.persistTurn(ctx, userID, question, assistantBuf.String(), collectedSources)
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			}
			switch ev.Kind {
			case agent.AgentDelta:
				if ev.Text != "" {
					assistantBuf.WriteString(ev.Text)
					writeSSE(w, flusher, "delta", map[string]string{"text": ev.Text})
				}
			case agent.AgentSearching:
				writeSSE(w, flusher, "searching", map[string]string{
					"name":  ev.ToolName,
					"query": searchHintOf(ev.ToolArgs),
				})
			case agent.AgentToolResult:
				if src, ok := ev.Meta.([]embeddingModel.SearchResult); ok {
					collectedSources = append(collectedSources, src...)
				}
				writeSSE(w, flusher, "sources", ev.Meta)
			case agent.AgentDone:
				s.persistTurn(ctx, userID, question, assistantBuf.String(), collectedSources)
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			case agent.AgentError:
				writeSSE(w, flusher, "error", map[string]string{"message": ev.Err.Error()})
				return nil
			}
		}
	}
}

// searchArgs 是 search_echos 工具的入参。三者皆可选，但至少需其一：
// query 走语义/关键词，tags/date_* 走结构化精确过滤。
type searchArgs struct {
	Query    string   `json:"query"`
	Tags     []string `json:"tags"`
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
}

// searchEchosTool 是注入给 agent 的领域工具：检索用户过往 Echo。
//
// 检索路由（一条规则）：只要带结构化过滤（tags / date_*）就走 QueryEchos（SQL 精确，
// 向量索引做不了元数据过滤）；纯 query 且向量已启用才走 embedding 语义检索，否则回退
// QueryEchos 的 content LIKE。allTags 用于把模型给的标签名解析成 ID（UUID 不进 prompt）。
func (s *CopilotService) searchEchosTool(allTags []echoModel.Tag, multimodal bool) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "search_echos",
			Description: "检索用户过往发布的 Echo（微博客/碎碎念）。可用 query 做语义/关键词检索，并可选地用 tags（标签名）与 date_from/date_to（日期范围）做精确筛选；三者可组合，但至少提供其一。query 传精炼核心词，不要整句。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"语义/关键词检索词，传与问题最相关的核心词（精炼，不要整句）；仅按标签或时间筛选时可省略"},"tags":{"type":"array","items":{"type":"string"},"description":"按标签名筛选（标签名而非ID），如 [\"读书\",\"旅行\"]；可用标签见系统提示"},"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"}}}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a searchArgs
			_ = json.Unmarshal(args, &a)
			a.Query = strings.TrimSpace(a.Query)
			tagIDs := resolveTagIDs(allTags, a.Tags)
			from := parseDay(a.DateFrom, false)
			to := parseDay(a.DateTo, true)

			structured := len(tagIDs) > 0 || from > 0 || to > 0
			if a.Query == "" && !structured {
				return agent.ToolOutput{}, errors.New("检索需要 query、tags 或日期范围至少其一")
			}

			var results []embeddingModel.SearchResult
			var execErr error
			switch {
			case structured:
				// 带结构化过滤：SQL 精确路径，query 降级为 content LIKE。
				results, execErr = s.queryEchos(ctx, a.Query, tagIDs, from, to)
			case s.embedding.Enabled(ctx):
				results, execErr = s.embedding.Search(ctx, a.Query, defaultTopK)
			default:
				results, execErr = s.queryEchos(ctx, a.Query, nil, 0, 0)
			}
			if execErr != nil {
				return agent.ToolOutput{}, execErr
			}
			// 命中后回查：Extension 文本（常开，仅几字 token）+ 配图（多模态开关）一次加载取齐。
			exts, images := s.enrichHits(ctx, results, multimodal)
			return agent.ToolOutput{
				Content: formatSearchResults(results, exts),
				Meta:    results,
				Images:  images,
			}, nil
		},
	}
}

// maxChatImages 是单轮注入模型的图片数上限（控制 payload 与成本）。
const maxChatImages = 4

// maxImageBytes 是单张图注入的字节上限；超过则跳过（避免超大图撑爆请求）。
const maxImageBytes = 5 << 20 // 5MB

// enrichHits 按命中顺序（最相关在前）回查 Echo（GetEchoById 带缓存），一次加载取齐三样：
//   - exts：每条命中的 Extension 渲染文本（音乐/网站/位置等分享，常开，喂给模型理解）；
//   - results[i].Files：命中 Echo 的图片附件元数据（常开，随 SSE 给前端展示缩略图，仅 URL 不含字节）；
//   - images：配图的 base64 ImagePart，仅 multimodal 开启时收集，累计到 maxChatImages 即止（喂模型）。
//
// 不存进 embedding 索引、只在检索命中后回查，向量库保持纯文本干净。读取失败静默跳过（best-effort）。
func (s *CopilotService) enrichHits(
	ctx context.Context,
	results []embeddingModel.SearchResult,
	multimodal bool,
) (map[string]string, []agent.ImagePart) {
	exts := make(map[string]string, len(results))
	var images []agent.ImagePart
	for i := range results {
		echo, err := s.echoService.GetEchoById(ctx, results[i].EchoID)
		if err != nil || echo == nil {
			continue
		}
		if txt := formatExtension(echo.Extension); txt != "" {
			exts[results[i].EchoID] = txt
		}
		results[i].Extension = echo.Extension // 前端展示用：扩展类型标签（音乐/网站/位置…）

		var files []fileModel.File
		for _, ef := range echo.EchoFiles {
			if !storage.NormalizeCategory(ef.File.Category).IsImageLike() {
				continue
			}
			files = append(files, ef.File) // 前端展示用：整条 File（含 storage_type/key/url 等）
			// 多模态：再把图片字节读成 base64 喂给模型（受 maxChatImages 上限约束）。
			if multimodal && s.storage != nil && len(images) < maxChatImages {
				if part, ok := s.loadImagePart(ctx, ef.File); ok {
					images = append(images, part)
				}
			}
		}
		results[i].Files = files
	}
	return exts, images
}

// formatExtension 把 Echo 的扩展分享渲染成一行供模型理解的文本（无扩展或缺字段返回空）。
// Payload 由 GORM json serializer 反序列化为 map，字符串值原样取出。
func formatExtension(ext *echoModel.EchoExtension) string {
	if ext == nil {
		return ""
	}
	str := func(k string) string {
		if v, ok := ext.Payload[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	switch ext.Type {
	case echoModel.Extension_MUSIC:
		if u := str("url"); u != "" {
			return "[音乐分享] " + u
		}
	case echoModel.Extension_VIDEO:
		if id := str("videoId"); id != "" {
			return "[视频分享] 视频ID " + id
		}
	case echoModel.Extension_GITHUBPROJ:
		if u := str("repoUrl"); u != "" {
			return "[GitHub 项目] " + u
		}
	case echoModel.Extension_WEBSITE:
		title, site := str("title"), str("site")
		switch {
		case title != "" && site != "":
			return "[网站] " + title + " " + site
		case site != "":
			return "[网站] " + site
		case title != "":
			return "[网站] " + title
		}
	case echoModel.Extension_LOCATION:
		if place := str("placeholder"); place != "" {
			return "[位置] " + place
		}
	case echoModel.Extension_TWEET:
		u, user := str("url"), str("username")
		switch {
		case u != "" && user != "":
			return "[X 推文] @" + user + " " + u
		case u != "":
			return "[X 推文] " + u
		}
	}
	return ""
}

// loadImagePart 把单个 File 读成 ImagePart：external 直接用公网直链；local/object 读字节做 base64。
// 非图片、超限、读失败均返回 ok=false 由调用方跳过。
func (s *CopilotService) loadImagePart(ctx context.Context, f fileModel.File) (agent.ImagePart, bool) {
	if !storage.NormalizeCategory(f.Category).IsImageLike() {
		return agent.ImagePart{}, false
	}
	mediaType := f.ContentType
	if mediaType == "" {
		mediaType = "image/jpeg"
	}

	st := storage.NormalizeStorageType(f.StorageType)
	if st == storage.StorageTypeExternal {
		if f.URL == "" {
			return agent.ImagePart{}, false
		}
		return agent.ImagePart{MediaType: mediaType, URL: f.URL}, true
	}

	if f.Size > maxImageBytes {
		return agent.ImagePart{}, false
	}
	reader, err := s.storage.GetSelector().Get(ctx, st, f.Key)
	if err != nil {
		return agent.ImagePart{}, false
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(io.LimitReader(reader, maxImageBytes))
	if err != nil || len(data) == 0 {
		return agent.ImagePart{}, false
	}
	return agent.ImagePart{MediaType: mediaType, Base64: base64.StdEncoding.EncodeToString(data)}, true
}

// queryEchos 走 echoService.QueryEchos（SQL 检索 + 可见性由 viewer 上下文裁决：
// /chat 为 admin，故 showPrivate=true，与向量索引可见性一致）。结果映射成与向量检索
// 同一形状，上层（formatSearchResults / SSE sources）无需区分两条路径。
func (s *CopilotService) queryEchos(ctx context.Context, search string, tagIDs []string, from, to int64) ([]embeddingModel.SearchResult, error) {
	page, err := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     1,
		PageSize: defaultTopK,
		Search:   search,
		TagIDs:   tagIDs,
		DateFrom: from,
		DateTo:   to,
	})
	if err != nil {
		return nil, err
	}
	results := make([]embeddingModel.SearchResult, 0, len(page.Items))
	for _, e := range page.Items {
		results = append(results, embeddingModel.SearchResult{
			EchoID:      e.ID,
			Content:     e.Content,
			Username:    e.Username,
			EchoCreated: e.CreatedAt,
			Distance:    0,
		})
	}
	return results, nil
}

// resolveTagIDs 把模型给的标签名（大小写不敏感）解析成标签 ID；匹配不上的名静默忽略。
func resolveTagIDs(allTags []echoModel.Tag, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	byName := make(map[string]string, len(allTags))
	for _, t := range allTags {
		byName[strings.ToLower(t.Name)] = t.ID
	}
	ids := make([]string, 0, len(names))
	for _, n := range names {
		if id, ok := byName[strings.ToLower(strings.TrimSpace(n))]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// parseDay 把 YYYY-MM-DD 解析成 Unix 秒（UTC）；endOfDay 为真时取当天 23:59:59，
// 用于把闭区间的右端覆盖到整天。解析失败或空串返回 0（视为未设置）。
func parseDay(s string, endOfDay bool) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return 0
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t.Unix()
}

// searchHintOf 从工具入参拼一条人读的检索提示（供 SSE searching 事件展示），
// 组合 query / #标签 / 日期范围。
func searchHintOf(args json.RawMessage) string {
	var a searchArgs
	_ = json.Unmarshal(args, &a)
	parts := make([]string, 0, 3)
	if q := strings.TrimSpace(a.Query); q != "" {
		parts = append(parts, q)
	}
	for _, t := range a.Tags {
		if t = strings.TrimSpace(t); t != "" {
			parts = append(parts, "#"+t)
		}
	}
	if from, to := strings.TrimSpace(a.DateFrom), strings.TrimSpace(a.DateTo); from != "" || to != "" {
		parts = append(parts, from+"~"+to)
	}
	return strings.Join(parts, " ")
}

// formatSearchResults 把检索命中拼成回喂模型的精简文本（文本快照 + 扩展分享，控制 token）。
// exts 是 echoID → Extension 渲染文本（来自 enrichHits），命中时补在内容之后。
func formatSearchResults(results []embeddingModel.SearchResult, exts map[string]string) string {
	if len(results) == 0 {
		return "（没有检索到相关的 Echo）"
	}
	var b strings.Builder
	for i, r := range results {
		day := time.Unix(r.EchoCreated, 0).UTC().Format("2006-01-02")
		parts := []string{fmt.Sprintf("【%d】(%s)", i+1, day)}
		if c := strings.TrimSpace(r.Content); c != "" {
			parts = append(parts, c)
		}
		if ext := exts[r.EchoID]; ext != "" {
			parts = append(parts, ext)
		}
		b.WriteString(strings.Join(parts, " "))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}
