package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// ClientError represents an error from the client package
type ClientError struct {
	Message string
	Err     error
}

func (e *ClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// GetEC2Client returns an EC2 client for testing or real usage
func GetEC2Client(ctx context.Context) (types.EC2ClientAPI, error) {
	// Check if we're in a test package
	if isTestPackage() {
		if ec2Client == nil {
			return nil, &ClientError{Message: "no EC2 client set for testing"}
		}
		return ec2Client, nil
	}

	// Load AWS configuration for real usage
	cfg, err := LoadAWSConfig(ctx)
	if err != nil {
		return nil, &ClientError{Message: "failed to load AWS config", Err: err}
	}

	ec2Client := ec2.NewFromConfig(cfg)
	return ec2Client, nil
}

// LoadAWSConfig loads AWS configuration and validates credentials
func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, checkCredentialsError(err)
	}

	// Try to get credentials to verify they exist
	_, err = cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Config{}, checkCredentialsError(err)
	}

	return cfg, nil
}

// checkCredentialsError enhances credential-related errors with helpful messages
func checkCredentialsError(err error) error {
	if err == nil {
		return nil
	}

	// Common credential error messages
	credMissing := "no EC2 IMDS role found"
	credExpired := "expired credentials"

	if err.Error() == credMissing || err.Error() == credExpired {
		homeDir, _ := os.UserHomeDir()
		awsConfigPath := filepath.Join(homeDir, ".aws", "credentials")
		
		return fmt.Errorf(`AWS credentials not found or invalid. To fix this:

1. Set up AWS credentials in one of these ways:
   a. Create credentials file at %s with:
      [default]
      aws_access_key_id = YOUR_ACCESS_KEY
      aws_secret_access_key = YOUR_SECRET_KEY
      
   b. Set environment variables:
      export AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
      export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
      
   c. Configure AWS CLI:
      aws configure

2. Verify your credentials are valid:
   aws sts get-caller-identity

Error details: %v`, awsConfigPath, err)
	}

	return err
}

// SetEC2Client sets the EC2 client (used for testing)
func SetEC2Client(client types.EC2ClientAPI) error {
	if !isTestPackage() {
		return &ClientError{Message: "cannot set EC2 client outside of test package"}
	}
	ec2Client = client
	return nil
}

// isTestPackage returns true if the code is running in a test package
func isTestPackage() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

var ec2Client types.EC2ClientAPI
