package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"resty.dev/v3"
)

func main() {
	address := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")
	flag.Parse()

	baseURL := "http://" + *address + "/update/"
	collector := newMetricsCollector()

	client := resty.New().SetTimeout(10 * time.Second)

	go collector.collect(time.Duration(*pollInterval) * time.Second)

	for {
		sendMetrics(client, baseURL, collector.metrics())
		time.Sleep(time.Duration(*reportInterval) * time.Second)
	}
}

type metricsCollector struct {
	mu       sync.RWMutex
	memStats runtime.MemStats
	counter  int
}

func newMetricsCollector() *metricsCollector {
	collector := &metricsCollector{}
	runtime.ReadMemStats(&collector.memStats)
	return collector
}

func (c *metricsCollector) collect(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		runtime.ReadMemStats(&c.memStats)
		c.counter++
		c.mu.Unlock()
	}
}

func (c *metricsCollector) metrics() map[string]any {
	randomValue := rand.Float64()

	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]any{
		"gauge/Alloc":         c.memStats.Alloc,
		"gauge/BuckHashSys":   c.memStats.BuckHashSys,
		"gauge/Frees":         c.memStats.Frees,
		"gauge/GCCPUFraction": c.memStats.GCCPUFraction,
		"gauge/GCSys":         c.memStats.GCSys,
		"gauge/HeapAlloc":     c.memStats.HeapAlloc,
		"gauge/HeapIdle":      c.memStats.HeapIdle,
		"gauge/HeapInuse":     c.memStats.HeapInuse,
		"gauge/HeapObjects":   c.memStats.HeapObjects,
		"gauge/HeapReleased":  c.memStats.HeapReleased,
		"gauge/HeapSys":       c.memStats.HeapSys,
		"gauge/LastGC":        c.memStats.LastGC,
		"gauge/Lookups":       c.memStats.Lookups,
		"gauge/MCacheInuse":   c.memStats.MCacheInuse,
		"gauge/MCacheSys":     c.memStats.MCacheSys,
		"gauge/MSpanInuse":    c.memStats.MSpanInuse,
		"gauge/MSpanSys":      c.memStats.MSpanSys,
		"gauge/Mallocs":       c.memStats.Mallocs,
		"gauge/NextGC":        c.memStats.NextGC,
		"gauge/NumForcedGC":   c.memStats.NumForcedGC,
		"gauge/NumGC":         c.memStats.NumGC,
		"gauge/OtherSys":      c.memStats.OtherSys,
		"gauge/PauseTotalNs":  c.memStats.PauseTotalNs,
		"gauge/StackInuse":    c.memStats.StackInuse,
		"gauge/StackSys":      c.memStats.StackSys,
		"gauge/Sys":           c.memStats.Sys,
		"gauge/TotalAlloc":    c.memStats.TotalAlloc,
		"gauge/RandomValue":   randomValue,
		"counter/PollCount":   c.counter,
	}
}

func sendMetrics(client *resty.Client, baseURL string, metrics map[string]any) {
	for name, value := range metrics {
		url := baseURL + name + "/" + fmt.Sprint(value)

		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)
		if err != nil {
			fmt.Println(err)
		}
	}
}
