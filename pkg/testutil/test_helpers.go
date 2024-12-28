package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/mock"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// EC2ClientKey is the context key for the EC2 client
	EC2ClientKey ContextKey = "ec2_client"
)

// SetupMockEC2Client sets up a mock EC2 client for testing
func SetupMockEC2Client() *mock.MockEC2Client {
	return mock.NewMockEC2Client()
}

// GetTestContext returns a context with a mock EC2 client
func GetTestContext() context.Context {
	return context.WithValue(context.Background(), EC2ClientKey, SetupMockEC2Client())
}

// GetTestContextWithClient returns a context with the given mock EC2 client
func GetTestContextWithClient(client *mock.MockEC2Client) context.Context {
	return context.WithValue(context.Background(), EC2ClientKey, client)
}

// SetupTestCommand sets up a test command with the given arguments
func SetupTestCommand(cmd *cobra.Command, args []string) error {
	if cmd.Context() == nil {
		cmd.SetContext(GetTestContext())
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

// GetEC2Client retrieves the EC2 client from the context
func GetEC2Client(ctx context.Context) *mock.MockEC2Client {
	if client, ok := ctx.Value(EC2ClientKey).(*mock.MockEC2Client); ok {
		return client
	}
	return nil
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

// MockInstance creates a mock EC2 instance with common fields
func MockInstance(id string) types.Instance {
	return types.Instance{
		InstanceId:   aws.String(id),
		InstanceType: types.InstanceTypeT2Micro,
		KeyName:      aws.String("test-key"),
		SubnetId:     aws.String("subnet-123"),
		VpcId:        aws.String("vpc-123"),
		State: &types.InstanceState{
			Name: types.InstanceStateNameRunning,
		},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-instance"),
			},
		},
	}
}

// MockImage creates a mock AMI with common fields
func MockImage(id string, os string, version string) types.Image {
	tags := []types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("test-ami"),
		},
	}
	if os != "" {
		tags = append(tags, types.Tag{
			Key:   aws.String("OS"),
			Value: aws.String(os),
		})
	}
	if version != "" {
		tags = append(tags, types.Tag{
			Key:   aws.String("Version"),
			Value: aws.String(version),
		})
	}
	return types.Image{
		ImageId:      aws.String(id),
		Name:        aws.String("test-ami"),
		Description: aws.String("Test AMI"),
		State:       types.ImageStateAvailable,
		Tags:        tags,
	}
}

// SetupInstanceMock configures common instance-related mock responses
func SetupInstanceMock(client *mock.MockEC2Client, instance types.Instance) {
	client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{instance},
			},
		},
	}
}

// SetupImageMock configures common image-related mock responses
func SetupImageMock(client *mock.MockEC2Client, images ...types.Image) {
	client.DescribeImagesOutput = &ec2.DescribeImagesOutput{
		Images: images,
	}
}

// SetupDefaultMocks configures common mock responses for testing
func SetupDefaultMocks(client *mock.MockEC2Client) {
	instance := MockInstance("i-123")
	image := MockImage("ami-123", "RHEL9", "1.0.0")
	SetupInstanceMock(client, instance)
	SetupImageMock(client, image)
}
