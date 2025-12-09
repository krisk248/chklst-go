package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Logger is a structured JSON logger
type Logger struct {
	minLevel LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(minLevel LogLevel) *Logger {
	return &Logger{
		minLevel: minLevel,
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, requestID string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		RequestID: requestID,
		Data:      data,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	fmt.Fprintln(os.Stdout, string(jsonBytes))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, data map[string]interface{}) {
	if l.minLevel == DEBUG {
		l.log(DEBUG, message, "", data)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, data map[string]interface{}) {
	l.log(INFO, message, "", data)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, data map[string]interface{}) {
	l.log(WARN, message, "", data)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}
	l.log(ERROR, message, "", data)
}

// WithRequestID logs with request ID
func (l *Logger) WithRequestID(requestID string) *RequestLogger {
	return &RequestLogger{
		logger:    l,
		requestID: requestID,
	}
}

// RequestLogger is a logger with request ID context
type RequestLogger struct {
	logger    *Logger
	requestID string
}

// Info logs an info message with request ID
func (rl *RequestLogger) Info(message string, data map[string]interface{}) {
	rl.logger.log(INFO, message, rl.requestID, data)
}

// Error logs an error message with request ID
func (rl *RequestLogger) Error(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}
	rl.logger.log(ERROR, message, rl.requestID, data)
}

// Global logger instance
var AppLogger *Logger

// InitLogger initializes the global logger
func InitLogger(minLevel LogLevel) {
	AppLogger = NewLogger(minLevel)
	AppLogger.Info("Logger initialized", map[string]interface{}{
		"min_level": minLevel,
	})
}
