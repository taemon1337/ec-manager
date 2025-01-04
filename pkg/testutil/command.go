package testutil

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	mockclient "github.com/taemon1337/ec-manager/pkg/mock"
)

// CommandTestCase represents a test case for a cobra command
type CommandTestCase struct {
	Name         string
	Args         []string
	WantErr      bool
	ErrContains  string
	MockEC2Setup func(*mockclient.MockEC2Client)
	MockSTSSetup func(*mockclient.MockSTSClient)
	MockIAMSetup func(*mockclient.MockIAMClient)
	SetupContext func(context.Context) context.Context
}

// RunCommandTest runs a command test with the given test cases
func RunCommandTest(t *testing.T, newCmd func() *cobra.Command, tests []CommandTestCase) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Create base context
			ctx := context.Background()

			// Allow custom context setup first
			if tt.SetupContext != nil {
				ctx = tt.SetupContext(ctx)
			}

			// Get mock clients from context or create new ones
			var mockEC2Client *mockclient.MockEC2Client
			var mockSTSClient *mockclient.MockSTSClient
			var mockIAMClient *mockclient.MockIAMClient

			if existingEC2Client := ctx.Value(mockclient.EC2ClientKey); existingEC2Client != nil {
				mockEC2Client = existingEC2Client.(*mockclient.MockEC2Client)
			} else {
				mockEC2Client = mockclient.NewMockEC2Client(t)
				ctx = context.WithValue(ctx, mockclient.EC2ClientKey, mockEC2Client)
			}

			if existingSTSClient := ctx.Value(mockclient.STSClientKey); existingSTSClient != nil {
				mockSTSClient = existingSTSClient.(*mockclient.MockSTSClient)
			} else {
				mockSTSClient = mockclient.NewMockSTSClient(t)
				ctx = context.WithValue(ctx, mockclient.STSClientKey, mockSTSClient)
			}

			if existingIAMClient := ctx.Value(mockclient.IAMClientKey); existingIAMClient != nil {
				mockIAMClient = existingIAMClient.(*mockclient.MockIAMClient)
			} else {
				mockIAMClient = mockclient.NewMockIAMClient(t)
				ctx = context.WithValue(ctx, mockclient.IAMClientKey, mockIAMClient)
			}

			// Set up mocks
			if tt.MockEC2Setup != nil {
				tt.MockEC2Setup(mockEC2Client)
			}
			if tt.MockSTSSetup != nil {
				tt.MockSTSSetup(mockSTSClient)
			}
			if tt.MockIAMSetup != nil {
				tt.MockIAMSetup(mockIAMClient)
			}

			cmd := newCmd()
			cmd.SetContext(ctx)

			// Parse flags
			cmd.SetArgs(tt.Args)
			err := cmd.Execute()

			if tt.WantErr {
				assert.Error(t, err)
				if tt.ErrContains != "" {
					assert.Contains(t, err.Error(), tt.ErrContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockEC2Client.AssertExpectations(t)
			mockSTSClient.AssertExpectations(t)
			mockIAMClient.AssertExpectations(t)
		})
	}
}
