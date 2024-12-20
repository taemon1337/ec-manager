package testutil

import (
	"bytes"
	"testing"

	"github.com/taemon1337/ami-migrate/pkg/logger"
)

// InitTestLogger initializes a test logger that writes to a buffer.
// This is safe to call multiple times and from multiple tests.
func InitTestLogger(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	logger.Init(logger.LogLevel("debug"))

	t.Cleanup(func() {
		logger.Reset()
	})

	return &buf
}
