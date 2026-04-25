package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/logging"
)

// WorkerConfig defines configuration for a safe worker
type WorkerConfig struct {
	Name             string
	PollInterval     time.Duration
	MaxConcurrent    int           // Maximum number of concurrent operations
	ShutdownTimeout  time.Duration // Time to wait for graceful shutdown
	HealthCheckInterval time.Duration // Interval for health checks
}

// SafeWorker provides a base for implementing workers with proper lifecycle management
type SafeWorker struct {
	config      WorkerConfig
	logger      *logging.ContextLogger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
	healthCheck time.Time
	lastError   error
	metrics     *WorkerMetrics
}

// WorkerMetrics tracks worker performance and health
type WorkerMetrics struct {
	mu               sync.RWMutex
	operationsTotal  int64
	operationsSuccess int64
	operationsFailed  int64
	averageLatency    time.Duration
	lastOperationTime time.Time
	startTime         time.Time
}

// NewSafeWorker creates a new safe worker instance
func NewSafeWorker(config WorkerConfig) *SafeWorker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SafeWorker{
		config:  config,
		logger:  logging.WithContext(map[string]interface{}{
			"worker_name": config.Name,
			"service": "safe-worker",
		}),
		ctx:     ctx,
		cancel:  cancel,
		metrics: &WorkerMetrics{
			startTime: time.Now(),
		},
	}
}

// Start begins the worker's main loop
func (sw *SafeWorker) Start(workFunc func(context.Context) error) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	
	if sw.running {
		return fmt.Errorf("worker %s is already running", sw.config.Name)
	}
	
	sw.running = true
	sw.metrics.startTime = time.Now()
	
	sw.logger.Info("Starting worker", map[string]interface{}{
		"poll_interval": sw.config.PollInterval.String(),
		"max_concurrent": sw.config.MaxConcurrent,
		"shutdown_timeout": sw.config.ShutdownTimeout.String(),
	})
	
	// Start main worker goroutine
	sw.wg.Add(1)
	go sw.runMainLoop(workFunc)
	
	// Start health check goroutine
	if sw.config.HealthCheckInterval > 0 {
		sw.wg.Add(1)
		go sw.runHealthCheck()
	}
	
	return nil
}

// Stop gracefully shuts down the worker
func (sw *SafeWorker) Stop() error {
	sw.mu.Lock()
	if !sw.running {
		sw.mu.Unlock()
		return nil
	}
	sw.running = false
	sw.mu.Unlock()
	
	sw.logger.Info("Stopping worker")
	
	// Signal cancellation
	sw.cancel()
	
	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		sw.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		sw.logger.Info("Worker stopped gracefully")
		return nil
	case <-time.After(sw.config.ShutdownTimeout):
		sw.logger.Warn("Worker shutdown timeout, forcing stop", nil)
		return fmt.Errorf("worker %s shutdown timeout", sw.config.Name)
	}
}

// IsRunning returns whether the worker is currently running
func (sw *SafeWorker) IsRunning() bool {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.running
}

// GetMetrics returns current worker metrics
func (sw *SafeWorker) GetMetrics() WorkerMetrics {
	sw.metrics.mu.RLock()
	defer sw.metrics.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	return WorkerMetrics{
		operationsTotal:   sw.metrics.operationsTotal,
		operationsSuccess: sw.metrics.operationsSuccess,
		operationsFailed:  sw.metrics.operationsFailed,
		averageLatency:    sw.metrics.averageLatency,
		lastOperationTime: sw.metrics.lastOperationTime,
		startTime:         sw.metrics.startTime,
	}
}

// GetHealth returns the current health status
func (sw *SafeWorker) GetHealth() map[string]interface{} {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	
	metrics := sw.GetMetrics()
	
	health := map[string]interface{}{
		"worker_name":        sw.config.Name,
		"running":           sw.running,
		"uptime":           time.Since(metrics.startTime).String(),
		"operations_total":  metrics.operationsTotal,
		"operations_success": metrics.operationsSuccess,
		"operations_failed":  metrics.operationsFailed,
		"success_rate":      sw.calculateSuccessRate(metrics),
		"average_latency":   metrics.averageLatency.String(),
		"last_operation":    metrics.lastOperationTime.Format(time.RFC3339),
		"last_health_check": sw.healthCheck.Format(time.RFC3339),
	}
	
	if sw.lastError != nil {
		health["last_error"] = sw.lastError.Error()
		health["last_error_time"] = time.Now().Format(time.RFC3339)
	}
	
	return health
}

