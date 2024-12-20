package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

// Mock STS client
type mockSTSClient struct {
	AssumeRoleOutput *sts.AssumeRoleOutput
	AssumeRoleError  error
	AssumeRoleInput  *sts.AssumeRoleInput
}

func (m *mockSTSClient) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	m.AssumeRoleInput = params
	return m.AssumeRoleOutput, m.AssumeRoleError
}

func TestLoginCmd(t *testing.T) {
	// Save original newSTSClient and restore after tests
	origNewSTSClient := newSTSClient
	defer func() { newSTSClient = origNewSTSClient }()

	// Create a temporary directory for test credentials
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Fixed time for testing
	testTime := time.Date(2024, 12, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		args           []string
		setupMock      func() *mockSTSClient
		expectedError  bool
		errorContains  string
		validateInput  func(t *testing.T, input *sts.AssumeRoleInput)
		validateOutput func(t *testing.T, credentialsPath string)
	}{
		{
			name: "successful login with basic role",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--profile", "test-profile",
			},
			setupMock: func() *mockSTSClient {
				expiration := testTime.Add(1 * time.Hour)
				return &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &types.Credentials{
							AccessKeyId:     aws.String("TESTACCESSKEY"),
							SecretAccessKey: aws.String("TESTSECRETKEY"),
							SessionToken:    aws.String("TESTSESSIONTOKEN"),
							Expiration:      &expiration,
						},
					},
				}
			},
			validateInput: func(t *testing.T, input *sts.AssumeRoleInput) {
				if input == nil {
					t.Fatal("AssumeRoleInput is nil")
				}
				assert.Equal(t, "arn:aws:iam::123456789012:role/TestRole", *input.RoleArn)
				assert.Equal(t, int32(3600), *input.DurationSeconds)
				assert.Equal(t, "ec-manager-session", *input.RoleSessionName)
			},
			validateOutput: func(t *testing.T, credentialsPath string) {
				cfg, err := ini.Load(credentialsPath)
				assert.NoError(t, err)

				section := cfg.Section("test-profile")
				assert.Equal(t, "TESTACCESSKEY", section.Key("aws_access_key_id").String())
				assert.Equal(t, "TESTSECRETKEY", section.Key("aws_secret_access_key").String())
				assert.Equal(t, "TESTSESSIONTOKEN", section.Key("aws_session_token").String())
			},
		},
		{
			name: "successful login with MFA",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--profile", "mfa-profile",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/test-device",
				"--mfa-token", "123456",
			},
			setupMock: func() *mockSTSClient {
				expiration := testTime.Add(1 * time.Hour)
				return &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &types.Credentials{
							AccessKeyId:     aws.String("MFAACCESSKEY"),
							SecretAccessKey: aws.String("MFASECRETKEY"),
							SessionToken:    aws.String("MFASESSIONTOKEN"),
							Expiration:      &expiration,
						},
					},
				}
			},
			validateInput: func(t *testing.T, input *sts.AssumeRoleInput) {
				if input == nil {
					t.Fatal("AssumeRoleInput is nil")
				}
				assert.Equal(t, "arn:aws:iam::123456789012:mfa/test-device", *input.SerialNumber)
				assert.Equal(t, "123456", *input.TokenCode)
			},
			validateOutput: func(t *testing.T, credentialsPath string) {
				cfg, err := ini.Load(credentialsPath)
				assert.NoError(t, err)

				section := cfg.Section("mfa-profile")
				assert.Equal(t, "MFAACCESSKEY", section.Key("aws_access_key_id").String())
			},
		},
		{
			name: "missing role arn",
			args: []string{
				"--profile", "test-profile",
			},
			expectedError: true,
			errorContains: "required flag(s) \"role-arn\" not set",
		},
		{
			name: "invalid MFA token",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/test-device",
				"--mfa-token", "invalid",
			},
			setupMock: func() *mockSTSClient {
				return &mockSTSClient{
					AssumeRoleError: fmt.Errorf("invalid MFA token"),
				}
			},
			expectedError: true,
			errorContains: "invalid MFA token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := NewLoginCmd()

			// Set args
			cmd.SetArgs(tt.args)

			// Create credentials directory
			credentialsDir := filepath.Join(tmpDir, ".aws")
			os.MkdirAll(credentialsDir, 0700)
			credentialsPath := filepath.Join(credentialsDir, "credentials")

			// Setup mock if provided
			var mockClient *mockSTSClient
			if tt.setupMock != nil {
				mockClient = tt.setupMock()
				// Override the newSTSClient function
				newSTSClient = func(cfg aws.Config) STSAPI {
					return mockClient // mockSTSClient implements STSAPI
				}
			}

			// Run command
			err := cmd.Execute()

			// Validate error
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			assert.NoError(t, err)

			// Validate input if needed
			if tt.validateInput != nil && mockClient != nil {
				tt.validateInput(t, mockClient.AssumeRoleInput)
			}

			// Validate output if needed
			if tt.validateOutput != nil {
				tt.validateOutput(t, credentialsPath)
			}
		})
	}
}
