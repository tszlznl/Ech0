package service

import (
	"net/http"

	model "github.com/lin-snow/ech0/internal/model/metric"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type SystemLogQuery struct {
	Tail    int
	Level   string
	Keyword string
}

type SystemLogStreamFilter struct {
	Level   string
	Keyword string
}

type Service interface {
	GetMetrics() (model.Metrics, error)
	WSSubsribeMetrics(w http.ResponseWriter, r *http.Request) error
	GetSystemLogs(query SystemLogQuery) ([]logUtil.LogEntry, error)
	WSSubscribeSystemLogs(w http.ResponseWriter, r *http.Request, filter SystemLogStreamFilter) error
	SSESubscribeSystemLogs(w http.ResponseWriter, r *http.Request, filter SystemLogStreamFilter) error
}
