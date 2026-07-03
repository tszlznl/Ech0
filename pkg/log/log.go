// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// initLoggerPanic 是初始化 logger 失败时的 panic 前缀。内联此常量以避免 pkg/ 反向依赖 internal/。
const initLoggerPanic = "初始化 Logger 失败"

// 全局日志记录器（slog）与其协作单例。
var (
	Logger        *slog.Logger
	loggerMu      sync.Mutex
	levelVar      = new(slog.LevelVar)
	fileWriter    *lumberjack.Logger
	currentConfig LogConfig
	streamHub     *LogStreamHub
	fileSinkStop  chan struct{}
	fileSinkDone  chan struct{}
	fileSinkID    int64
)

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别: debug, info, warn, error, panic
	Level string `yaml:"level"   json:"level"`
	// 日志格式: json, console
	Format string `yaml:"format"  json:"format"`
	// 是否输出到控制台
	Console bool `yaml:"console" json:"console"`
	// Color 控制控制台叶子是否彩色（dev 开 / prod 关）；仅作用于控制台，不影响文件与内存流。
	Color bool `yaml:"-"       json:"-"`
	// 文件输出配置
	File FileConfig `yaml:"file"    json:"file"`
	// 内存流式日志配置
	Stream StreamConfig `yaml:"stream"  json:"stream"`
}

// FileConfig 文件输出配置
type FileConfig struct {
	// 是否启用文件输出
	Enable bool `yaml:"enable"     json:"enable"`
	// 日志文件路径
	Filename string `yaml:"filename"   json:"filename"`
	// 单个文件最大大小（MB）
	MaxSize int `yaml:"maxsize"    json:"maxsize"`
	// 保留的旧文件数量
	MaxBackups int `yaml:"maxbackups" json:"maxbackups"`
	// 保留天数
	MaxAge int `yaml:"maxage"     json:"maxage"`
	// 是否压缩旧文件
	Compress bool `yaml:"compress"   json:"compress"`
}

// StreamConfig 内存流式日志配置
type StreamConfig struct {
	// 内存缓冲区大小
	BufferSize int `yaml:"buffer_size" json:"buffer_size"`
	// 最近日志保留数量（用于实时页面/内存回看）
	RecentSize int `yaml:"recent_size" json:"recent_size"`
	// 缓冲区溢出策略: drop_oldest / drop_newest
	DropPolicy string `yaml:"drop_policy" json:"drop_policy"`
	// 异步落盘批量大小
	FlushBatch int `yaml:"flush_batch" json:"flush_batch"`
	// 异步落盘刷新间隔（毫秒）
	FlushIntervalMs int `yaml:"flush_interval_ms" json:"flush_interval_ms"`
}

