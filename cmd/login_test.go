package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock STS client
type mockSTSClient struct {
	AssumeRoleOutput *sts.AssumeRoleOutput
	AssumeRoleError  error
	AssumeRoleInput  *sts.AssumeRoleInput
}

func (m *mockSTSClient) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	// Store the input for validation
	m.AssumeRoleInput = params

	if m.AssumeRoleError != nil {
		return nil, m.AssumeRoleError
	}
	return m.AssumeRoleOutput, nil
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
	Output *sts.GetCallerIdentityOutput
	Error  error
}

func (m *mockSTSIdentityClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.Output, m.Error
}

func TestLoginCmd(t *testing.T) {
	// Store original client creation functions
	origNewSTSClient := newSTSClient
	origNewIAMClient := newIAMClient
	origNewSTSIdentityClient := newSTSIdentityClient
	origLoadConfig := loadConfig
	origHome := os.Getenv("HOME")

	// Restore original functions after all tests
	defer func() {
		newSTSClient = origNewSTSClient
		newIAMClient = origNewIAMClient
		newSTSIdentityClient = origNewSTSIdentityClient
		loadConfig = origLoadConfig
		os.Setenv("HOME", origHome)
	}()

	// Create temporary AWS directory for testing
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create AWS config directory
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0700); err != nil {
		t.Fatalf("Failed to create AWS directory: %v", err)
	}

	// Mock AWS config loading
	loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{
			Region: "us-west-2",
			Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "AKIADEFAULT",
					SecretAccessKey: "defaultsecret",
				}, nil
			}),
		}, nil
	}

	// Set test time
	testTime := time.Date(2024, 12, 20, 19, 19, 3, 0, time.UTC)

	tests := []struct {
		name       string
		args       []string
		setupMocks func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "list_roles_success",
			args: []string{"--list-roles"},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{}
				mockIAM := &mockIAMClient{
					GetUserOutput: &iam.GetUserOutput{
						User: &types.User{
							Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
						},
					},
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("role1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role1"),
							},
							{
								RoleName: aws.String("role2"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role2"),
							},
						},
					},
				}
				mockSTSIdentity := &mockSTSIdentityClient{}
				return mockSTS, mockIAM, mockSTSIdentity
			},
		},
		{
			name: "list_roles_with_fallback_to_STS",
			args: []string{"--list-roles"},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{}
				mockIAM := &mockIAMClient{
					GetUserError: fmt.Errorf("access denied"),
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("role1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role1"),
							},
							{
								RoleName: aws.String("role2"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role2"),
							},
						},
					},
				}
				mockSTSIdentity := &mockSTSIdentityClient{
					Output: &sts.GetCallerIdentityOutput{
						Account: aws.String("123456789012"),
					},
				}
				return mockSTS, mockIAM, mockSTSIdentity
			},
		},
		{
			name: "no_role_ARN_provided_shows_available_roles",
			args: []string{},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{}
				mockIAM := &mockIAMClient{
					GetUserOutput: &iam.GetUserOutput{
						User: &types.User{
							Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
						},
					},
					ListRolesOutput: &iam.ListRolesOutput{
						Roles: []types.Role{
							{
								RoleName: aws.String("role1"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role1"),
							},
							{
								RoleName: aws.String("role2"),
								Arn:      aws.String("arn:aws:iam::123456789012:role/role2"),
							},
						},
					},
				}
				mockSTSIdentity := &mockSTSIdentityClient{}
				return mockSTS, mockIAM, mockSTSIdentity
			},
			wantErr: true,
			errMsg:  "--role-arn is required. Please select one of the roles above",
		},
		{
			name: "successful_login_with_basic_role",
			args: []string{"--profile", "test", "--role-arn", "arn:aws:iam::123456789012:role/role1"},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &ststypes.Credentials{
							AccessKeyId:     aws.String("AKIATEST"),
							SecretAccessKey: aws.String("testsecret"),
							SessionToken:    aws.String("testtoken"),
							Expiration:      aws.Time(testTime),
						},
					},
				}
				mockIAM := &mockIAMClient{}
				mockSTSIdentity := &mockSTSIdentityClient{}

				// Set up the expected AssumeRoleInput
				mockSTS.AssumeRoleInput = &sts.AssumeRoleInput{
					RoleArn:         aws.String("arn:aws:iam::123456789012:role/role1"),
					RoleSessionName: aws.String("ec-manager-session"),
				}

				return mockSTS, mockIAM, mockSTSIdentity
			},
		},
		{
			name: "successful_login_with_MFA",
			args: []string{"--profile", "test", "--role-arn", "arn:aws:iam::123456789012:role/role1", "--mfa-token", "123456"},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{
					AssumeRoleOutput: &sts.AssumeRoleOutput{
						Credentials: &ststypes.Credentials{
							AccessKeyId:     aws.String("AKIATEST"),
							SecretAccessKey: aws.String("testsecret"),
							SessionToken:    aws.String("testtoken"),
							Expiration:      aws.Time(testTime),
						},
					},
				}
				mockIAM := &mockIAMClient{}
				mockSTSIdentity := &mockSTSIdentityClient{
					Output: &sts.GetCallerIdentityOutput{
						Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
					},
				}

				// Set up the expected AssumeRoleInput with MFA
				mockSTS.AssumeRoleInput = &sts.AssumeRoleInput{
					RoleArn:         aws.String("arn:aws:iam::123456789012:role/role1"),
					RoleSessionName: aws.String("ec-manager-session"),
					SerialNumber:    aws.String("arn:aws:iam::123456789012:mfa/test-user"),
					TokenCode:       aws.String("123456"),
				}

				return mockSTS, mockIAM, mockSTSIdentity
			},
		},
		{
			name: "invalid_MFA_token",
			args: []string{"--profile", "test", "--role-arn", "arn:aws:iam::123456789012:role/role1", "--mfa-token", "123456"},
			setupMocks: func(t *testing.T) (*mockSTSClient, *mockIAMClient, *mockSTSIdentityClient) {
				mockSTS := &mockSTSClient{
					AssumeRoleError: fmt.Errorf("invalid MFA token"),
				}
				mockIAM := &mockIAMClient{}
				mockSTSIdentity := &mockSTSIdentityClient{
					Output: &sts.GetCallerIdentityOutput{
						Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
					},
				}
				return mockSTS, mockIAM, mockSTSIdentity
			},
			wantErr: true,
			errMsg:  "invalid MFA token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary AWS directory for testing
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)

			// Create AWS config directory
			awsDir = filepath.Join(tmpDir, ".aws")
			if err := os.MkdirAll(awsDir, 0700); err != nil {
				t.Fatalf("Failed to create AWS directory: %v", err)
			}

			// Setup mocks if provided
			if tt.setupMocks != nil {
				mockSTS, mockIAM, mockSTSIdentity := tt.setupMocks(t)
				newSTSClient = func(cfg aws.Config) STSAPI {
					return mockSTS
				}
				newIAMClient = func(cfg aws.Config) iamAPI {
					return mockIAM
				}
				newSTSIdentityClient = func(cfg aws.Config) stsIdentityAPI {
					return mockSTSIdentity
				}
			}

			// Create command
			cmd := NewLoginCmd()
			cmd.SilenceUsage = true

			// Set command arguments
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
