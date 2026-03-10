package log

import (
	"strings"

	"github.com/owen/xatu/internal/aws"
)

// ParseLevel extracts the log level from a message string.
func ParseLevel(message string) aws.LogLevel {
	upper := strings.ToUpper(message)

	// Check common patterns: level at the start, or after a timestamp, or as a JSON field
	for _, pattern := range []struct {
		keyword string
		level   aws.LogLevel
	}{
		{"FATAL", aws.LevelFatal},
		{"ERROR", aws.LevelError},
		{"WARN", aws.LevelWarn},
		{"INFO", aws.LevelInfo},
		{"DEBUG", aws.LevelDebug},
	} {
		if strings.Contains(upper, pattern.keyword) {
			return pattern.level
		}
	}

	return aws.LevelUnknown
}

// ExtractKV extracts key=value pairs from a log message.
func ExtractKV(message string) map[string]string {
	attrs := make(map[string]string)
	parts := strings.Fields(message)

	for _, part := range parts {
		if idx := strings.Index(part, "="); idx > 0 && idx < len(part)-1 {
			key := part[:idx]
			value := strings.Trim(part[idx+1:], "\"'")
			attrs[key] = value
		}
	}

	return attrs
}
