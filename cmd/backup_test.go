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

func TestBackupCmd(t *testing.T) {
	tests := []testutil.CommandTestCase{
		{
			Name: "success",
			Args: []string{"-i", "i-1234567890abcdef0"},
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
							Tags: []types.Tag{
								{
									Key:   aws.String("OS"),
									Value: aws.String("Windows"),
								},
							},
						},
					},
				}, nil)

				// Mock CreateImage
				mockEC2Client.On("CreateImage", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.CreateImageOutput{
					ImageId: aws.String("ami-1234567890abcdef0"),
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
			Name:        "error",
			Args:        []string{"-i", "i-1234567890abcdef0"},
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

	testutil.RunCommandTest(t, NewBackupCmd, tests)
}
