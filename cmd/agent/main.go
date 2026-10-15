package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"resty.dev/v3"
)

func main() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var counter int
	metrics := map[string]func() any{
		"gauge/Alloc":         func() any { return memStats.Alloc },
		"gauge/BuckHashSys":   func() any { return memStats.BuckHashSys },
		"gauge/Frees":         func() any { return memStats.Frees },
		"gauge/GCCPUFraction": func() any { return memStats.GCCPUFraction },
		"gauge/GCSys":         func() any { return memStats.GCSys },
		"gauge/HeapAlloc":     func() any { return memStats.HeapAlloc },
		"gauge/HeapIdle":      func() any { return memStats.HeapIdle },
		"gauge/HeapInuse":     func() any { return memStats.HeapInuse },
		"gauge/HeapObjects":   func() any { return memStats.HeapObjects },
		"gauge/HeapReleased":  func() any { return memStats.HeapReleased },
		"gauge/HeapSys":       func() any { return memStats.HeapSys },
		"gauge/LastGC":        func() any { return memStats.LastGC },
		"gauge/Lookups":       func() any { return memStats.Lookups },
		"gauge/MCacheInuse":   func() any { return memStats.MCacheInuse },
		"gauge/MCacheSys":     func() any { return memStats.MCacheSys },
		"gauge/MSpanInuse":    func() any { return memStats.MSpanInuse },
		"gauge/MSpanSys":      func() any { return memStats.MSpanSys },
		"gauge/Mallocs":       func() any { return memStats.Mallocs },
		"gauge/NextGC":        func() any { return memStats.NextGC },
		"gauge/NumForcedGC":   func() any { return memStats.NumForcedGC },
		"gauge/NumGC":         func() any { return memStats.NumGC },
		"gauge/OtherSys":      func() any { return memStats.OtherSys },
		"gauge/PauseTotalNs":  func() any { return memStats.PauseTotalNs },
		"gauge/StackInuse":    func() any { return memStats.StackInuse },
		"gauge/StackSys":      func() any { return memStats.StackSys },
		"gauge/Sys":           func() any { return memStats.Sys },
		"gauge/TotalAlloc":    func() any { return memStats.TotalAlloc },
		"gauge/RandomValue":   func() any { return rand.Float64() },
		"counter/PollCount":   func() any { return counter },
	}

	client := resty.New().SetTimeout(10 * time.Second)

	go collectMetrics(&counter, &memStats)

	for {
		sendMetrics(client, metrics)
		time.Sleep(10 * time.Second)
	}
}

var baseUrl = "http://localhost:8080/update/"

func sendMetrics(client *resty.Client, metrics map[string]func() any) {
	for k, v := range metrics {
		var url = baseUrl + k + "/" + fmt.Sprint(v())

		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func collectMetrics(counter *int, memStats *runtime.MemStats) {
	for range time.Tick(2 * time.Second) {
		runtime.ReadMemStats(memStats)
		*counter += 1
	}
}
