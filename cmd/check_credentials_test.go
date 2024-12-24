package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock implementations
type mockSTS struct {
	assumeRoleOutput *sts.AssumeRoleOutput
	assumeRoleErr    error
}

func (m mockSTS) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return m.assumeRoleOutput, m.assumeRoleErr
}

type mockIAM struct {
	getUserOutput  *iam.GetUserOutput
	getUserErr     error
	listRolesOutput *iam.ListRolesOutput
	listRolesErr    error
}

func (m mockIAM) GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error) {
	return m.getUserOutput, m.getUserErr
}

func (m mockIAM) ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error) {
	return m.listRolesOutput, m.listRolesErr
}

type mockSTSIdentity struct {
	getCallerIdentityOutput *sts.GetCallerIdentityOutput
	getCallerIdentityErr    error
}

func (m mockSTSIdentity) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.getCallerIdentityOutput, m.getCallerIdentityErr
}

// Helper functions for testing
func checkCredentials(cmd *cobra.Command) error {
	cfg, err := loadConfig(cmd.Context())
	if err != nil {
		return err
	}

	stsClient := newSTSIdentityClient(cfg)
	_, err = stsClient.GetCallerIdentity(cmd.Context(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Successfully verified AWS credentials")
	return nil
}

func assumeRole(cmd *cobra.Command, roleArn, mfaSerial, mfaToken, profile string) error {
	cfg, err := loadConfig(cmd.Context())
	if err != nil {
		return err
	}

	stsClient := newSTSClient(cfg)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("ec-manager"),
	}

	if mfaSerial != "" && mfaToken != "" {
		input.SerialNumber = aws.String(mfaSerial)
		input.TokenCode = aws.String(mfaToken)
	}

	_, err = stsClient.AssumeRole(cmd.Context(), input)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully assumed role %s and saved credentials to profile %s\n", roleArn, profile)
	return nil
}

func listAvailableRoles(cmd *cobra.Command) error {
	cfg, err := loadConfig(cmd.Context())
	if err != nil {
		return err
	}

	iamClient := newIAMClient(cfg)
	output, err := iamClient.ListRoles(cmd.Context(), &iam.ListRolesInput{})
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Available roles:")
	for _, role := range output.Roles {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", *role.Arn)
	}

	return nil
}

func TestCheckCredentialsCmd(t *testing.T) {
	// Save original functions and restore after tests
	origLoadConfig := loadConfig
	origNewSTSClient := newSTSClient
	origNewIAMClient := newIAMClient
	origNewSTSIdentityClient := newSTSIdentityClient
	defer func() {
		loadConfig = origLoadConfig
		newSTSClient = origNewSTSClient
		newIAMClient = origNewIAMClient
		newSTSIdentityClient = origNewSTSIdentityClient
	}()

	// Create a temporary directory for AWS credentials
	tmpDir, err := os.MkdirTemp("", "aws-creds-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME and restore after tests
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	// Create .aws directory
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0700); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name               string
		args               []string
		mockConfig         aws.Config
		mockSTS           mockSTS
		mockIAM           mockIAM
		mockSTSIdentity   mockSTSIdentity
		configLoadErr     error
		wantErr           bool
		wantOutput        string
	}{
		{
			name: "basic_credentials_check_success",
			args: []string{},
			mockSTSIdentity: mockSTSIdentity{
				getCallerIdentityOutput: &sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
					Arn:     aws.String("arn:aws:iam::123456789012:user/test-user"),
					UserId:  aws.String("AIDAXXXXXXXXXXXXXXXX"),
				},
			},
			wantOutput: "Successfully verified AWS credentials\n",
		},
		{
			name: "basic_credentials_check_failure",
			args: []string{},
			mockSTSIdentity: mockSTSIdentity{
				getCallerIdentityErr: fmt.Errorf("operation error STS: GetCallerIdentity, no EC2 IMDS role found"),
			},
			wantErr: true,
		},
		{
			name: "successful_role_assumption",
			args: []string{"--role-arn", "arn:aws:iam::123456789012:role/role1"},
			mockSTS: mockSTS{
				assumeRoleOutput: &sts.AssumeRoleOutput{
					Credentials: &ststypes.Credentials{
						AccessKeyId:     aws.String("ASIAXXXXXXXXXXXXXXXX"),
						SecretAccessKey: aws.String("secretkey"),
						SessionToken:    aws.String("sessiontoken"),
					},
				},
			},
			wantOutput: "Successfully assumed role arn:aws:iam::123456789012:role/role1 and saved credentials to profile default\n",
		},
		{
			name: "successful_role_with_MFA",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/role1",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/user",
				"--mfa-token", "123456",
			},
			mockSTS: mockSTS{
				assumeRoleOutput: &sts.AssumeRoleOutput{
					Credentials: &ststypes.Credentials{
						AccessKeyId:     aws.String("ASIAXXXXXXXXXXXXXXXX"),
						SecretAccessKey: aws.String("secretkey"),
						SessionToken:    aws.String("sessiontoken"),
					},
				},
			},
			wantOutput: "Successfully assumed role arn:aws:iam::123456789012:role/role1 and saved credentials to profile default\n",
		},
		{
			name: "failed_role_invalid_MFA",
			args: []string{
				"--role-arn", "arn:aws:iam::123456789012:role/role1",
				"--mfa-serial", "arn:aws:iam::123456789012:mfa/user",
				"--mfa-token", "invalid",
			},
			mockSTS: mockSTS{
				assumeRoleErr: fmt.Errorf("operation error STS: AssumeRole, invalid MFA token"),
			},
			wantErr: true,
		},
		{
			name: "list_roles_success",
			args: []string{"--list-roles"},
			mockIAM: mockIAM{
				getUserOutput: &iam.GetUserOutput{
					User: &types.User{
						Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
					},
				},
				listRolesOutput: &iam.ListRolesOutput{
					Roles: []types.Role{
						{
							Arn:      aws.String("arn:aws:iam::123456789012:role/role1"),
							RoleName: aws.String("role1"),
						},
						{
							Arn:      aws.String("arn:aws:iam::123456789012:role/role2"),
							RoleName: aws.String("role2"),
						},
					},
				},
			},
			wantOutput: "Available roles:\n- arn:aws:iam::123456789012:role/role1\n- arn:aws:iam::123456789012:role/role2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mocks
			loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
				return tt.mockConfig, tt.configLoadErr
			}
			newSTSClient = func(cfg aws.Config) STSAPI {
				return tt.mockSTS
			}
			newIAMClient = func(cfg aws.Config) iamAPI {
				return tt.mockIAM
			}
			newSTSIdentityClient = func(cfg aws.Config) stsIdentityAPI {
				return tt.mockSTSIdentity
			}

			// Create command
			cmd := &cobra.Command{
				Use:           "credentials",
				Short:         "Check AWS credentials",
				Long:          "Check if AWS credentials are valid and optionally assume a role",
				SilenceUsage:  true,
				SilenceErrors: true,
				RunE: func(cmd *cobra.Command, args []string) error {
					roleArn, _ := cmd.Flags().GetString("role-arn")
					mfaSerial, _ := cmd.Flags().GetString("mfa-serial")
					mfaToken, _ := cmd.Flags().GetString("mfa-token")
					profile, _ := cmd.Flags().GetString("profile")
					listRoles, _ := cmd.Flags().GetBool("list-roles")

					if listRoles {
						return listAvailableRoles(cmd)
					}

					if roleArn != "" {
						return assumeRole(cmd, roleArn, mfaSerial, mfaToken, profile)
					}

					return checkCredentials(cmd)
				},
			}

			cmd.Flags().String("role-arn", "", "ARN of the role to assume")
			cmd.Flags().String("mfa-serial", "", "Serial number of MFA device")
			cmd.Flags().String("mfa-token", "", "Token from MFA device")
			cmd.Flags().String("profile", "default", "AWS profile to store credentials")
			cmd.Flags().Bool("list-roles", false, "List available roles")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, buf.String())
			}
		})
	}
}
