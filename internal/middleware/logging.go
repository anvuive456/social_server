package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggingConfig struct {
	Output    io.Writer
	Format    string
	SkipPaths []string
}

func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Output: os.Stdout,
		Format: "[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
	}
}

func Logging(config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}

			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}

			return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage,
			)
		},
		Output:    config.Output,
		SkipPaths: config.SkipPaths,
	})
}

func CustomLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.BodySize,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.Latency,
		)
	})
}

func JSONLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","ip":"%s","user_agent":"%s","error":"%s"}%s`,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
			param.ErrorMessage,
			"\n",
		)
	})
}

func LoggingWithSkipPaths(skipPaths []string) gin.HandlerFunc {
	config := DefaultLoggingConfig()
	config.SkipPaths = skipPaths
	return Logging(config)
}