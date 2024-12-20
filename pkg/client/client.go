package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ami-migrate/pkg/types"
)

var (
	ec2Client types.EC2ClientAPI
	inTest    bool
)

// GetEC2Client returns the EC2 client, creating it if necessary
func GetEC2Client() types.EC2ClientAPI {
	if ec2Client != nil {
		return ec2Client
	}

	if inTest {
		panic("EC2 client not set in test mode. You must call SetEC2Client with a mock client before using GetEC2Client in tests")
	}

	// Ensure we're not in a test package
	if isTestPackage() {
		panic("Attempting to use real AWS client in test code. Use SetEC2Client with a mock client instead")
	}

	// Load AWS configuration for real usage
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}
	ec2Client = ec2.NewFromConfig(cfg)
	return ec2Client
}

// SetEC2Client sets the EC2 client (used for testing)
func SetEC2Client(client types.EC2ClientAPI) {
	if !isTestPackage() {
		panic("SetEC2Client should only be called from test code")
	}
	ec2Client = client
	inTest = true
}

// ResetClient resets the client (used for testing)
func ResetClient() {
	if !isTestPackage() {
		panic("ResetClient should only be called from test code")
	}
	ec2Client = nil
	inTest = false
}

// isTestPackage returns true if the code is running in a test package
func isTestPackage() bool {
	return testing.Testing()
}
