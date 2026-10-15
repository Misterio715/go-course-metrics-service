package handler

import (
	"fmt"
	"html"
	"net/http"
	"sort"
	"strconv"
	"strings"

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
	router.POST("/update/:type/:metric/:value", h.update)
	router.GET("/value/:type/:metric", h.value)
	router.GET("/", h.metrics)
}

func (h *MetricsHandler) update(c *gin.Context) {
	name := c.Param("metric")
	value := c.Param("value")

	switch c.Param("type") {
	case "gauge":
		gauge, err := strconv.ParseFloat(value, 64)
		if err != nil {
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
