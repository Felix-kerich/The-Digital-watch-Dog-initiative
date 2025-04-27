package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance
var Logger *logrus.Logger

// NamedLogger represents a logger with a specific name for a component
type NamedLogger struct {
	logger *logrus.Logger
	name   string
}

// NewLogger creates a new named logger for a specific component
func NewLogger(name string) *NamedLogger {
	return &NamedLogger{
		logger: Logger,
		name:   name,
	}
}

// Info logs an info level message with the named logger
func (l *NamedLogger) Info(message string, fields map[string]interface{}) {
	contextFields := logrus.Fields{
		"component": l.name,
	}
	
	for k, v := range fields {
		contextFields[k] = v
	}
	
	l.logger.WithFields(contextFields).Info(message)
}

// Error logs an error level message with the named logger
func (l *NamedLogger) Error(message string, fields map[string]interface{}) {
	contextFields := logrus.Fields{
		"component": l.name,
	}
	
	for k, v := range fields {
		contextFields[k] = v
	}
	
	l.logger.WithFields(contextFields).Error(message)
}

// Warn logs a warning level message with the named logger
func (l *NamedLogger) Warn(message string, fields map[string]interface{}) {
	contextFields := logrus.Fields{
		"component": l.name,
	}
	
	for k, v := range fields {
		contextFields[k] = v
	}
	
	l.logger.WithFields(contextFields).Warn(message)
}

// Debug logs a debug level message with the named logger
func (l *NamedLogger) Debug(message string, fields map[string]interface{}) {
	contextFields := logrus.Fields{
		"component": l.name,
	}
	
	for k, v := range fields {
		contextFields[k] = v
	}
	
	l.logger.WithFields(contextFields).Debug(message)
}

// InitLogger initializes the logger with proper configuration
func InitLogger() {
	Logger = logrus.New()

	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Default to info level
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// Configure log format
	Logger.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := filepath.Base(f.File)
			return funcName, fileName + ":" + fmt.Sprintf("%d", f.Line)
		},
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Log to both file and stdout in production
	if os.Getenv("APP_ENV") == "production" {
		// Create log file
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
			os.Exit(1)
		}

		logFile, err := os.OpenFile(
			filepath.Join(logDir, "api.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			os.Exit(1)
		}

		// Write logs to both stdout and file
		mw := io.MultiWriter(os.Stdout, logFile)
		Logger.SetOutput(mw)
	} else {
		// In development, just write to stdout
		Logger.SetOutput(os.Stdout)
		// Use text formatter for better readability in development
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Enable caller information
	Logger.SetReportCaller(true)

	Logger.Info("Logger initialized successfully")
}

// LoggerMiddleware creates a middleware for logging HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Get the real IP address
		ip := c.Request.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = c.Request.RemoteAddr
		} else {
			// X-Forwarded-For can contain multiple IPs, take the first one
			ip = strings.Split(ip, ",")[0]
		}

		// Logging after request
		latency := time.Since(start)
		clientIP := ip
		method := c.Request.Method
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()

		// Get user ID if authenticated
		userID, exists := c.Get("userID")
		if !exists {
			userID = "anonymous"
		}

		entry := Logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"latency":    latency,
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"user_agent": userAgent,
			"user_id":    userID,
		})

		if statusCode >= 500 {
			entry.Error("Server error")
		} else if statusCode >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request processed")
		}
	}
}
