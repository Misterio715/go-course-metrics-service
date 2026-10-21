package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
	"github.com/gin-gonic/gin"
)

type MetricsService interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetMetrics() (map[string]float64, map[string]int64)
}

type MetricsHandler struct {
	service MetricsService
}

func NewMetricsHandler(service MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

func (h *MetricsHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/update", h.updateJSON)
	router.POST("/value", h.valueJSON)
	router.POST("/update/:type/:metric/:value", h.update)
	router.GET("/value/:type/:metric", h.value)
	router.GET("/", h.metrics)
}

func (h *MetricsHandler) updateJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var metric models.Metrics
	if err := json.NewDecoder(c.Request.Body).Decode(&metric); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		metric.Delta = nil
		h.service.AddGauge(metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		metric.Value = nil
		h.service.AddCounter(metric.ID, *metric.Delta)
		value, ok := h.service.GetCounter(metric.ID)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		metric.Delta = &value
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(metric)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", body)
}

func (h *MetricsHandler) valueJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var metric models.Metrics
	if err := json.NewDecoder(c.Request.Body).Decode(&metric); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.Gauge:
		value, ok := h.service.GetGauge(metric.ID)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		metric.Delta = nil
		metric.Value = &value
	case models.Counter:
		value, ok := h.service.GetCounter(metric.ID)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		metric.Delta = &value
		metric.Value = nil
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(metric)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", body)
}

func (h *MetricsHandler) update(c *gin.Context) {
	name := c.Param("metric")
	value := c.Param("value")

	switch c.Param("type") {
	case "gauge":
		gauge, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(gauge) || math.IsInf(gauge, 0) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		h.service.AddGauge(name, gauge)
		c.Status(http.StatusOK)
	case "counter":
		counter, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		h.service.AddCounter(name, counter)
		c.Status(http.StatusOK)
	default:
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

func (h *MetricsHandler) value(c *gin.Context) {
	name := c.Param("metric")

	switch c.Param("type") {
	case "gauge":
		value, ok := h.service.GetGauge(name)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, "%v", value)
	case "counter":
		value, ok := h.service.GetCounter(name)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, "%v", value)
	default:
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func (h *MetricsHandler) metrics(c *gin.Context) {
	gauges, counters := h.service.GetMetrics()

	var page strings.Builder
	page.WriteString("<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>")

	gaugeNames := make([]string, 0, len(gauges))
	for name := range gauges {
		gaugeNames = append(gaugeNames, name)
	}
	sort.Strings(gaugeNames)
	for _, name := range gaugeNames {
		fmt.Fprintf(&page, "<li>gauge %s: %v</li>", html.EscapeString(name), gauges[name])
	}

	counterNames := make([]string, 0, len(counters))
	for name := range counters {
		counterNames = append(counterNames, name)
	}
	sort.Strings(counterNames)
	for _, name := range counterNames {
		fmt.Fprintf(&page, "<li>counter %s: %v</li>", html.EscapeString(name), counters[name])
	}

	page.WriteString("</ul></body></html>")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page.String()))
}
