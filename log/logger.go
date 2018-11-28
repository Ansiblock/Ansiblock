package log

import (
	"log"
	"sync"

	"go.uber.org/zap"
)

var logger = zap.NewNop()
var once sync.Once

// Init initializes a thread-safe singleton logger
// This would be called from a main method when the application starts up
// This function would ideally, take zap configuration, but is left out
// in favor of simplicity using the example logger.
func Init() {
	once.Do(func() {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			log.Fatal("Can't create zap logger")
		}
	})
}

// Debug logs a debug message with the given fields
func Debug(message string, fields ...zap.Field) {
	logger.Debug(message, fields...)
}

// Info logs a debug message with the given fields
func Info(message string, fields ...zap.Field) {
	logger.Info(message, fields...)
}

// Warn logs a debug message with the given fields
func Warn(message string, fields ...zap.Field) {
	logger.Warn(message, fields...)
}

// Error logs a debug message with the given fields
func Error(message string, fields ...zap.Field) {
	logger.Error(message, fields...)
}

// Fatal logs a message than calls os.Exit(1)
func Fatal(message string, fields ...zap.Field) {
	logger.Fatal(message, fields...)
}

// Panic logs a message at PanicLevel and them panics
func Panic(message string, fields ...zap.Field) {
	logger.Panic(message, fields...)
}
