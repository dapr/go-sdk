package client

import (
	"io"
	"log"
	"os"
)

const (
	daprLogLevelEnvVarName = "DAPR_LOG_LEVEL"
)

var (
	logger Logger
)

// Logger is the interface for Dapr client logger implementation.
// This interface is compatible with the standard library logger.
type Logger interface {
	Printf(format string, v ...any)
	Print(v ...any)
	Println(v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Fatalln(v ...any)
	Panic(v ...any)
	Panicf(format string, v ...any)
	Panicln(v ...any)
}

// SetLogger sets the logger for the dapr client.
func SetLogger(l Logger) {
	logger = l
}

// createDefaultLogger creates a default logger for the dapr client.
// The logger is set to stdout by default, but can be disabled by setting DAPR_LOG_LEVEL to "production".
func createDefaultLogger() {
	var logWriter io.Writer = os.Stdout
	if os.Getenv(daprLogLevelEnvVarName) == "production" {
		logWriter = io.Discard
	}

	logger = log.New(logWriter, "", 0)
}
