package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	model "github.com/lin-snow/ech0/internal/model/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type DashboardService struct {
	monitor *monitor.Monitor
}

func NewDashboardService(
	monitor *monitor.Monitor,
) *DashboardService {
	return &DashboardService{
		monitor: monitor,
	}
}

func (dashboardService *DashboardService) GetMetrics() (model.Metrics, error) {
	return dashboardService.monitor.GetMetrics(), nil
}

func (s *DashboardService) WSSubsribeMetrics(w http.ResponseWriter, r *http.Request) error {
	// WebSocket 升级
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 禁用日志输出
		// log.Printf("websocket upgrade failed: %v", err)
		return err
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		defer func() { _ = conn.Close() }()

		for {
			rawMetrics := s.monitor.GetMetrics()
			formatted := fmtUtil.FormatMetrics(&rawMetrics)

			resp := struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
				Data any    `json:"data"`
			}{
				Code: 1,
				Msg:  "metrics update",
				Data: formatted,
			}

			data, _ := json.Marshal(resp)

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				// 禁用日志输出
				// log.Println("write ws error:", err)
				return
			}

			time.Sleep(5 * time.Second)
		}
	}()
	return nil
}

func (s *DashboardService) GetSystemLogs(query SystemLogQuery) ([]logUtil.LogEntry, error) {
	tail := query.Tail
	if tail <= 0 {
		tail = 200
	}
	return logUtil.QueryLogFileTail(logUtil.CurrentLogFilePath(), tail, query.Level, query.Keyword)
}

func (s *DashboardService) WSSubscribeSystemLogs(
	w http.ResponseWriter,
	r *http.Request,
	filter SystemLogStreamFilter,
) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	_, stream, cancel := logUtil.SubscribeLogs(256)
	level := strings.ToLower(strings.TrimSpace(filter.Level))
	keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))

	go func() {
		defer cancel()
		defer func() { _ = conn.Close() }()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, readErr := conn.ReadMessage(); readErr != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-done:
				return
			case entry, ok := <-stream:
				if !ok {
					return
				}
				if !matchesSystemLogFilter(entry, level, keyword) {
					continue
				}
				resp := struct {
					Code int              `json:"code"`
					Msg  string           `json:"msg"`
					Data logUtil.LogEntry `json:"data"`
				}{
					Code: 1,
					Msg:  "system log update",
					Data: entry,
				}
				payload, _ := json.Marshal(resp)
				if writeErr := conn.WriteMessage(websocket.TextMessage, payload); writeErr != nil {
					return
				}
			}
		}
	}()
	return nil
}

func (s *DashboardService) SSESubscribeSystemLogs(
	w http.ResponseWriter,
	r *http.Request,
	filter SystemLogStreamFilter,
) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")

	_, stream, cancel := logUtil.SubscribeLogs(512)
	defer cancel()

	level := strings.ToLower(strings.TrimSpace(filter.Level))
	keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case entry, ok := <-stream:
			if !ok {
				return nil
			}
			if !matchesSystemLogFilter(entry, level, keyword) {
				continue
			}
			resp := struct {
				Code int              `json:"code"`
				Msg  string           `json:"msg"`
				Data logUtil.LogEntry `json:"data"`
			}{
				Code: 1,
				Msg:  "system log update",
				Data: entry,
			}
			payload, _ := json.Marshal(resp)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func matchesSystemLogFilter(entry logUtil.LogEntry, level string, keyword string) bool {
	if level != "" && level != "all" && strings.ToLower(entry.Level) != level {
		return false
	}
	if keyword == "" {
		return true
	}
	target := strings.ToLower(entry.Msg + " " + entry.Raw)
	return strings.Contains(target, keyword)
}
