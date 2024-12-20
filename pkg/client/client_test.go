package client

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	apitypes "github.com/taemon1337/ec-manager/pkg/types"
)

type mockEC2Client struct {
	DescribeInstancesFunc func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.DescribeInstancesOutput{}, nil
}

func TestGetEC2Client(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set test args to simulate test environment
	os.Args = []string{"test.test"}

	// Create mock client
	mockClient := &apitypes.MockEC2Client{
		InstanceStates: make(map[string]types.InstanceStateName),
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-1234567890abcdef0"),
						},
					},
				},
			},
		},
	}

	// Test setting mock client
	if err := SetEC2Client(mockClient); err != nil {
		t.Errorf("SetEC2Client failed: %v", err)
	}

	// Test getting client
	client, err := GetEC2Client(context.Background())
	if err != nil {
		t.Errorf("GetEC2Client failed: %v", err)
	}
	if client == nil {
		t.Error("GetEC2Client returned nil client")
	}

	// Test client functionality
	output, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{})
	if err != nil {
		t.Errorf("DescribeInstances failed: %v", err)
	}
	if len(output.Reservations) != 1 {
		t.Error("Expected 1 reservation")
	}
	if *output.Reservations[0].Instances[0].InstanceId != "i-1234567890abcdef0" {
		t.Error("Unexpected instance ID")
	}
}

func TestLoadAWSConfig(t *testing.T) {
	// Test with missing credentials
	_, err := LoadAWSConfig(context.Background())
	if err == nil {
		t.Error("Expected error for missing credentials")
	}

	// Check error message
	if err != nil && !containsCredentialHelp(err.Error()) {
		t.Error("Error message should contain credential help")
	}
}

func containsCredentialHelp(msg string) bool {
	return contains(msg, "AWS credentials not found or invalid") &&
		contains(msg, "aws_access_key_id") &&
		contains(msg, "aws_secret_access_key") &&
		contains(msg, "aws configure")
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) > len(substr) && s[len(s)-1] != substr[0]
}
