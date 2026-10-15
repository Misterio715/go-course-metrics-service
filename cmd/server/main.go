package main

import (
	"net/http"

	"github.com/Misterio715/go-course-metrics-service/internal/handler"
	"github.com/Misterio715/go-course-metrics-service/internal/repository"
	"github.com/Misterio715/go-course-metrics-service/internal/service"
)

func main() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	mux := http.NewServeMux()
	metricsHandler.RegisterRoutes(mux)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}
