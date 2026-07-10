package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Misterio715/go-course-metrics-service/internal/handler"
	"github.com/Misterio715/go-course-metrics-service/internal/repository"
	"github.com/Misterio715/go-course-metrics-service/internal/service"
)

func main() {
	if err := newServerHandler().Run(":8080"); err != nil {
		log.Println(err)
	}
}

func newServerHandler() *gin.Engine {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	router := gin.New()
	router.Use(handler.LogRequests(log.Default()))
	metricsHandler.RegisterRoutes(router)

	return router
}
