package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

// SetupTestCommand is a helper function for setting up test commands
func SetupTestCommand(cmd *cobra.Command, args []string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}

// SetupTestCommandWithContext is a helper function for setting up test commands with context
func SetupTestCommandWithContext(ctx context.Context, cmd *cobra.Command, args []string) error {
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// SetupTestCommandWithMockClient is a helper function for setting up test commands with a mock client
func SetupTestCommandWithMockClient(cmd *cobra.Command, args []string, mockClient ecTypes.EC2Client) error {
	client.SetMockMode(true)
	client.SetMockClient(mockClient)
	defer func() {
		client.SetMockMode(false)
		client.SetMockClient(nil)
	}()
	
	cmd.SetArgs(args)
	return cmd.Execute()
}

// AssertErrorContains checks if an error message contains a substring
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", substr)
		return
	}
	msg := err.Error()
	if !strings.Contains(msg, substr) {
		t.Errorf("expected error containing %q, got %q", substr, msg)
	}
}
