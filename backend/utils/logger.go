package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var (
	defaultLogger *slog.Logger
	jsonLogger    *slog.Logger
)

// LogLevel represents logging levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level  LogLevel `json:"level"`
	Format string   `json:"format"` // "text" or "json"
	File   string   `json:"file,omitempty"`
}

// InitLogger initializes the global logger
func InitLogger(config LoggerConfig) error {
	level := slog.LevelInfo
	switch config.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize source attribute to show only filename and line
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					source.File = getFileName(source.File)
				}
			}
			return a
		},
	}

	var handler slog.Handler
	var output = os.Stdout
	
	// If file is specified, write to file
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	if config.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler)
	defaultLogger = logger
	jsonLogger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	
	// Set as default logger for the slog package
	slog.SetDefault(logger)
	
	return nil
}

// GetLogger returns the default logger
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		_ = InitLogger(LoggerConfig{
			Level:  LogLevelInfo,
			Format: "text",
		})
	}
	return defaultLogger
}

// GetJSONLogger returns a JSON logger (useful for structured logging in production)
func GetJSONLogger() *slog.Logger {
	if jsonLogger == nil {
		_ = InitLogger(LoggerConfig{
			Level:  LogLevelInfo,
			Format: "json",
		})
	}
	return jsonLogger
}

// ContextLogger creates a logger with context information
func ContextLogger(ctx context.Context) *slog.Logger {
	logger := GetLogger()
	
	// Add request ID if available
	if reqID, ok := ctx.Value("requestID").(string); ok {
		logger = logger.With("request_id", reqID)
	}
	
	// Add user ID if available
	if userID, ok := ctx.Value("userID").(string); ok {
		logger = logger.With("user_id", userID)
	}
	
	// Add trace ID if available
	if traceID, ok := ctx.Value("traceID").(string); ok {
		logger = logger.With("trace_id", traceID)
	}
	
	return logger
}

// LogError logs an error with additional context
func LogError(ctx context.Context, err error, msg string, attrs ...any) {
	logger := ContextLogger(ctx)
	
	// Add caller information
	if pc, file, line, ok := runtime.Caller(1); ok {
		attrs = append(attrs, 
			"caller", fmt.Sprintf("%s:%d", getFileName(file), line),
			"function", runtime.FuncForPC(pc).Name(),
		)
	}
	
	attrs = append(attrs, "error", err.Error())
	logger.Error(msg, attrs...)
}

// LogInfo logs an info message with context
func LogInfo(ctx context.Context, msg string, attrs ...any) {
	logger := ContextLogger(ctx)
	logger.Info(msg, attrs...)
}

// LogDebug logs a debug message with context
func LogDebug(ctx context.Context, msg string, attrs ...any) {
	logger := ContextLogger(ctx)
	logger.Debug(msg, attrs...)
}

// LogWarn logs a warning message with context
func LogWarn(ctx context.Context, msg string, attrs ...any) {
	logger := ContextLogger(ctx)
	logger.Warn(msg, attrs...)
}

// LogRepositoryOperation logs database operations
func LogRepositoryOperation(ctx context.Context, operation, collection string, duration time.Duration, err error, attrs ...any) {
	logger := ContextLogger(ctx)
	
	allAttrs := []any{
		"operation", operation,
		"collection", collection,
		"duration_ms", duration.Milliseconds(),
	}
	allAttrs = append(allAttrs, attrs...)
	
	if err != nil {
		allAttrs = append(allAttrs, "error", err.Error())
		logger.Error("Repository operation failed", allAttrs...)
	} else {
		logger.Debug("Repository operation completed", allAttrs...)
	}
}

// LogHTTPRequest logs HTTP requests
func LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, attrs ...any) {
	logger := ContextLogger(ctx)
	
	allAttrs := []any{
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	}
	allAttrs = append(allAttrs, attrs...)
	
	level := slog.LevelInfo
	if statusCode >= 400 && statusCode < 500 {
		level = slog.LevelWarn
	} else if statusCode >= 500 {
		level = slog.LevelError
	}
	
	logger.Log(ctx, level, "HTTP request", allAttrs...)
}

// LogMiddleware logs middleware operations
func LogMiddleware(ctx context.Context, middlewareName string, success bool, duration time.Duration, attrs ...any) {
	logger := ContextLogger(ctx)
	
	allAttrs := []any{
		"middleware", middlewareName,
		"success", success,
		"duration_ms", duration.Milliseconds(),
	}
	allAttrs = append(allAttrs, attrs...)
	
	if success {
		logger.Debug("Middleware executed successfully", allAttrs...)
	} else {
		logger.Warn("Middleware execution failed", allAttrs...)
	}
}

// Performance logging utilities
type PerformanceLogger struct {
	ctx       context.Context
	operation string
	startTime time.Time
	logger    *slog.Logger
}

// StartPerformanceLogging starts performance monitoring for an operation
func StartPerformanceLogging(ctx context.Context, operation string) *PerformanceLogger {
	return &PerformanceLogger{
		ctx:       ctx,
		operation: operation,
		startTime: time.Now(),
		logger:    ContextLogger(ctx),
	}
}

// End completes the performance logging
func (p *PerformanceLogger) End(attrs ...any) {
	duration := time.Since(p.startTime)
	
	allAttrs := []any{
		"operation", p.operation,
		"duration_ms", duration.Milliseconds(),
	}
	allAttrs = append(allAttrs, attrs...)
	
	// Log as warning if operation takes too long
	if duration > 5*time.Second {
		p.logger.Warn("Slow operation detected", allAttrs...)
	} else {
		p.logger.Debug("Operation completed", allAttrs...)
	}
}

// EndWithError completes the performance logging with an error
func (p *PerformanceLogger) EndWithError(err error, attrs ...any) {
	duration := time.Since(p.startTime)
	
	allAttrs := []any{
		"operation", p.operation,
		"duration_ms", duration.Milliseconds(),
		"error", err.Error(),
	}
	allAttrs = append(allAttrs, attrs...)
	
	p.logger.Error("Operation failed", allAttrs...)
}

// Helper function to extract filename from full path
func getFileName(fullPath string) string {
	for i := len(fullPath) - 1; i >= 0; i-- {
		if fullPath[i] == '/' || fullPath[i] == '\\' {
			return fullPath[i+1:]
		}
	}
	return fullPath
}

// Structured error types for better logging
type LoggedError struct {
	Err        error                  `json:"error"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

func (e *LoggedError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

// NewLoggedError creates a new logged error
func NewLoggedError(code, message string, err error, context map[string]interface{}) *LoggedError {
	return &LoggedError{
		Err:       err,
		Code:      code,
		Message:   message,
		Context:   context,
		Timestamp: time.Now(),
	}
}

// LogAndReturnError logs an error and returns a LoggedError
func LogAndReturnError(ctx context.Context, code, message string, err error, context map[string]interface{}) *LoggedError {
	loggedErr := NewLoggedError(code, message, err, context)
	
	attrs := []any{
		"error_code", code,
		"error_message", message,
	}
	
	for k, v := range context {
		attrs = append(attrs, k, v)
	}
	
	LogError(ctx, err, message, attrs...)
	
	return loggedErr
}