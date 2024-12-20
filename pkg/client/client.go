package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ec-manager/pkg/types"
)

var (
	ec2Client types.EC2ClientAPI
	inTest    bool
)

// ClientError represents an error from the EC2 client operations
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

// GetEC2Client returns the EC2 client, creating it if necessary
func GetEC2Client(ctx context.Context) (types.EC2ClientAPI, error) {
	if ec2Client != nil {
		return ec2Client, nil
	}

	if inTest {
		return nil, &ClientError{Message: "EC2 client not set in test mode. You must call SetEC2Client with a mock client before using GetEC2Client in tests"}
	}

	// Ensure we're not in a test package
	if isTestPackage() {
		return nil, &ClientError{Message: "Attempting to use real AWS client in test code. Use SetEC2Client with a mock client instead"}
	}

	// Load AWS configuration for real usage
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, &ClientError{Message: "failed to load AWS config", Err: err}
	}
	
	ec2Client = ec2.NewFromConfig(cfg)
	return ec2Client, nil
}

// SetEC2Client sets the EC2 client (used for testing)
func SetEC2Client(client types.EC2ClientAPI) error {
	if !isTestPackage() {
		return &ClientError{Message: "SetEC2Client should only be called from test code"}
	}
	ec2Client = client
	inTest = true
	return nil
}

// ResetClient resets the client (used for testing)
func ResetClient() error {
	if !isTestPackage() {
		return &ClientError{Message: "ResetClient should only be called from test code"}
	}
	ec2Client = nil
	inTest = false
	return nil
}

// isTestPackage returns true if the code is running in a test package
func isTestPackage() bool {
	return strings.HasSuffix(os.Args[0], ".test") || 
		(len(os.Args) > 1 && os.Args[1] == "-test.v")
}
