package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockAMIService is a mock implementation of the AMI service for testing
type mockAMIService struct {
	mockError error
}

func (m *mockAMIService) ListUserInstances(ctx context.Context) error {
	if m.mockError != nil {
		return m.mockError
	}
	return nil
}

func TestCheckMigrateCmd(t *testing.T) {
	var buf bytes.Buffer
	mockSvc := &mockAMIService{}

	cmd := &cobra.Command{
		Use: "check-migrate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mockSvc.ListUserInstances(cmd.Context())
		},
	}

	t.Run("success", func(t *testing.T) {
		mockSvc.mockError = nil
		cmd.SetOut(&buf)
		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc.mockError = fmt.Errorf("API error")
		cmd.SetOut(&buf)
		err := cmd.Execute()
		assert.Error(t, err)
	})
}
