package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	mockclient "github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	ectypes "github.com/taemon1337/ec-manager/pkg/types"
)

var testTime = time.Date(2025, 1, 3, 15, 28, 3, 0, time.UTC)

func TestCheckMigrateCmd(t *testing.T) {
	tests := []testutil.CommandTestCase{
		{
			Name: "success",
			Args: []string{"--check-instance-id", "i-1234567890abcdef0", "--check-target-ami", "ami-0987654321fedcba0"},
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
									State:           &types.InstanceState{Name: types.InstanceStateNameRunning},
									InstanceType:    types.InstanceTypeT2Micro,
									LaunchTime:      aws.Time(testTime),
								},
							},
						},
					},
				}, nil)

				// Mock DescribeImages
				mockEC2Client.On("DescribeImages", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId:  aws.String("ami-0987654321fedcba0"),
							Platform: types.PlatformValuesWindows,
						},
					},
				}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				return ctx
			},
		},
		{
			Name:        "instance_not_found",
			Args:        []string{"--check-instance-id", "i-1234567890abcdef0", "--check-target-ami", "ami-0987654321fedcba0"},
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
			Name:        "ami_not_found",
			Args:        []string{"--check-instance-id", "i-1234567890abcdef0", "--check-target-ami", "ami-0987654321fedcba0"},
			WantErr:     true,
			ErrContains: "AMI not found",
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
									State:           &types.InstanceState{Name: types.InstanceStateNameRunning},
									InstanceType:    types.InstanceTypeT2Micro,
									LaunchTime:      aws.Time(testTime),
								},
							},
						},
					},
				}, nil)

				// Mock DescribeImages with empty response
				mockEC2Client.On("DescribeImages", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				return ctx
			},
		},
		{
			Name:        "error",
			Args:        []string{"--check-instance-id", "i-1234567890abcdef0", "--check-target-ami", "ami-0987654321fedcba0"},
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

	testutil.RunCommandTest(t, NewCheckMigrateCmd, tests)
}