// runMainLoop executes the worker's main processing loop
func (sw *SafeWorker) runMainLoop(workFunc func(context.Context) error) {
	defer sw.wg.Done()
	
	ticker := time.NewTicker(sw.config.PollInterval)
	defer ticker.Stop()
	
	// Run once immediately
	sw.runWork(workFunc)
	
	for {
		select {
		case <-sw.ctx.Done():
			sw.logger.Info("Main loop stopped due to context cancellation")
			return
		case <-ticker.C:
			sw.runWork(workFunc)
		}
	}
}

// runWork executes a single work cycle with metrics tracking
func (sw *SafeWorker) runWork(workFunc func(context.Context) error) {
	start := time.Now()
	
	// Update last health check
	sw.mu.Lock()
	sw.healthCheck = start
	sw.mu.Unlock()
	
	// Execute work function
	err := workFunc(sw.ctx)
	
	// Record metrics
	sw.recordMetrics(time.Since(start), err)
	
	if err != nil {
		sw.mu.Lock()
		sw.lastError = err
		sw.mu.Unlock()
		
		sw.logger.Error("Work cycle failed", err, map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	} else {
		sw.logger.Debug("Work cycle completed successfully", map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	}
}

// runHealthCheck performs periodic health checks
func (sw *SafeWorker) runHealthCheck() {
	defer sw.wg.Done()
	
	ticker := time.NewTicker(sw.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sw.ctx.Done():
			sw.logger.Info("Health check loop stopped due to context cancellation")
			return
		case <-ticker.C:
			sw.performHealthCheck()
		}
	}
}

// performHealthCheck executes a health check
func (sw *SafeWorker) performHealthCheck() {
	sw.mu.Lock()
	sw.healthCheck = time.Now()
	sw.mu.Unlock()
	
	// Check if worker is healthy
	health := sw.GetHealth()
	
	// Log health status periodically
	sw.logger.Debug("Health check completed", map[string]interface{}{
		"success_rate": health["success_rate"],
		"operations_total": health["operations_total"],
		"uptime": health["uptime"],
	})
}

// recordMetrics updates worker metrics after each operation
func (sw *SafeWorker) recordMetrics(duration time.Duration, err error) {
	sw.metrics.mu.Lock()
	defer sw.metrics.mu.Unlock()
	
	sw.metrics.operationsTotal++
	sw.metrics.lastOperationTime = time.Now()
	
	if err == nil {
		sw.metrics.operationsSuccess++
	} else {
		sw.metrics.operationsFailed++
	}
	
	// Update average latency (simple moving average)
	if sw.metrics.operationsTotal == 1 {
		sw.metrics.averageLatency = duration
	} else {
		// Weighted average: give more weight to recent operations
		weight := 0.1
		sw.metrics.averageLatency = time.Duration(
			(float64(sw.metrics.averageLatency) * (1 - weight)) + 
			(float64(duration) * weight),
		)
	}
}

// calculateSuccessRate calculates the success rate as a percentage
func (sw *SafeWorker) calculateSuccessRate(metrics WorkerMetrics) float64 {
	if metrics.operationsTotal == 0 {
		return 0.0
	}
	
	return float64(metrics.operationsSuccess) / float64(metrics.operationsTotal) * 100.0
}

// ConcurrencyLimiter limits the number of concurrent operations
type ConcurrencyLimiter struct {
	semaphore chan struct{}
	maxCount  int
	current   int
	mu        sync.RWMutex
}

// NewConcurrencyLimiter creates a new concurrency limiter
func NewConcurrencyLimiter(maxCount int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, maxCount),
		maxCount:  maxCount,
	}
}