// LogEntry 标准化日志条目
type LogEntry struct {
	Time   string         `json:"time"`
	Level  string         `json:"level"`
	Msg    string         `json:"msg"`
	Module string         `json:"module,omitempty"`
	Caller string         `json:"caller,omitempty"`
	Error  string         `json:"error,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
	Raw    string         `json:"raw,omitempty"`
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:   "info",
		Format:  "json",
		Console: false,
		File: FileConfig{
			Enable:     true,
			Filename:   "data/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
		Stream: StreamConfig{
			BufferSize:      2048,
			RecentSize:      2000,
			DropPolicy:      "drop_oldest",
			FlushBatch:      128,
			FlushIntervalMs: 500,
		},
	}
}

// InitLogger 使用默认配置初始化日志记录器
func InitLogger() {
	InitLoggerWithConfig(DefaultLogConfig())
}

// InitLoggerWithConfig 使用自定义配置初始化日志记录器
func InitLoggerWithConfig(config LogConfig) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	initializeLogger(config)
}

func initializeLogger(config LogConfig) {
	currentConfig = config

	Logger = nil
	stopFileSink()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
	if streamHub != nil {
		streamHub.Close()
		streamHub = nil
	}

	// 解析日志级别（无效时回退 info）。用 LevelVar 让各叶子共享同一个动态级别。
	levelVar.Set(parseLevel(config.Level))

	// 初始化内存日志流 Hub
	streamHub = newLogStreamHub(
		safePositive(config.Stream.BufferSize, 2048),
		safePositive(config.Stream.RecentSize, 2000),
		normalizeDropPolicy(config.Stream.DropPolicy),
	)

	// 文件输出：仍由内存流的异步消费者写入（保持异步，避免把 fsync/轮转压到日志热路径）。
	if config.File.Enable {
		logDir := filepath.Dir(config.File.Filename)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			panic(initLoggerPanic + ": 创建日志目录失败: " + err.Error())
		}
		fileWriter = &lumberjack.Logger{
			Filename:   config.File.Filename,
			MaxSize:    config.File.MaxSize,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,
			Compress:   config.File.Compress,
			LocalTime:  true,
		}
		startFileSink(config)
	}

	// 组装叶子 handler：控制台叶（始终存在，保证 stdout 有输出，对齐旧兜底行为）+ 内存环叶。
	leaves := []slog.Handler{
		newConsoleLeaf(config, levelVar),
		&ringHandler{hub: streamHub, level: levelVar},
	}

	Logger = slog.New(&fanoutHandler{leaves: leaves})
}

// parseLevel 把配置字符串解析为 slog.Level（含 panic/fatal 自定义级别）。
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "panic":
		return LevelPanic
	case "fatal":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

// GetLogger 获取日志记录器实例
func GetLogger() *slog.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if Logger == nil {
		cfg := currentConfig
		if cfg == (LogConfig{}) {
			cfg = DefaultLogConfig()
		}
		initializeLogger(cfg)
	}
	return Logger
}

// CloseLogger 关闭日志记录器，释放资源
func CloseLogger() {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	Logger = nil
	stopFileSink()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
	if streamHub != nil {
		streamHub.Close()
		streamHub = nil
	}
}

// ReopenLogger 使用最近的配置重新初始化日志
func ReopenLogger() {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if Logger != nil {
		return
	}

	cfg := currentConfig
	if cfg == (LogConfig{}) {
		cfg = DefaultLogConfig()
	}
	initializeLogger(cfg)
}

// Debug 打印调试级别日志
//
//go:noinline
func Debug(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelDebug, msg, attrs)
}

// Info 打印信息级别日志
//
//go:noinline
func Info(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelInfo, msg, attrs)
}

// Warn 打印警告级别日志
//
//go:noinline
func Warn(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelWarn, msg, attrs)
}

// Error 打印错误级别日志
//
//go:noinline
func Error(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelError, msg, attrs)
}

// Panic 打印恐慌级别日志并触发 panic
//
//go:noinline
func Panic(msg string, attrs ...slog.Attr) {
	logWithStack(LevelPanic, msg, attrs)
	panic(msg)
}

// Fatal 打印致命错误级别日志并终止程序
//
//go:noinline
func Fatal(msg string, attrs ...slog.Attr) {
	logWithStack(LevelFatal, msg, attrs)
	// 退出前同步排空异步落盘，保证最后一行落到 app.log。
	loggerMu.Lock()
	stopFileSink()
	loggerMu.Unlock()
	os.Exit(1)
}

// logWithAttrs 手动捕获真实调用点的 PC 再交给 Handler，
// 这样通过包级函数打的日志 caller 指向业务代码，而非本文件。
//
//go:noinline
func logWithAttrs(level slog.Level, msg string, attrs []slog.Attr) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger\n")
		}
	}()

	logger := GetLogger()
	ctx := context.Background()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // 跳过 Callers / logWithAttrs / 包级封装 → 业务调用点
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = logger.Handler().Handle(ctx, r)
}

// logWithStack 与 logWithAttrs 相同，但额外附带调用栈（供 Panic/Fatal 使用）。
//
//go:noinline
func logWithStack(level slog.Level, msg string, attrs []slog.Attr) {
	logger := GetLogger()
	ctx := context.Background()
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	r.AddAttrs(slog.String("stacktrace", string(debug.Stack())))
	_ = logger.Handler().Handle(ctx, r)
}

// SubscribeLogs 订阅实时日志流。
func SubscribeLogs(bufferSize int) (int64, <-chan LogEntry, func()) {
	loggerMu.Lock()
	if streamHub == nil {
		cfg := currentConfig
		if cfg == (LogConfig{}) {
			cfg = DefaultLogConfig()
		}
		initializeLogger(cfg)
	}
	hub := streamHub
	loggerMu.Unlock()
	return hub.Subscribe(bufferSize)
}

// RecentLogs 返回最近日志（内存窗口）。
func RecentLogs(limit int) []LogEntry {
	loggerMu.Lock()
	hub := streamHub
	loggerMu.Unlock()
	if hub == nil {
		return nil
	}
	return hub.Recent(limit)
}

// CurrentLogFilePath 返回当前日志文件路径。
func CurrentLogFilePath() string {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if currentConfig.File.Filename == "" {
		return DefaultLogConfig().File.Filename
	}
	return currentConfig.File.Filename
}

// QueryLogFileTail 查询 app.log 历史日志（JSON 行）。
func QueryLogFileTail(path string, limit int, level, keyword string) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 5000 {
		limit = 5000
	}
	level = strings.ToLower(strings.TrimSpace(level))
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []LogEntry{}, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	buffer := make([]LogEntry, 0, limit)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entry := parseLogLine(line)
		if !matchLogFilters(entry, level, keyword) {
			continue
		}
		if len(buffer) < limit {
			buffer = append(buffer, entry)
			continue
		}
		copy(buffer, buffer[1:])
		buffer[len(buffer)-1] = entry
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return buffer, nil
}

type LogStreamHub struct {
	mu               sync.RWMutex
	subs             map[int64]chan LogEntry
	nextID           int64
	recent           []LogEntry
	recentCap        int
	recentPos        int
	recentLen        int
	dropPolicy       string
	dropped          atomic.Uint64
	closed           bool
	defaultSubBuffer int
}

func newLogStreamHub(subBufferSize, recentSize int, dropPolicy string) *LogStreamHub {
	if subBufferSize <= 0 {
		subBufferSize = 2048
	}
	if recentSize <= 0 {
		recentSize = 2000
	}
	return &LogStreamHub{
		subs:             make(map[int64]chan LogEntry),
		recent:           make([]LogEntry, recentSize),
		recentCap:        recentSize,
		dropPolicy:       dropPolicy,
		defaultSubBuffer: subBufferSize,
	}
}

func (h *LogStreamHub) Subscribe(bufferSize int) (int64, <-chan LogEntry, func()) {
	if bufferSize <= 0 {
		bufferSize = h.defaultSubBuffer
		if bufferSize <= 0 {
			bufferSize = 256
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		ch := make(chan LogEntry)
		close(ch)
		return 0, ch, func() {}
	}
	h.nextID++
	id := h.nextID
	ch := make(chan LogEntry, bufferSize)
	h.subs[id] = ch
	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		c, ok := h.subs[id]
		if !ok {
			return
		}
		delete(h.subs, id)
		close(c)
	}
	return id, ch, cancel
}

func (h *LogStreamHub) Publish(entry LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.recent[h.recentPos] = entry
	h.recentPos = (h.recentPos + 1) % h.recentCap
	if h.recentLen < h.recentCap {
		h.recentLen++
	}
	for _, ch := range h.subs {
		select {
		case ch <- entry:
		default:
			if h.dropPolicy == "drop_oldest" {
				select {
				case <-ch:
				default:
				}
				select {
				case ch <- entry:
				default:
					h.dropped.Add(1)
				}
				continue
			}
			h.dropped.Add(1)
		}
	}
}

func (h *LogStreamHub) Recent(limit int) []LogEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.recentLen == 0 {
		return nil
	}
	if limit <= 0 || limit > h.recentLen {
		limit = h.recentLen
	}
	out := make([]LogEntry, 0, limit)
	start := h.recentPos - limit
	if start < 0 {
		start += h.recentCap
	}
	for i := 0; i < limit; i++ {
		idx := (start + i) % h.recentCap
		out = append(out, h.recent[idx])
	}
	return out
}

func (h *LogStreamHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, ch := range h.subs {
		close(ch)
		delete(h.subs, id)
	}
}

func startFileSink(config LogConfig) {
	if streamHub == nil || fileWriter == nil {
		return
	}
	fileSinkStop = make(chan struct{})
	fileSinkDone = make(chan struct{})
	bufferSize := safePositive(config.Stream.BufferSize, 2048)
	var stream <-chan LogEntry
	fileSinkID, stream, _ = streamHub.Subscribe(bufferSize)
	flushBatch := safePositive(config.Stream.FlushBatch, 128)
	flushInterval := time.Duration(safePositive(config.Stream.FlushIntervalMs, 500)) * time.Millisecond
	go func() {
		defer close(fileSinkDone)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		lines := make([]string, 0, flushBatch)
		flush := func() {
			if len(lines) == 0 || fileWriter == nil {
				return
			}
			payload := strings.Join(lines, "\n") + "\n"
			_, _ = fileWriter.Write([]byte(payload))
			lines = lines[:0]
		}
		appendLine := func(entry LogEntry) {
			if line := strings.TrimSpace(entry.Raw); line != "" {
				lines = append(lines, line)
			}
		}

		for {
			select {
			case <-fileSinkStop:
				// 停止前排空剩余缓冲，保证已投递的行不丢（Fatal 依赖此保证）。
				for {
					select {
					case entry, ok := <-stream:
						if ok {
							appendLine(entry)
							continue
						}
						flush()
						return
					default:
						flush()
						return
					}
				}
			case <-ticker.C:
				flush()
			case entry, ok := <-stream:
				if !ok {
					flush()
					return
				}
				appendLine(entry)
				if len(lines) >= flushBatch {
					flush()
				}
			}
		}
	}()
}

func stopFileSink() {
	if fileSinkStop != nil {
		close(fileSinkStop)
		fileSinkStop = nil
	}
	if streamHub != nil && fileSinkID > 0 {
		streamHub.mu.Lock()
		if ch, ok := streamHub.subs[fileSinkID]; ok {
			delete(streamHub.subs, fileSinkID)
			close(ch)
		}
		streamHub.mu.Unlock()
		fileSinkID = 0
	}
	if fileSinkDone != nil {
		<-fileSinkDone
		fileSinkDone = nil
	}
}

func safePositive(v int, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func normalizeDropPolicy(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "drop_newest":
		return "drop_newest"
	default:
		return "drop_oldest"
	}
}

func parseLogLine(line string) LogEntry {
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return LogEntry{
			Level: "info",
			Msg:   line,
			Raw:   line,
		}
	}
	return parseMapAsEntry(payload, line)
}

func parseMapAsEntry(payload map[string]any, raw string) LogEntry {
	entry := LogEntry{
		Time:   toString(payload["time"]),
		Level:  strings.ToLower(toString(payload["level"])),
		Msg:    toString(payload["msg"]),
		Module: toString(payload["module"]),
		Caller: toString(payload["caller"]),
		Error:  toString(payload["error"]),
		Raw:    raw,
	}
	fields := make(map[string]any)
	for k, v := range payload {
		switch k {
		case "time", "level", "msg", "module", "caller", "error":
			continue
		case "source":
			// 防御：万一有原生 slog 行漏进来（source 是嵌套对象），拍平成 caller，不糊进 Fields。
			if entry.Caller == "" {
				entry.Caller = sourceMapToCaller(v)
			}
			continue
		default:
			fields[k] = v
			if k == "err" && entry.Error == "" {
				entry.Error = toString(v)
			}
		}
	}
	if len(fields) > 0 {
		entry.Fields = fields
	}
	if entry.Msg == "" {
		entry.Msg = raw
	}
	return entry
}

// sourceMapToCaller 把 slog 的嵌套 source 对象 {file,line,function} 拍平成 pkg/file.go:line。
func sourceMapToCaller(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	file := toString(m["file"])
	if file == "" {
		return ""
	}
	caller := trimSourcePath(file)
	if line := toString(m["line"]); line != "" {
		caller += ":" + line
	}
	return caller
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch tv := v.(type) {
	case string:
		return tv
	default:
		b, err := json.Marshal(tv)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func matchLogFilters(entry LogEntry, level, keyword string) bool {
	if level != "" && level != "all" && strings.ToLower(entry.Level) != level {
		return false
	}
	if keyword != "" {
		raw := strings.ToLower(entry.Raw)
		msg := strings.ToLower(entry.Msg)
		if !strings.Contains(raw, keyword) && !strings.Contains(msg, keyword) {
			return false
		}
	}
	return true
}
