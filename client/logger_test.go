package client

import (
	"bytes"
	"io"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	t.Run("should create default logger", func(t *testing.T) {
		createDefaultLogger()
		if logger == nil {
			t.Fatal("expected logger to be created")
		}
	})

	t.Run("should set logger", func(t *testing.T) {
		testLogger := log.New(io.Discard, "", 0)
		SetLogger(testLogger)

		if logger != testLogger {
			t.Fatalf("expected logger to be set to %v, got %v", testLogger, logger)
		}
	})

	t.Run("should log", func(t *testing.T) {
		writer := bytes.NewBuffer([]byte{})
		SetLogger(log.New(writer, "", 0))
		logger.Printf("test %s", "log")

		if writer.String() != "test log\n" {
			t.Fatalf("expected log to be %s, got %s", "test log\n", writer.String())
		}
	})
}
