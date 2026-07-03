// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lin-snow/ech0/pkg/log/tint"
)

// LevelPanic / LevelFatal 是标准库 slog 没有的两个级别，映射到 Error 之上，
// 仅由 Panic()/Fatal() 内部使用（其余级别与 slog 内置对齐）。
const (
	LevelPanic = slog.LevelError + 4
	LevelFatal = slog.LevelError + 8
)

// fanoutHandler 把每条记录广播给多片叶子 handler，取代旧的 streamCore + zapcore.Tee。
// 叶子通常是：控制台叶（tint 彩色 / JSON）+ 内存环叶（ringHandler，喂后台实时页与异步落盘）。
type fanoutHandler struct {
	leaves []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, leaf := range h.leaves {
		if leaf.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, leaf := range h.leaves {
		if leaf.Enabled(ctx, r.Level) {
			// Clone：一条 Record 分发给多个 handler 时官方建议克隆，避免 attr 背存被并发改写。
			_ = leaf.Handle(ctx, r.Clone())
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	leaves := make([]slog.Handler, len(h.leaves))
	for i, leaf := range h.leaves {
		leaves[i] = leaf.WithAttrs(attrs)
	}
	return &fanoutHandler{leaves: leaves}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	leaves := make([]slog.Handler, len(h.leaves))
	for i, leaf := range h.leaves {
		leaves[i] = leaf.WithGroup(name)
	}
	return &fanoutHandler{leaves: leaves}
}

// newConsoleLeaf 构建控制台叶子：Format==json 用 slog.JSONHandler（prod 结构化 stdout），
// 否则用 tint（dev 彩色 / prod 无色纯文本，由 Color 控制）。文件与内存流不受它影响。
func newConsoleLeaf(config LogConfig, level slog.Leveler) slog.Handler {
	if config.Format == "json" {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: fileReplace,
		})
	}
	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		NoColor:    !config.Color,
		TimeFormat: "15:04:05",
	})
}

// ringHandler 把每条记录直接转成 LogEntry 投递给 LogStreamHub，
// 取代旧 streamCore.Write→buildLogEntry 的「编码成 JSON 再解析回 map」回环。
type ringHandler struct {
	hub    *LogStreamHub
	level  slog.Leveler
	attrs  []slog.Attr
	groups []string
}

func (h *ringHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ringHandler) Handle(_ context.Context, r slog.Record) error {
	if h.hub != nil {
		h.hub.Publish(recordToEntry(r, h.groups, h.attrs))
	}
	return nil
}

func (h *ringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

func (h *ringHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	nh := *h
	nh.groups = append(append([]string{}, h.groups...), name)
	return &nh
}

// recordToEntry 从 slog.Record 直接构建 LogEntry（含扁平化的单行 Raw JSON）。
// module / error(err) 提升为顶层字段，其余进 Fields —— 与旧 zap 行的键格式保持一致。
func recordToEntry(r slog.Record, groups []string, base []slog.Attr) LogEntry {
	e := LogEntry{
		Time:  r.Time.Format(time.RFC3339),
		Level: levelString(r.Level),
		Msg:   r.Message,
	}
	fields := make(map[string]any)
	prefix := ""
	if len(groups) > 0 {
		prefix = strings.Join(groups, ".") + "."
	}
	put := func(a slog.Attr) {
		if a.Equal(slog.Attr{}) {
			return
		}
		if prefix == "" {
			switch a.Key {
			case "module":
				e.Module = a.Value.String()
				return
			case "error", "err":
				if e.Error == "" {
					e.Error = a.Value.String()
				}
				return
			}
		}
		fields[prefix+a.Key] = a.Value.Any()
	}
	for _, a := range base {
		put(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		put(a)
		return true
	})
	if r.PC != 0 {
		e.Caller = shortCaller(r.PC)
	}
	if len(fields) > 0 {
		e.Fields = fields
	}
	e.Raw = compactEntryJSON(e)
	return e
}

// compactEntryJSON 生成与旧 zap 行一致的扁平单行 JSON（字段平铺在顶层，非嵌套）。
// app.log 落盘的正是这一行，QueryLogFileTail / 前端关键字搜索都读它。
func compactEntryJSON(e LogEntry) string {
	m := make(map[string]any, len(e.Fields)+6)
	m["time"] = e.Time
	m["level"] = e.Level
	m["msg"] = e.Msg
	if e.Module != "" {
		m["module"] = e.Module
	}
	if e.Caller != "" {
		m["caller"] = e.Caller
	}
	if e.Error != "" {
		m["error"] = e.Error
	}
	maps.Copy(m, e.Fields)
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// levelString 复刻 zap 的小写级别名（含 panic/fatal），保证 app.log 与前端字段契约不变。
func levelString(l slog.Level) string {
	switch {
	case l >= LevelFatal:
		return "fatal"
	case l >= LevelPanic:
		return "panic"
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	case l >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

// shortCaller 复刻 zapcore.ShortCallerEncoder 的 pkg/file.go:line 形式。
func shortCaller(pc uintptr) string {
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if f.File == "" {
		return ""
	}
	return trimSourcePath(f.File) + ":" + strconv.Itoa(f.Line)
}

// fileReplace 用于 JSON 控制台叶子：把 slog 默认的 UPPERCASE level 转小写、
// 把嵌套的 source 对象拍平成 caller 字符串，对齐 app.log 的键/大小写契约。
func fileReplace(groups []string, a slog.Attr) slog.Attr {
	if len(groups) != 0 {
		return a
	}
	switch a.Key {
	case slog.LevelKey:
		if lvl, ok := a.Value.Any().(slog.Level); ok {
			return slog.String(slog.LevelKey, levelString(lvl))
		}
	case slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
			return slog.String("caller", trimSourcePath(src.File)+":"+strconv.Itoa(src.Line))
		}
	}
	return a
}

// trimSourcePath 把绝对路径裁成最后两段 dir/file.go。
func trimSourcePath(file string) string {
	if file == "" {
		return ""
	}
	if slash := strings.LastIndexByte(file, '/'); slash >= 0 {
		if prev := strings.LastIndexByte(file[:slash], '/'); prev >= 0 {
			return file[prev+1:]
		}
		return file[slash+1:]
	}
	return file
}

// Err 是每个 zap.Error(err) 迁移到 slog 的目标：用 String 而非 Any，
// 否则 slog.JSONHandler 会把 error 值 json.Marshal 成 "{}" 丢掉消息。
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("error", err.Error())
}
