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

// Variables to allow mocking in tests
var (
	loadConfig = config.LoadDefaultConfig
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

func NewLoginCmd() *cobra.Command {
	var (
		profile     string
		roleArn     string
		mfaSerial   string
		mfaToken    string
		duration    int32
		sessionName string
		listRoles   bool
	)

	cmd := &cobra.Command{
		Use:           "login",
		Short:         "Login to AWS and get temporary credentials",
		Long:          `Login to AWS and get temporary credentials using STS AssumeRole. The credentials will be stored in ~/.aws/credentials under the specified profile.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Load the shared AWS configuration
			cfg, err := loadConfig(ctx)
			if err != nil {
				return fmt.Errorf("unable to load AWS config: %w", err)
			}

			// If --list-roles is specified, discover and display available roles
			if listRoles {
				roles, err := discoverRoleARN(ctx, cfg)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStderr(), err.Error())
					return nil
				}
				
				fmt.Fprintln(cmd.OutOrStdout(), "Available roles:")
				for _, role := range roles {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", role)
				}
				return nil
			}

			// Validate required role ARN
			if roleArn == "" {
				roles, err := discoverRoleARN(ctx, cfg)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStderr(), err.Error())
					return fmt.Errorf("--role-arn is required. Use --list-roles to see available roles")
				}
				
				if len(roles) == 0 {
					return fmt.Errorf("--role-arn is required and no roles were found")
				}
				
				fmt.Fprintln(cmd.OutOrStdout(), "Available roles:")
				for _, role := range roles {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", role)
				}
				return fmt.Errorf("--role-arn is required. Please select one of the roles above")
			}

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
			section, err := iniFile.NewSection(profile)
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

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully logged in. Temporary credentials saved to profile '%s'\n", profile)
			fmt.Fprintf(cmd.OutOrStdout(), "Credentials will expire at: %v\n", result.Credentials.Expiration)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&profile, "profile", "default", "AWS profile to store credentials")
	cmd.Flags().StringVar(&roleArn, "role-arn", "", "ARN of the role to assume")
	cmd.Flags().StringVar(&mfaSerial, "mfa-serial", "", "ARN of the MFA device (if required)")
	cmd.Flags().StringVar(&mfaToken, "mfa-token", "", "MFA token code (if required)")
	cmd.Flags().Int32Var(&duration, "duration", 3600, "Duration in seconds for the temporary credentials")
	cmd.Flags().StringVar(&sessionName, "session-name", "ec-manager-session", "Name for the role session")
	cmd.Flags().BoolVar(&listRoles, "list-roles", false, "List available roles and exit")

	return cmd
}