// Acquire attempts to acquire a slot for concurrent operation
func (cl *ConcurrencyLimiter) Acquire(ctx context.Context) error {
	select {
	case cl.semaphore <- struct{}{}:
		cl.mu.Lock()
		cl.current++
		cl.mu.Unlock()
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for concurrency slot: %w", ctx.Err())
	}
}

// Release releases a slot for concurrent operation
func (cl *ConcurrencyLimiter) Release() {
	<-cl.semaphore
	cl.mu.Lock()
	cl.current--
	cl.mu.Unlock()
}

// GetCurrent returns the current number of active operations
func (cl *ConcurrencyLimiter) GetCurrent() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.current
}

// GetMax returns the maximum number of concurrent operations
func (cl *ConcurrencyLimiter) GetMax() int {
	return cl.maxCount
}

// WithConcurrency executes a function with concurrency limiting
func (cl *ConcurrencyLimiter) WithConcurrency(ctx context.Context, workFunc func() error) error {
	if err := cl.Acquire(ctx); err != nil {
		return err
	}
	defer cl.Release()
	
	return workFunc()
}

// SafeWorkerPool manages multiple workers with proper lifecycle
type SafeWorkerPool struct {
	workers   []*SafeWorker
	limiter   *ConcurrencyLimiter
	logger    *logging.ContextLogger
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
}

// NewSafeWorkerPool creates a new worker pool
func NewSafeWorkerPool(maxConcurrent int, name string) *SafeWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SafeWorkerPool{
		limiter: NewConcurrencyLimiter(maxConcurrent),
		logger: logging.WithContext(map[string]interface{}{
			"pool_name": name,
			"service": "safe-worker-pool",
		}),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddWorker adds a worker to the pool
func (swp *SafeWorkerPool) AddWorker(worker *SafeWorker) {
	swp.mu.Lock()
	defer swp.mu.Unlock()
	
	swp.workers = append(swp.workers, worker)
}

// StartAll starts all workers in the pool
func (swp *SafeWorkerPool) StartAll(workFunc func(context.Context) error) error {
	swp.mu.Lock()
	defer swp.mu.Unlock()
	
	if swp.running {
		return fmt.Errorf("worker pool is already running")
	}
	
	swp.running = true
	
	swp.logger.Info("Starting worker pool", map[string]interface{}{
		"worker_count": len(swp.workers),
		"max_concurrent": swp.limiter.GetMax(),
	})
	
	for _, worker := range swp.workers {
		if err := worker.Start(workFunc); err != nil {
			// Stop any already started workers
			swp.stopAll()
			return fmt.Errorf("failed to start worker %s: %w", worker.config.Name, err)
		}
	}
	
	return nil
}

// StopAll stops all workers in the pool
func (swp *SafeWorkerPool) StopAll() error {
	swp.mu.Lock()
	defer swp.mu.Unlock()
	
	return swp.stopAll()
}

// stopAll internal method to stop all workers
func (swp *SafeWorkerPool) stopAll() error {
	if !swp.running {
		return nil
	}
	
	swp.running = false
	swp.logger.Info("Stopping worker pool")
	
	// Signal cancellation
	swp.cancel()
	
	// Stop all workers
	var errors []error
	for _, worker := range swp.workers {
		if err := worker.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop worker %s: %w", worker.config.Name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors stopping workers: %v", errors)
	}
	
	swp.logger.Info("Worker pool stopped")
	return nil
}

// GetPoolHealth returns health information for all workers
func (swp *SafeWorkerPool) GetPoolHealth() map[string]interface{} {
	swp.mu.RLock()
	defer swp.mu.RUnlock()
	
	workerHealth := make([]map[string]interface{}, len(swp.workers))
	for i, worker := range swp.workers {
		workerHealth[i] = worker.GetHealth()
	}
	
	loggerContext := swp.logger.GetContext()
	return map[string]interface{}{
		"pool_name":       loggerContext["pool_name"],
		"running":         swp.running,
		"worker_count":    len(swp.workers),
		"max_concurrent":  swp.limiter.GetMax(),
		"current_active":  swp.limiter.GetCurrent(),
		"workers":         workerHealth,
	}
}
