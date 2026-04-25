package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Logger provides structured logging capabilities
type Logger struct {
	serviceName string
	logLevel    LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
		logLevel:    getLogLevelFromEnv(),
	}
}

// getLogLevelFromEnv determines log level from environment
func getLogLevelFromEnv() LogLevel {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.logLevel
}

// formatLevel converts LogLevel to string
func (l *Logger) formatLevel(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "INFO"
	}
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, message string, context map[string]interface{}, err error) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     l.formatLevel(level),
		Message:   message,
		Service:   l.serviceName,
		Context:   context,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Format log message for console output
	logMessage := fmt.Sprintf("[%s] %s [%s] %s", 
		entry.Timestamp.Format("2006-01-02T15:04:05Z"),
		entry.Level,
		entry.Service,
		entry.Message,
	)

	// Add context information
	if len(entry.Context) > 0 {
		for k, v := range entry.Context {
			logMessage += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	// Add error information
	if entry.Error != "" {
		logMessage += fmt.Sprintf(" error=%s", entry.Error)
	}

	// Write to standard output
	log.Println(logMessage)

	// Exit on fatal errors
	if level == LogLevelFatal {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LogLevelDebug, message, ctx, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LogLevelInfo, message, ctx, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LogLevelWarn, message, ctx, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LogLevelError, message, ctx, err)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LogLevelFatal, message, ctx, err)
}

// WithContext creates a logger with additional context
func (l *Logger) WithContext(context map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		context: context,
	}
}

// ContextLogger provides logging with pre-set context
type ContextLogger struct {
	logger  *Logger
	context map[string]interface{}
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(message string, additionalContext ...map[string]interface{}) {
	ctx := cl.mergeContext(additionalContext...)
	cl.logger.Debug(message, ctx)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(message string, additionalContext ...map[string]interface{}) {
	ctx := cl.mergeContext(additionalContext...)
	cl.logger.Info(message, ctx)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(message string, additionalContext ...map[string]interface{}) {
	ctx := cl.mergeContext(additionalContext...)
	cl.logger.Warn(message, ctx)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(message string, err error, additionalContext ...map[string]interface{}) {
	ctx := cl.mergeContext(additionalContext...)
	cl.logger.Error(message, err, ctx)
}

// Fatal logs a fatal message with context and exits
func (cl *ContextLogger) Fatal(message string, err error, additionalContext ...map[string]interface{}) {
	ctx := cl.mergeContext(additionalContext...)
	cl.logger.Fatal(message, err, ctx)
}

// mergeContext combines the base context with additional context
func (cl *ContextLogger) mergeContext(additionalContext ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy base context
	for k, v := range cl.context {
		result[k] = v
	}
	
	// Add additional context
	for _, additional := range additionalContext {
		for k, v := range additional {
			result[k] = v
		}
	}
	
	return result
}

// WithContext creates a new ContextLogger with additional context
func (cl *ContextLogger) WithContext(context map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger:  cl.logger,
		context: cl.mergeContext(context),
	}
}

// GetContext returns a copy of the current context
func (cl *ContextLogger) GetContext() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range cl.context {
		result[k] = v
	}
	return result
}

// Global logger instance
var defaultLogger = NewLogger("indel")

// Global logging functions for convenience
func Debug(message string, context ...map[string]interface{}) {
	defaultLogger.Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	defaultLogger.Info(message, context...)
}

func Warn(message string, context ...map[string]interface{}) {
	defaultLogger.Warn(message, context...)
}

func Error(message string, err error, context ...map[string]interface{}) {
	defaultLogger.Error(message, err, context...)
}

func Fatal(message string, err error, context ...map[string]interface{}) {
	defaultLogger.Fatal(message, err, context...)
}

// WithContext creates a context logger with the global logger
func WithContext(context map[string]interface{}) *ContextLogger {
	return defaultLogger.WithContext(context)
}

// ExtractContextFromRequest extracts common context from a request
func ExtractContextFromRequest(requestID, userID, policyID string) map[string]interface{} {
	ctx := make(map[string]interface{})
	if requestID != "" {
		ctx["request_id"] = requestID
	}
	if userID != "" {
		ctx["user_id"] = userID
	}
	if policyID != "" {
		ctx["policy_id"] = policyID
	}
	return ctx
}

// ExtractContextFromContext extracts context from a Go context
func ExtractContextFromContext(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	
	if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
		result["request_id"] = reqID
	}
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		result["user_id"] = userID
	}
	if policyID, ok := ctx.Value("policy_id").(string); ok && policyID != "" {
		result["policy_id"] = policyID
	}
	if corrID, ok := ctx.Value("correlation_id").(string); ok && corrID != "" {
		result["correlation_id"] = corrID
	}
	
	return result
}
