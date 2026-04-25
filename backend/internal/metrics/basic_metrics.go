package metrics

import (
	"sync"
	"time"
)

// MetricsCollector provides basic metrics collection capabilities
type MetricsCollector struct {
	mu                sync.RWMutex
	counters          map[string]int64
	gauges            map[string]float64
	histograms        map[string]*Histogram
	timers            map[string]*Timer
	startTime         time.Time
}

// Histogram tracks distribution of values
type Histogram struct {
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
	mu      sync.RWMutex
}

// Timer tracks timing information
type Timer struct {
	count   int64
	sum     time.Duration
	min     time.Duration
	max     time.Duration
	last    time.Duration
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:  make(map[string]int64),
		gauges:    make(map[string]float64),
		histograms: make(map[string]*Histogram),
		timers:    make(map[string]*Timer),
		startTime: time.Now(),
	}
}

// Increment increments a counter metric
func (mc *MetricsCollector) Increment(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name]++
}

// Add adds a value to a counter metric
func (mc *MetricsCollector) Add(name string, value int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name] += value
}

// Set sets a gauge metric value
func (mc *MetricsCollector) Set(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.gauges[name] = value
}

// RecordValue records a value in a histogram
func (mc *MetricsCollector) RecordValue(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	hist, exists := mc.histograms[name]
	if !exists {
		// Create histogram with default buckets
		hist = &Histogram{
			buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0, 2500.0, 5000.0, 10000.0},
			counts:  make([]int64, 16),
		}
		mc.histograms[name] = hist
	}
	
	hist.mu.Lock()
	defer hist.mu.Unlock()
	
	hist.sum += value
	hist.count++
	
	// Find appropriate bucket
	for i, bucket := range hist.buckets {
		if value <= bucket {
			hist.counts[i]++
			break
		}
	}
}

// RecordDuration records a duration in a timer
func (mc *MetricsCollector) RecordDuration(name string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	timer, exists := mc.timers[name]
	if !exists {
		timer = &Timer{}
		mc.timers[name] = timer
	}
	
	timer.mu.Lock()
	defer timer.mu.Unlock()
	
	timer.count++
	timer.sum += duration
	timer.last = duration
	
	if timer.min == 0 || duration < timer.min {
		timer.min = duration
	}
	if duration > timer.max {
		timer.max = duration
	}
}

// TimeOperation measures the duration of an operation
func (mc *MetricsCollector) TimeOperation(name string, operation func()) {
	start := time.Now()
	operation()
	mc.RecordDuration(name, time.Since(start))
}

// GetCounters returns all counter metrics
func (mc *MetricsCollector) GetCounters() map[string]int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range mc.counters {
		result[k] = v
	}
	return result
}

// GetGauges returns all gauge metrics
func (mc *MetricsCollector) GetGauges() map[string]float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]float64)
	for k, v := range mc.gauges {
		result[k] = v
	}
	return result
}

// GetHistograms returns all histogram metrics
func (mc *MetricsCollector) GetHistograms() map[string]HistogramData {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]HistogramData)
	for name, hist := range mc.histograms {
		hist.mu.RLock()
		result[name] = HistogramData{
			Buckets: hist.buckets,
			Counts:  make([]int64, len(hist.counts)),
			Sum:     hist.sum,
			Count:   hist.count,
		}
		copy(result[name].Counts, hist.counts)
		hist.mu.RUnlock()
	}
	return result
}

// GetTimers returns all timer metrics
func (mc *MetricsCollector) GetTimers() map[string]TimerData {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]TimerData)
	for name, timer := range mc.timers {
		timer.mu.RLock()
		result[name] = TimerData{
			Count: timer.count,
			Sum:   timer.sum,
			Min:   timer.min,
			Max:   timer.max,
			Last:  timer.last,
		}
		timer.mu.RUnlock()
	}
	return result
}

// GetUptime returns the uptime of the metrics collector
func (mc *MetricsCollector) GetUptime() time.Duration {
	return time.Since(mc.startTime)
}

