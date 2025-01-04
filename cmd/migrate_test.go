package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	mockclient "github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	ectypes "github.com/taemon1337/ec-manager/pkg/types"
)

func TestMigrateCmd(t *testing.T) {
	tests := []testutil.CommandTestCase{
		{
			Name: "success",
			Args: []string{"-i", "i-1234567890abcdef0", "-a", "ami-0987654321fedcba0"},
			SetupContext: func(ctx context.Context) context.Context {
				mockEC2Client := mockclient.NewMockEC2Client(t)

				// Mock DescribeInstances
				mockEC2Client.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:      aws.String("i-1234567890abcdef0"),
									Platform:        types.PlatformValuesWindows,
									PlatformDetails: aws.String("Windows"),
									ImageId:         aws.String("ami-0987654321fedcba0"),
									SubnetId:        aws.String("subnet-1234567890abcdef0"),
									KeyName:         aws.String("test-key"),
								},
							},
						},
					},
				}, nil)

				// Mock RunInstances
				mockEC2Client.On("RunInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-1234567890abcdef1"),
						},
					},
				}, nil)

				// Mock CreateTags
				mockEC2Client.On("CreateTags", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.CreateTagsOutput{}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				return ctx
			},
		},
		{
			Name:        "instance_not_found",
			Args:        []string{"-i", "i-1234567890abcdef0", "-a", "ami-0987654321fedcba0"},
			WantErr:     true,
			ErrContains: "instance not found",
			SetupContext: func(ctx context.Context) context.Context {
				mockEC2Client := mockclient.NewMockEC2Client(t)

				// Mock DescribeInstances with empty response
				mockEC2Client.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				return ctx
			},
		},
		{
			Name:        "error",
			Args:        []string{"-i", "i-1234567890abcdef0", "-a", "ami-0987654321fedcba0"},
			WantErr:     true,
			ErrContains: "failed to describe instance",
			SetupContext: func(ctx context.Context) context.Context {
				mockEC2Client := mockclient.NewMockEC2Client(t)

				// Mock DescribeInstances with error
				mockEC2Client.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(nil, errors.New("failed to describe instance"))

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				return ctx
			},
		},
	}

	testutil.RunCommandTest(t, NewMigrateCmd, tests)
}
