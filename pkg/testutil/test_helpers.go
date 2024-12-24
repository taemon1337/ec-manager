package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/mock"
)

// SetupMockEC2Client sets up a mock EC2 client for testing
func SetupMockEC2Client() *mock.MockEC2Client {
	return mock.NewMockEC2Client()
}

// GetMockClient returns a client with mock EC2 client
func GetMockClient() *client.Client {
	cfg := &client.Config{
		MockMode: true,
	}
	client, err := client.NewClient(cfg)
	if err != nil {
		return nil
	}
	return client
}

// CleanupMockEC2Client cleans up the mock EC2 client
func CleanupMockEC2Client() {
	// Nothing to clean up for now
}

// SetupTestCommand sets up a test command with the given arguments
func SetupTestCommand(cmd *cobra.Command, args []string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}

// GetEC2Client retrieves the EC2 client from the command context
func GetEC2Client(cmd *cobra.Command) ec2.DescribeInstancesAPIClient {
	// Get the client from the command context
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Create a new client with mock mode
	cfg := &client.Config{
		MockMode: true,
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		return nil
	}

	return c.GetEC2Client()
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
