package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/Misterio715/go-course-metrics-service/internal/handler"
	"github.com/Misterio715/go-course-metrics-service/internal/repository"
	"github.com/Misterio715/go-course-metrics-service/internal/service"
)

func main() {
	address := flag.String("a", "localhost:8080", "HTTP server address")
	flag.Parse()

	if envAddress, ok := os.LookupEnv("ADDRESS"); ok {
		*address = envAddress
	}

	if err := newServerHandler().Run(*address); err != nil {
		log.Info().Err(err).Msg("HTTP server stopped")
	}
}

func newServerHandler() *gin.Engine {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	router := gin.New()
	router.Use(handler.LogRequests(&log.Logger))
	router.NoRoute(func(ctx *gin.Context) {
		ctx.Data(http.StatusNotFound, gin.MIMEPlain, []byte("404 page not found"))
	})
	metricsHandler.RegisterRoutes(router)

	return router
}
