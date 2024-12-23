package testutil

import (
	"bytes"
	"testing"

	"github.com/taemon1337/ec-manager/pkg/logger"
)

// InitTestLogger initializes a logger for testing with debug level
func InitTestLogger() {
	logger.Init(logger.DebugLevel)
}

// SetupTestLogger sets up a test logger that writes to a buffer with info level
func SetupTestLogger(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	logger.Init(logger.InfoLevel)
	return &buf
}
