package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// STSAPI interface for mocking
type STSAPI interface {
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// iamAPI interface for mocking
type iamAPI interface {
	GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error)
	ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error)
}

// stsIdentityAPI interface for mocking
type stsIdentityAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

var (
	loadConfig   = config.LoadDefaultConfig
	newSTSClient = func(cfg aws.Config) STSAPI {
		return sts.NewFromConfig(cfg)
	}
	newIAMClient = func(cfg aws.Config) iamAPI {
		return iam.NewFromConfig(cfg)
	}
	newSTSIdentityClient = func(cfg aws.Config) stsIdentityAPI {
		return sts.NewFromConfig(cfg)
	}
)

// isCredentialError checks if the error is related to missing or invalid credentials
func isCredentialError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	credentialErrors := []string{
		"no EC2 IMDS role found",
		"get credentials",
		"failed to refresh cached credentials",
		"operation error STS: GetCallerIdentity",
		"InvalidClientTokenId",
		"ExpiredToken",
		"AccessDenied",
	}

	for _, credErr := range credentialErrors {
		if strings.Contains(errStr, credErr) {
			return true
		}
	}
	return false
}

// getCredentialHelp returns a helpful message for credential setup
func getCredentialHelp() string {
	return `
No valid AWS credentials found. To fix this:

1. Set AWS credentials using environment variables:
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=your_region (e.g., us-east-1)

2. Or configure credentials in ~/.aws/credentials:
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   region = your_region

3. Or if running on EC2, ensure the instance has an IAM role attached

For more information, visit: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html`
}

// discoverRoleARN attempts to discover available roles for the user
func discoverRoleARN(ctx context.Context, cfg aws.Config) ([]string, error) {
	var roleARNs []string

	// Create IAM client
	iamClient := newIAMClient(cfg)

	// Try to get current user info
	user, err := iamClient.GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		// Fallback to STS GetCallerIdentity if IAM access is restricted
		stsIdentityClient := newSTSIdentityClient(cfg)
		identity, err := stsIdentityClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to get caller identity: %w", err)
		}

		// Extract account ID from identity
		accountID := *identity.Account

		// List roles (if we have permission)
		roles, err := iamClient.ListRoles(ctx, &iam.ListRolesInput{})
		if err == nil {
			for _, role := range roles.Roles {
				if role.RoleName != nil && role.Arn != nil {
					roleARNs = append(roleARNs, *role.Arn)
				}
			}
		}

		if len(roleARNs) == 0 {
			// If we can't list roles, at least we have the account ID
			return nil, fmt.Errorf("could not list roles. Your AWS Account ID is: %s\nUse this to construct your role ARN: arn:aws:iam::%s:role/YOUR_ROLE_NAME", accountID, accountID)
		}

		return roleARNs, nil
	}

	// We have user info, get account ID from ARN
	accountID := strings.Split(*user.User.Arn, ":")[4]

	// List roles
	roles, err := iamClient.ListRoles(ctx, &iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("could not list roles. Your AWS Account ID is: %s\nUse this to construct your role ARN: arn:aws:iam::%s:role/YOUR_ROLE_NAME", accountID, accountID)
	}

	// Collect role ARNs
	for _, role := range roles.Roles {
		if role.RoleName != nil && role.Arn != nil {
			roleARNs = append(roleARNs, *role.Arn)
		}
	}

	return roleARNs, nil
}

var checkCredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Check AWS credentials and optionally assume a role",
	Long: `Check AWS credentials and optionally assume a role. You can either:
1. Verify direct credentials (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY)
2. Assume an IAM role using --role-arn (optional)

The credentials will be stored in ~/.aws/credentials under the specified profile.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// For listing roles, we'll use the default credential chain
		if listRoles {
			// If in mock mode, return mock roles
			if mockMode {
				fmt.Fprintln(cmd.OutOrStdout(), "Available roles (mock):")
				fmt.Fprintln(cmd.OutOrStdout(), "- arn:aws:iam::123456789012:role/mock-role-1")
				fmt.Fprintln(cmd.OutOrStdout(), "- arn:aws:iam::123456789012:role/mock-role-2")
				return nil
			}

			// Use LoadDefaultConfig which will try environment variables, shared credentials, etc.
			cfg, err := loadConfig(ctx)
			if err != nil {
				if isCredentialError(err) {
					return fmt.Errorf("failed to load AWS credentials\n%s", getCredentialHelp())
				}
				return fmt.Errorf("unable to load AWS config: %w", err)
			}

			roles, err := discoverRoleARN(ctx, cfg)
			if err != nil {
				if isCredentialError(err) {
					return fmt.Errorf("failed to list roles: no valid AWS credentials found\n%s", getCredentialHelp())
				}
				return fmt.Errorf("failed to list roles: %w\nMake sure you have valid AWS credentials with IAM:ListRoles permission", err)
			}

			if len(roles) == 0 {
				return fmt.Errorf("no roles found. Make sure you have valid AWS credentials with IAM:ListRoles permission")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Available roles:")
			for _, role := range roles {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", role)
			}
			return nil
		}

		// If in mock mode, return success
		if mockMode {
			if roleArn != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Successfully assumed role %s (mock)\n", roleArn)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Successfully verified AWS credentials (mock)")
			}
			return nil
		}

		// Load config for authentication
		cfg, err := loadConfig(ctx)
		if err != nil {
			if isCredentialError(err) {
				return fmt.Errorf("failed to load AWS credentials\n%s", getCredentialHelp())
			}
			return fmt.Errorf("unable to load AWS config: %w", err)
		}

		// If role ARN is provided, assume the role
		if roleArn != "" {
			// Create STS client using the mockable function
			stsClient := newSTSClient(cfg)

			// Prepare AssumeRole input
			input := &sts.AssumeRoleInput{
				RoleArn:         &roleArn,
				RoleSessionName: &sessionName,
				DurationSeconds: &duration,
			}

			// If MFA is provided, add it to the request
			if mfaSerial != "" && mfaToken != "" {
				input.SerialNumber = &mfaSerial
				input.TokenCode = &mfaToken
			}

			// Call AssumeRole
			result, err := stsClient.AssumeRole(ctx, input)
			if err != nil {
				return fmt.Errorf("failed to assume role: %w", err)
			}

			// Get AWS credentials file path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("unable to get home directory: %w", err)
			}
			credentialsPath := filepath.Join(homeDir, ".aws", "credentials")

			// Ensure .aws directory exists
			awsDir := filepath.Dir(credentialsPath)
			if err := os.MkdirAll(awsDir, 0700); err != nil {
				return fmt.Errorf("failed to create .aws directory: %w", err)
			}

			// Load or create credentials file
			var iniFile *ini.File
			if _, err := os.Stat(credentialsPath); err == nil {
				iniFile, err = ini.Load(credentialsPath)
				if err != nil {
					return fmt.Errorf("failed to load credentials file: %w", err)
				}
			} else {
				iniFile = ini.Empty()
			}

			// Create or update profile section
			profileSection := profile
			if profileSection == "" {
				profileSection = "default"
			}
			section, err := iniFile.NewSection(profileSection)
			if err != nil {
				return fmt.Errorf("failed to create profile section: %w", err)
			}

			// Set credentials
			section.Key("aws_access_key_id").SetValue(*result.Credentials.AccessKeyId)
			section.Key("aws_secret_access_key").SetValue(*result.Credentials.SecretAccessKey)
			section.Key("aws_session_token").SetValue(*result.Credentials.SessionToken)

			// Save the file
			if err := iniFile.SaveTo(credentialsPath); err != nil {
				return fmt.Errorf("failed to save credentials file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully assumed role %s and saved credentials to profile %s\n", roleArn, profileSection)
			return nil
		} else {
			// Verify we can make AWS calls with the current credentials
			stsClient := newSTSIdentityClient(cfg)
			_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if err != nil {
				if isCredentialError(err) {
					return fmt.Errorf("failed to verify AWS credentials\n%s", getCredentialHelp())
				}
				return fmt.Errorf("failed to verify AWS credentials: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Successfully verified AWS credentials")
			return nil
		}
	},
}

var (
	profile     string
	roleArn     string
	mfaSerial   string
	mfaToken    string
	duration    int32
	sessionName string
	listRoles   bool
)

func init() {
	checkCmd.AddCommand(checkCredentialsCmd)

	checkCredentialsCmd.Flags().StringVar(&profile, "profile", "", "AWS profile to save credentials to")
	checkCredentialsCmd.Flags().StringVar(&roleArn, "role-arn", "", "ARN of the role to assume")
	checkCredentialsCmd.Flags().StringVar(&mfaSerial, "mfa-serial", "", "Serial number of the MFA device")
	checkCredentialsCmd.Flags().StringVar(&mfaToken, "mfa-token", "", "Token from the MFA device")
	checkCredentialsCmd.Flags().Int32Var(&duration, "duration", 3600, "Duration in seconds for the assumed role session")
	checkCredentialsCmd.Flags().StringVar(&sessionName, "session-name", "ec-manager", "Name for the assumed role session")
	checkCredentialsCmd.Flags().BoolVar(&listRoles, "list-roles", false, "List available roles")
}
