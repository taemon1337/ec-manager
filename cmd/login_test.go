package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
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

// Mock IAM client
type mockIAMClient struct {
	GetUserOutput  *iam.GetUserOutput
	GetUserError   error
	ListRolesOutput *iam.ListRolesOutput
	ListRolesError  error
}

func (m *mockIAMClient) GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error) {
	return m.GetUserOutput, m.GetUserError
}

func (m *mockIAMClient) ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error) {
	return m.ListRolesOutput, m.ListRolesError
}

// Mock STS identity client
type mockSTSIdentityClient struct {
	GetCallerIdentityOutput *sts.GetCallerIdentityOutput
	GetCallerIdentityError  error
}

func (m *mockSTSIdentityClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.GetCallerIdentityOutput, m.GetCallerIdentityError
}

func TestLoginCmd(t *testing.T) {
	// Save original clients and restore after tests
	origNewSTSClient := newSTSClient
	origNewIAMClient := newIAMClient
	origNewSTSIdentityClient := newSTSIdentityClient
	defer func() { 
		newSTSClient = origNewSTSClient
		newIAMClient = origNewIAMClient
		newSTSIdentityClient = origNewSTSIdentityClient
	}()

	// Create a temporary directory for test credentials
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create AWS credentials directory
	credentialsDir := filepath.Join(tmpDir, ".aws")
	err := os.MkdirAll(credentialsDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create credentials directory: %v", err)
	}

	// Fixed time for testing
	testTime := time.Date(2024, 12, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		args           []string
		setupMocks     func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient)
		expectedError  bool
		errorContains  string
		validateInput  func(t *testing.T, input *sts.AssumeRoleInput)
		validateOutput func(t *testing.T, credentialsPath string)
	}{
		{
			name: "list roles success",
			args: []string{"--list-roles"},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				return nil, &mockIAMClient{
					GetUserOutput: &iam.GetUserOutput{
						User: &types.User{
							Arn: aws.String("arn:aws:iam::123456789012:user/testuser"),
						},
					},
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("TestRole1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole1"),
							},
							{
								RoleName: aws.String("TestRole2"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole2"),
							},
						},
					},
				}, nil
			},
			expectedError: false,
		},
		{
			name: "list roles with fallback to STS",
			args: []string{"--list-roles"},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				return nil, &mockIAMClient{
					GetUserError: fmt.Errorf("access denied"),
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("TestRole1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole1"),
							},
						},
					},
				}, &mockSTSIdentityClient{
					GetCallerIdentityOutput: &sts.GetCallerIdentityOutput{
						Account: aws.String("123456789012"),
					},
				}
			},
			expectedError: false,
		},
		{
			name: "no role ARN provided shows available roles",
			args: []string{},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				return nil, &mockIAMClient{
					GetUserOutput: &iam.GetUserOutput{
						User: &types.User{
							Arn: aws.String("arn:aws:iam::123456789012:user/testuser"),
						},
					},
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("TestRole1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole1"),
							},
						},
					},
				}, nil
			},
			expectedError: true,
			errorContains: "--role-arn is required",
		},
		{
			name: "successful login with basic role",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--profile", "test-profile",
			},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				expiration := testTime.Add(1 * time.Hour)
				stsClient := &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &ststypes.Credentials{
							AccessKeyId:     aws.String("TESTACCESSKEY"),
							SecretAccessKey: aws.String("TESTSECRETKEY"),
							SessionToken:    aws.String("TESTSESSIONTOKEN"),
							Expiration:      &expiration,
						},
					},
				}
				return stsClient, nil, nil
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
			expectedError: false,
		},
		{
			name: "successful login with MFA",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--profile", "mfa-profile",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/test-device",
				"--mfa-token", "123456",
			},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				expiration := testTime.Add(1 * time.Hour)
				stsClient := &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &ststypes.Credentials{
							AccessKeyId:     aws.String("MFAACCESSKEY"),
							SecretAccessKey: aws.String("MFASECRETKEY"),
							SessionToken:    aws.String("MFASESSIONTOKEN"),
							Expiration:      &expiration,
						},
					},
				}
				return stsClient, nil, nil
			},
			validateInput: func(t *testing.T, input *sts.AssumeRoleInput) {
				if input == nil {
					t.Fatal("AssumeRoleInput is nil")
				}
				assert.Equal(t, "arn:aws:iam::123456789012:role/TestRole", *input.RoleArn)
				assert.Equal(t, "arn:aws:iam::123456789012:mfa/test-device", *input.SerialNumber)
				assert.Equal(t, "123456", *input.TokenCode)
			},
			validateOutput: func(t *testing.T, credentialsPath string) {
				cfg, err := ini.Load(credentialsPath)
				assert.NoError(t, err)

				section := cfg.Section("mfa-profile")
				assert.Equal(t, "MFAACCESSKEY", section.Key("aws_access_key_id").String())
				assert.Equal(t, "MFASECRETKEY", section.Key("aws_secret_access_key").String())
				assert.Equal(t, "MFASESSIONTOKEN", section.Key("aws_session_token").String())
			},
			expectedError: false,
		},
		{
			name: "invalid MFA token",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/TestRole",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/test-device",
				"--mfa-token", "invalid",
			},
			setupMocks: func() (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				return &mockSTSClient{
					AssumeRoleError: fmt.Errorf("invalid MFA token"),
				}, nil, nil
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

			// Setup mocks if provided
			if tt.setupMocks != nil {
				stsClient, iamClient, stsIdentityClient := tt.setupMocks()
				if stsClient != nil {
					newSTSClient = func(cfg aws.Config) STSAPI {
						return stsClient
					}
				}
				if iamClient != nil {
					newIAMClient = func(cfg aws.Config) iamAPI {
						return iamClient
					}
				}
				if stsIdentityClient != nil {
					newSTSIdentityClient = func(cfg aws.Config) stsIdentityAPI {
						return stsIdentityClient
					}
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
			if tt.validateInput != nil {
				stsClient, _, _ := tt.setupMocks()
				if stsClient != nil {
					tt.validateInput(t, stsClient.AssumeRoleInput)
				}
			}

			// Validate output if needed
			if tt.validateOutput != nil {
				tt.validateOutput(t, filepath.Join(credentialsDir, "credentials"))
			}
		})
	}
}
