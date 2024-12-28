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

	// Create a new root command
	rootCmd := &cobra.Command{Use: "root"}
	
	// Create the check command
	checkCmd := &cobra.Command{Use: "check"}
	rootCmd.AddCommand(checkCmd)

	// Create the migrate subcommand
	checkMigrateCmd := &cobra.Command{
		Use: "migrate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mockSvc.ListUserInstances(cmd.Context())
		},
	}
	checkCmd.AddCommand(checkMigrateCmd)

	t.Run("success", func(t *testing.T) {
		mockSvc.mockError = nil
		rootCmd.SetOut(&buf)
		rootCmd.SetArgs([]string{"check", "migrate"})
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc.mockError = fmt.Errorf("API error")
		rootCmd.SetOut(&buf)
		rootCmd.SetArgs([]string{"check", "migrate"})
		err := rootCmd.Execute()
		assert.Error(t, err)
	})
}
