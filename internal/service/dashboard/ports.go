package service

import (
	"net/http"

	model "github.com/lin-snow/ech0/internal/model/metric"
)

type Service interface {
	GetMetrics() (model.Metrics, error)
	WSSubsribeMetrics(w http.ResponseWriter, r *http.Request) error
}
