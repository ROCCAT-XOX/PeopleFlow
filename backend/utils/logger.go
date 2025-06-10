// backend/utils/logger.go
package utils

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// Logger configuration
type LoggerConfig struct {
	Level       LogLevel
	Format      string // "json" or "text"
	AddSource   bool
	ServiceName string
	Environment string
}

// InitLogger initializes the global logger
func InitLogger(config LoggerConfig) *slog.Logger {
	var level slog.Level
	switch config.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}
			// Add caller information
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				return slog.Group("caller",
					slog.String("file", source.File),
					slog.Int("line", source.Line),
					slog.String("function", source.Function),
				)
			}
			return a
		},
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Add default attributes
	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", config.ServiceName),
		slog.String("environment", config.Environment),
		slog.String("version", getVersion()),
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// getVersion returns the application version (you can implement this based on your needs)
func getVersion() string {
	// This could be set from build flags or environment variables
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	return "dev"
}

// RequestLogger creates a logger with request context
func RequestLogger(ctx context.Context, logger *slog.Logger) *slog.Logger {
	// Extract request ID from context if available
	if requestID, ok := ctx.Value("requestID").(string); ok {
		logger = logger.With("requestID", requestID)
	}

	// Extract user ID from context if available
	if userID, ok := ctx.Value("userID").(string); ok {
		logger = logger.With("userID", userID)
	}

	return logger
}

// LogError logs an error with additional context
func LogError(logger *slog.Logger, err error, msg string, attrs ...slog.Attr) {
	if err == nil {
		return
	}

	// Add stack trace for better debugging
	_, file, line, _ := runtime.Caller(1)

	attrs = append(attrs,
		slog.String("error", err.Error()),
		slog.String("file", file),
		slog.Int("line", line),
	)

	logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

// LogPanic logs a panic and recovers
func LogPanic(logger *slog.Logger) {
	if r := recover(); r != nil {
		// Get stack trace
		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logger.Error("Panic recovered",
			"panic", r,
			"stack_trace", stackTrace,
		)

		// Re-panic after logging
		panic(r)
	}
}

// Middleware for HTTP handlers to add request context to logger
func LoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := generateRequestID()

			// Create logger with request context
			reqLogger := logger.With(
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"request_id", requestID,
			)

			// Add logger to context
			ctx := context.WithValue(r.Context(), "logger", reqLogger)
			ctx = context.WithValue(ctx, "requestID", requestID)

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log request completion
			reqLogger.Info("Request completed",
				"status", wrapped.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"bytes_written", wrapped.bytesWritten,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(bytes []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(bytes)
	rw.bytesWritten += n
	return n, err
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return primitive.NewObjectID().Hex()
}

// GetLoggerFromContext extracts logger from context
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
