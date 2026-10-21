package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/Misterio715/go-course-metrics-service/internal/handler"
	"github.com/Misterio715/go-course-metrics-service/internal/repository"
	"github.com/Misterio715/go-course-metrics-service/internal/service"
)

func main() {
	address := flag.String("a", "localhost:8080", "HTTP server address")
	storeInterval := flag.Int("i", 300, "metrics storage interval in seconds (0 saves synchronously)")
	fileStoragePath := flag.String("f", filepath.Join(os.TempDir(), "metrics-db.json"), "metrics storage file path")
	restore := flag.Bool("r", true, "restore metrics from the storage file on startup")
	flag.Parse()

	if envAddress, ok := os.LookupEnv("ADDRESS"); ok {
		*address = envAddress
	}
	if envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		interval, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid STORE_INTERVAL")
		}
		*storeInterval = interval
	}
	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		*fileStoragePath = envFileStoragePath
	}
	if envRestore, ok := os.LookupEnv("RESTORE"); ok {
		value, err := strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid RESTORE")
		}
		*restore = value
	}
	if *storeInterval < 0 || int64(*storeInterval) > int64((1<<63-1)/time.Second) {
		log.Fatal().Msg("Store interval is out of range")
	}

	storage := repository.NewFileStorage(*fileStoragePath, *storeInterval == 0)
	if *restore {
		if err := storage.Restore(); err != nil {
			log.Fatal().Err(err).Str("path", *fileStoragePath).Msg("Failed to restore metrics")
		}
	}

	if *storeInterval > 0 {
		ticker := time.NewTicker(time.Duration(*storeInterval) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				if err := storage.Save(); err != nil {
					log.Error().Err(err).Str("path", *fileStoragePath).Msg("Failed to save metrics")
				}
			}
		}()
	}

	if err := newServerHandler(storage).Run(*address); err != nil {
		log.Info().Err(err).Msg("HTTP server stopped")
	}
}

func newServerHandler(storage service.MetricsRepository) *gin.Engine {
	metricsService := service.NewMetricsService(storage)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	router := gin.New()
	router.Use(handler.LogRequests(&log.Logger))
	router.Use(handler.Gzip())
	router.NoRoute(func(ctx *gin.Context) {
		ctx.Data(http.StatusNotFound, gin.MIMEPlain, []byte("404 page not found"))
	})
	metricsHandler.RegisterRoutes(router)

	return router
}
