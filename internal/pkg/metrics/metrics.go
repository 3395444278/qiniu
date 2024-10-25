package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	APILatency   map[string][]time.Duration
	CacheHitRate float64
	RequestCount int64
	ErrorCount   int64
	mu           sync.RWMutex
}

var globalMetrics = &Metrics{
	APILatency: make(map[string][]time.Duration),
}

func RecordLatency(endpoint string, duration time.Duration) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.APILatency[endpoint] = append(
		globalMetrics.APILatency[endpoint],
		duration,
	)
}

func RecordCacheHit(hit bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	if hit {
		globalMetrics.CacheHitRate = (globalMetrics.CacheHitRate*float64(globalMetrics.RequestCount) + 1) /
			float64(globalMetrics.RequestCount+1)
	} else {
		globalMetrics.CacheHitRate = (globalMetrics.CacheHitRate * float64(globalMetrics.RequestCount)) /
			float64(globalMetrics.RequestCount+1)
	}
	globalMetrics.RequestCount++
}
