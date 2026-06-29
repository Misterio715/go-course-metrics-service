package handler

import (
	"net/http"
	"strconv"
)

type MetricsService interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
}

type MetricsHandler struct {
	service MetricsService
}

func NewMetricsHandler(service MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

func (h *MetricsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/update/gauge/{metric}/{value}", HandlerFunc(h.gauge))
	mux.Handle("/update/gauge/", HandlerFunc(h.gauge))
	mux.Handle("/update/counter/{metric}/{value}", HandlerFunc(h.counter))
	mux.Handle("/update/counter/", HandlerFunc(h.counter))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}

func (h *MetricsHandler) gauge(w http.ResponseWriter, r *http.Request) error {
	metric, value, err := parseGauge(r)
	if err != nil {
		return err
	}

	h.service.AddGauge(metric, value)
	return nil
}

func (h *MetricsHandler) counter(w http.ResponseWriter, r *http.Request) error {
	metric, value, err := parseCounter(r)
	if err != nil {
		return err
	}

	h.service.AddCounter(metric, value)
	return nil
}

func parseGauge(r *http.Request) (string, float64, error) {
	name := r.PathValue("metric")
	if name == "" {
		return "", 0, HTTPError{Code: http.StatusNotFound, Msg: "missing metric"}
	}

	value := r.PathValue("value")
	if value == "" {
		return "", 0, HTTPError{Code: http.StatusBadRequest, Msg: "missing value"}
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", 0, HTTPError{Code: http.StatusBadRequest, Msg: "invalid value", Err: err}
	}

	return name, floatValue, nil
}

func parseCounter(r *http.Request) (string, int64, error) {
	name := r.PathValue("metric")
	if name == "" {
		return "", 0, HTTPError{Code: http.StatusNotFound, Msg: "missing metric"}
	}

	value := r.PathValue("value")
	if value == "" {
		return "", 0, HTTPError{Code: http.StatusBadRequest, Msg: "missing value"}
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", 0, HTTPError{Code: http.StatusBadRequest, Msg: "invalid value", Err: err}
	}

	return name, intValue, nil
}
