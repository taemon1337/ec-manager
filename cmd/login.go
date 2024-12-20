package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// STSAPI interface for mocking
type STSAPI interface {
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// Variable to allow mocking in tests
var newSTSClient = func(cfg aws.Config) STSAPI {
	return sts.NewFromConfig(cfg)
}

func NewLoginCmd() *cobra.Command {
	var (
		profile     string
		roleArn     string
		mfaSerial   string
		mfaToken    string
		duration    int32
		sessionName string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to AWS and get temporary credentials",
		Long: `Login to AWS and get temporary credentials using STS AssumeRole.
The credentials will be stored in ~/.aws/credentials under the specified profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Load the shared AWS configuration
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("unable to load AWS config: %w", err)
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

			fmt.Printf("Successfully logged in. Temporary credentials saved to profile '%s'\n", profile)
			fmt.Printf("Credentials will expire at: %v\n", result.Credentials.Expiration)

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

	// Mark required flags
	cmd.MarkFlagRequired("role-arn")

	return cmd
}