// GetAllMetrics returns all metrics in a structured format
func (mc *MetricsCollector) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"uptime":    mc.GetUptime().String(),
		"counters":  mc.GetCounters(),
		"gauges":    mc.GetGauges(),
		"histograms": mc.GetHistograms(),
		"timers":    mc.GetTimers(),
	}
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.counters = make(map[string]int64)
	mc.gauges = make(map[string]float64)
	mc.histograms = make(map[string]*Histogram)
	mc.timers = make(map[string]*Timer)
	mc.startTime = time.Now()
}

// HistogramData represents histogram data for export
type HistogramData struct {
	Buckets []float64 `json:"buckets"`
	Counts  []int64   `json:"counts"`
	Sum     float64   `json:"sum"`
	Count   int64     `json:"count"`
}

// TimerData represents timer data for export
type TimerData struct {
	Count int64         `json:"count"`
	Sum   time.Duration `json:"sum"`
	Min   time.Duration `json:"min"`
	Max   time.Duration `json:"max"`
	Last  time.Duration `json:"last"`
}

// Global metrics collector instance
var GlobalMetrics = NewMetricsCollector()

// Convenience functions for global metrics
func Increment(name string) {
	GlobalMetrics.Increment(name)
}

func Add(name string, value int64) {
	GlobalMetrics.Add(name, value)
}

func Set(name string, value float64) {
	GlobalMetrics.Set(name, value)
}

func RecordValue(name string, value float64) {
	GlobalMetrics.RecordValue(name, value)
}

func RecordDuration(name string, duration time.Duration) {
	GlobalMetrics.RecordDuration(name, duration)
}

func TimeOperation(name string, operation func()) {
	GlobalMetrics.TimeOperation(name, operation)
}

// Business-specific metrics helpers
func IncrementPolicyActivation() {
	Increment("policy_activations_total")
}

func IncrementPolicyCreation() {
	Increment("policy_creations_total")
}

func IncrementClaimGeneration() {
	Increment("claim_generations_total")
}

func IncrementDisruptionEvents() {
	Increment("disruption_events_total")
}

func IncrementKafkaMessagesSent(topic string) {
	Increment("kafka_messages_sent_total")
	Set("kafka_messages_sent_last_topic", float64(hashString(topic)))
}

func IncrementKafkaMessagesFailed(topic string) {
	Increment("kafka_messages_failed_total")
	Set("kafka_messages_failed_last_topic", float64(hashString(topic)))
}

func IncrementDatabaseQueries(queryType string) {
	Increment("database_queries_total")
	RecordValue("database_query_type", float64(hashString(queryType)))
}

func IncrementDatabaseErrors() {
	Increment("database_errors_total")
}

func IncrementMLRequests(endpoint string) {
	Increment("ml_requests_total")
	RecordValue("ml_endpoint", float64(hashString(endpoint)))
}

func IncrementMLErrors(endpoint string) {
	Increment("ml_errors_total")
	RecordValue("ml_error_endpoint", float64(hashString(endpoint)))
}

func RecordMLLatency(duration time.Duration) {
	RecordDuration("ml_request_duration", duration)
}

func RecordPolicyProcessingTime(duration time.Duration) {
	RecordDuration("policy_processing_duration", duration)
}

func RecordClaimProcessingTime(duration time.Duration) {
	RecordDuration("claim_processing_duration", duration)
}

func SetActivePolicies(count int64) {
	Set("active_policies_count", float64(count))
}

func SetLockedPolicies(count int64) {
	Set("locked_policies_count", float64(count))
}

func SetActiveWorkers(count int64) {
	Set("active_workers_count", float64(count))
}

func SetPendingClaims(count int64) {
	Set("pending_claims_count", float64(count))
}

func hashString(s string) int64 {
	// Simple hash function for categorizing string values
	hash := int64(0)
	for _, c := range s {
		hash = hash*31 + int64(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
