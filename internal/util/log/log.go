// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	model "github.com/lin-snow/ech0/internal/model/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志记录器
var (
	Logger        *zap.Logger
	loggerMu      sync.Mutex
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

	if Logger != nil {
		_ = Logger.Sync()
		Logger = nil
	}
	stopFileSink()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
	if streamHub != nil {
		streamHub.Close()
		streamHub = nil
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "module",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 初始化内存日志流 Hub
	streamHub = newLogStreamHub(
		safePositive(config.Stream.BufferSize, 2048),
		safePositive(config.Stream.RecentSize, 2000),
		normalizeDropPolicy(config.Stream.DropPolicy),
	)

	var cores []zapcore.Core

	// 控制台输出
	if config.Console {
		consoleConfig := encoderConfig

		var consoleEncoder zapcore.Encoder
		if config.Format == "json" {
			consoleEncoder = zapcore.NewJSONEncoder(consoleConfig)
		} else {
			consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			consoleEncoder = zapcore.NewConsoleEncoder(consoleConfig)
		}

		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出由内存流异步消费者处理，避免阻塞主日志链路
	if config.File.Enable {
		// 确保日志目录存在
		logDir := filepath.Dir(config.File.Filename)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			panic(model.INIT_LOGGER_PANIC + ": 创建日志目录失败: " + err.Error())
		}

		// 配置日志轮转
		writer := &lumberjack.Logger{
			Filename:   config.File.Filename,
			MaxSize:    config.File.MaxSize,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,
			Compress:   config.File.Compress,
			LocalTime:  true,
		}
		fileWriter = writer
		startFileSink(config)
	}

	// 如果没有配置任何输出，使用默认控制台输出
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// 合并所有核心
	core := zapcore.NewTee(cores...)
	core = &streamCore{
		Core:   core,
		level:  level,
		hub:    streamHub,
		parser: zapcore.NewJSONEncoder(encoderConfig),
	}

	// 创建 logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// GetLogger 获取日志记录器实例
func GetLogger() *zap.Logger {
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

	if Logger != nil {
		_ = Logger.Sync()
		Logger = nil
	}
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
func Debug(msg string, fields ...zap.Field) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger.Debug\n")
		}
	}()
	GetLogger().Debug(msg, fields...)
}

// Info 打印信息级别日志
func Info(msg string, fields ...zap.Field) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger.Info\n")
		}
	}()
	GetLogger().Info(msg, fields...)
}

// Warn 打印警告级别日志
func Warn(msg string, fields ...zap.Field) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger.Warn\n")
		}
	}()
	GetLogger().Warn(msg, fields...)
}

// Error 打印错误级别日志
func Error(msg string, fields ...zap.Field) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger.Error\n")
		}
	}()
	GetLogger().Error(msg, fields...)
}

// Panic 打印恐慌级别日志并触发 panic
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

// Fatal 打印致命错误级别日志并终止程序
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
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

type streamCore struct {
	zapcore.Core
	level  zapcore.Level
	hub    *LogStreamHub
	parser zapcore.Encoder
}

func (c *streamCore) Enabled(level zapcore.Level) bool {
	return c.Core.Enabled(level)
}

func (c *streamCore) With(fields []zap.Field) zapcore.Core {
	return &streamCore{
		Core:   c.Core.With(fields),
		level:  c.level,
		hub:    c.hub,
		parser: c.parser.Clone(),
	}
}

func (c *streamCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *streamCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	err := c.Core.Write(entry, fields)
	if c.hub != nil {
		c.hub.Publish(buildLogEntry(c.parser.Clone(), entry, fields))
	}
	return err
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

		for {
			select {
			case <-fileSinkStop:
				flush()
				return
			case <-ticker.C:
				flush()
			case entry, ok := <-stream:
				if !ok {
					flush()
					return
				}
				line := strings.TrimSpace(entry.Raw)
				if line == "" {
					continue
				}
				lines = append(lines, line)
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

func buildLogEntry(enc zapcore.Encoder, entry zapcore.Entry, fields []zap.Field) LogEntry {
	e := LogEntry{
		Time:   entry.Time.Format(time.RFC3339),
		Level:  entry.Level.String(),
		Msg:    entry.Message,
		Module: entry.LoggerName,
		Caller: entry.Caller.TrimmedPath(),
	}
	if len(fields) == 0 {
		obj, _ := json.Marshal(e)
		e.Raw = string(obj)
		return e
	}

	buf, err := enc.EncodeEntry(entry, fields)
	if err == nil && buf != nil {
		line := strings.TrimSpace(buf.String())
		e.Raw = line
		buf.Free()
	}

	var parsed map[string]any
	if e.Raw != "" && json.Unmarshal([]byte(e.Raw), &parsed) == nil {
		e = parseMapAsEntry(parsed, e.Raw)
		return e
	}

	fieldMap := make(map[string]any, len(fields))
	for _, f := range fields {
		fieldMap[f.Key] = "<complex>"
	}
	e.Fields = fieldMap
	obj, _ := json.Marshal(e)
	e.Raw = string(obj)
	return e
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
