package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/mock"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// EC2ClientKey is the context key for the EC2 client
	EC2ClientKey ContextKey = "ec2_client"
)

// SetupMockEC2Client sets up a mock EC2 client for testing
func SetupMockEC2Client() *mock.MockEC2Client {
	return mock.NewMockEC2Client()
}

// GetTestContext returns a context with a mock EC2 client
func GetTestContext() context.Context {
	return context.WithValue(context.Background(), EC2ClientKey, SetupMockEC2Client())
}

// GetTestContextWithClient returns a context with the given mock EC2 client
func GetTestContextWithClient(client *mock.MockEC2Client) context.Context {
	return context.WithValue(context.Background(), EC2ClientKey, client)
}

// SetupTestCommand sets up a test command with the given arguments
func SetupTestCommand(cmd *cobra.Command, args []string) error {
	if cmd.Context() == nil {
		cmd.SetContext(GetTestContext())
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

// GetEC2Client retrieves the EC2 client from the context
func GetEC2Client(ctx context.Context) *mock.MockEC2Client {
	if client, ok := ctx.Value(EC2ClientKey).(*mock.MockEC2Client); ok {
		return client
	}
	return nil
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
