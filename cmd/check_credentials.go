package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/types"
	"gopkg.in/ini.v1"
)

// loadConfig loads AWS configuration
func loadConfig(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx)
}

// saveCredentials saves AWS credentials to the credentials file
func saveCredentials(profile, accessKey, secretKey, sessionToken string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	awsDir := filepath.Join(homeDir, ".aws")
	credentialsPath := filepath.Join(awsDir, "credentials")

	// Check if .aws directory exists
	if mkdirErr := os.MkdirAll(awsDir, 0700); mkdirErr != nil {
		return fmt.Errorf("failed to create .aws directory: %w", mkdirErr)
	}

	// Check if credentials file exists
	if _, statErr := os.Stat(credentialsPath); statErr == nil {
		iniFile, err := ini.Load(credentialsPath)
		if err != nil {
			return fmt.Errorf("failed to load credentials file: %w", err)
		}

		profileSection := profile
		if profileSection == "" {
			profileSection = "default"
		}

		section, err := iniFile.NewSection(profileSection)
		if err != nil {
			return fmt.Errorf("failed to create profile section: %w", err)
		}

		section.Key("aws_access_key_id").SetValue(accessKey)
		section.Key("aws_secret_access_key").SetValue(secretKey)
		section.Key("aws_session_token").SetValue(sessionToken)

		if err := iniFile.SaveTo(credentialsPath); err != nil {
			return fmt.Errorf("failed to save credentials file: %w", err)
		}

		return nil
	} else {
		iniFile := ini.Empty()

		profileSection := profile
		if profileSection == "" {
			profileSection = "default"
		}

		section, err := iniFile.NewSection(profileSection)
		if err != nil {
			return fmt.Errorf("failed to create profile section: %w", err)
		}

		section.Key("aws_access_key_id").SetValue(accessKey)
		section.Key("aws_secret_access_key").SetValue(secretKey)
		section.Key("aws_session_token").SetValue(sessionToken)

		if err := iniFile.SaveTo(credentialsPath); err != nil {
			return fmt.Errorf("failed to save credentials file: %w", err)
		}

		return nil
	}
}

// isCredentialError checks if the error is related to missing or invalid credentials
func isCredentialError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "NoCredentialProviders") ||
		strings.Contains(errStr, "operation error STS: GetCallerIdentity") ||
		strings.Contains(errStr, "InvalidClientTokenId") ||
		strings.Contains(errStr, "ExpiredToken") ||
		strings.Contains(errStr, "AccessDenied") ||
		strings.Contains(errStr, "InvalidToken")
}

// getCredentialHelp returns a helpful message for credential setup
func getCredentialHelp() string {
	return `To configure AWS credentials, you can:
1. Set environment variables:
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_SESSION_TOKEN=your_session_token (optional)

2. Use AWS CLI:
   aws configure

3. Create/edit ~/.aws/credentials file:
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   aws_session_token = your_session_token (optional)`
}

// discoverRoleARN attempts to discover available roles for the user
func discoverRoleARN(ctx context.Context, iamClient types.IAMClient) ([]string, error) {
	var roles []string

	// List roles
	output, err := iamClient.ListRoles(ctx, &iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	for _, role := range output.Roles {
		if role.AssumeRolePolicyDocument == nil {
			continue
		}

		// Check if the role can be assumed by the current user/role
		if strings.Contains(*role.AssumeRolePolicyDocument, "sts:AssumeRole") {
			roles = append(roles, *role.Arn)
		}
	}

	return roles, nil
}

// ec2ClientWrapper wraps the EC2 client to implement the EC2Client interface
type ec2ClientWrapper struct {
	*ec2.Client
}

// NewInstanceRunningWaiter implements the EC2Client interface
func (c *ec2ClientWrapper) NewInstanceRunningWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
} {
	return ec2.NewInstanceRunningWaiter(c.Client)
}

// NewInstanceStoppedWaiter implements the EC2Client interface
func (c *ec2ClientWrapper) NewInstanceStoppedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
} {
	return ec2.NewInstanceStoppedWaiter(c.Client)
}

// NewInstanceTerminatedWaiter implements the EC2Client interface
func (c *ec2ClientWrapper) NewInstanceTerminatedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
} {
	return ec2.NewInstanceTerminatedWaiter(c.Client)
}

// NewVolumeAvailableWaiter implements the EC2Client interface
func (c *ec2ClientWrapper) NewVolumeAvailableWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
} {
	return ec2.NewVolumeAvailableWaiter(c.Client)
}

