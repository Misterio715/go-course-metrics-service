package main

import (
	"log"
	"net/http"

	"github.com/Misterio715/go-course-metrics-service/internal/handler"
	"github.com/Misterio715/go-course-metrics-service/internal/repository"
	"github.com/Misterio715/go-course-metrics-service/internal/service"
)

func main() {
	err := http.ListenAndServe(":8080", handler.LogRequests(log.Default(), newServerHandler()))
	if err != nil {
		return
	}
}

func newServerHandler() http.Handler {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	mux := http.NewServeMux()
	metricsHandler.RegisterRoutes(mux)

	return mux
}