var (
	checkCredentialsCmd = &cobra.Command{
		Use:   "credentials",
		Short: "Check AWS credentials and optionally assume a role",
		Long: `Check AWS credentials and optionally assume a role. You can either:
1. Use existing credentials
2. Assume a role (with optional MFA)
3. Discover available roles that you can assume

The credentials will be stored in ~/.aws/credentials under the specified profile.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Load AWS configuration
			cfg, err := loadConfig(ctx)
			if err != nil {
				if isCredentialError(err) {
					fmt.Println("No valid AWS credentials found.")
					fmt.Println(getCredentialHelp())
					return err
				}
				return fmt.Errorf("failed to load AWS config: %w", err)
			}

			// Get EC2 client from context
			ec2Client, ok := ctx.Value(types.EC2ClientKey).(types.EC2Client)
			if !ok {
				ec2Client = &ec2ClientWrapper{ec2.NewFromConfig(cfg)}
			}

			// Get STS client from context
			stsClient, ok := ctx.Value(types.STSClientKey).(types.STSClient)
			if !ok {
				stsClient = sts.NewFromConfig(cfg)
			}

			// Get IAM client from context
			iamClient, ok := ctx.Value(types.IAMClientKey).(types.IAMClient)
			if !ok {
				iamClient = iam.NewFromConfig(cfg)
			}

			// Check EC2 access
			_, ec2Err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})

			// Check current credentials
			callerIdentity, stsErr := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

			// Check IAM access
			_, iamErr := iamClient.ListUsers(ctx, &iam.ListUsersInput{})

			// Handle errors
			if ec2Err != nil {
				return fmt.Errorf("failed to check EC2 access: %w", ec2Err)
			}
			if stsErr != nil {
				if isCredentialError(stsErr) {
					fmt.Println("No valid AWS credentials found.")
					fmt.Println(getCredentialHelp())
					return stsErr
				}
				return fmt.Errorf("failed to check STS access: %w", stsErr)
			}
			if iamErr != nil {
				return fmt.Errorf("failed to check IAM access: %w", iamErr)
			}

			fmt.Printf("Current identity: %s\n", *callerIdentity.Arn)

			// If discover flag is set, try to find available roles
			if discover {
				roles, err := discoverRoleARN(ctx, iamClient)
				if err != nil {
					return fmt.Errorf("failed to discover roles: %w", err)
				}

				if len(roles) == 0 {
					fmt.Println("No assumable roles found")
					return nil
				}

				fmt.Println("\nAvailable roles:")
				for _, role := range roles {
					fmt.Printf("- %s\n", role)
				}
				return nil
			}

			// If roleARN is provided, attempt to assume the role
			if roleARN != "" {
				input := &sts.AssumeRoleInput{
					RoleArn:         aws.String(roleARN),
					RoleSessionName: aws.String("ec-manager-session"),
				}

				if mfaToken != "" {
					input.SerialNumber = aws.String(mfaSerial)
					input.TokenCode = aws.String(mfaToken)
				}

				assumeRoleOutput, err := stsClient.AssumeRole(ctx, input)
				if err != nil {
					return fmt.Errorf("failed to assume role: %w", err)
				}

				// Save the temporary credentials
				err = saveCredentials(profile,
					*assumeRoleOutput.Credentials.AccessKeyId,
					*assumeRoleOutput.Credentials.SecretAccessKey,
					*assumeRoleOutput.Credentials.SessionToken)
				if err != nil {
					return fmt.Errorf("failed to save credentials: %w", err)
				}

				fmt.Printf("Successfully assumed role %s\n", roleARN)
				fmt.Printf("Temporary credentials saved to profile: %s\n", profile)
			}

			return nil
		},
	}

	roleARN   string
	mfaSerial string
	mfaToken  string
	profile   string
	discover  bool
)

// NewCheckCredentialsCmd creates a new check credentials command
func NewCheckCredentialsCmd() *cobra.Command {
	return checkCredentialsCmd
}

func init() {
	checkCredentialsCmd.Flags().StringVarP(&roleARN, "role", "r", "", "Role ARN to assume")
	checkCredentialsCmd.Flags().StringVarP(&mfaSerial, "mfa-serial", "s", "", "MFA device serial number")
	checkCredentialsCmd.Flags().StringVarP(&mfaToken, "mfa-token", "t", "", "MFA token code")
	checkCredentialsCmd.Flags().StringVarP(&profile, "profile", "p", "default", "AWS profile to save credentials to")
	checkCredentialsCmd.Flags().BoolVarP(&discover, "discover", "d", false, "Discover available roles")
}
