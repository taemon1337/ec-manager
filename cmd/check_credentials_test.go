package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
	mockclient "github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	ectypes "github.com/taemon1337/ec-manager/pkg/types"
)

func TestCheckCredentialsCmd(t *testing.T) {
	tests := []testutil.CommandTestCase{
		{
			Name: "success",
			Args: []string{},
			SetupContext: func(ctx context.Context) context.Context {
				mockEC2Client := mockclient.NewMockEC2Client(t)
				mockIAMClient := mockclient.NewMockIAMClient(t)
				mockSTSClient := mockclient.NewMockSTSClient(t)

				// Mock EC2 DescribeInstances
				mockEC2Client.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-1234567890abcdef0"),
								},
							},
						},
					},
				}, nil)

				// Mock GetCallerIdentity
				mockSTSClient.On("GetCallerIdentity", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
					Arn:     aws.String("arn:aws:iam::123456789012:user/test-user"),
					UserId:  aws.String("AIDAXXXXXXXXXXXXXXXX"),
				}, nil)

				// Mock IAM ListUsers
				mockIAMClient.On("ListUsers", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&iam.ListUsersOutput{
					Users: []iamtypes.User{
						{
							UserName: aws.String("test-user"),
							Arn:      aws.String("arn:aws:iam::123456789012:user/test-user"),
						},
					},
				}, nil)

				// Mock IAM GetUser
				mockIAMClient.On("GetUser", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&iam.GetUserOutput{
					User: &iamtypes.User{
						UserName: aws.String("test-user"),
						Arn:      aws.String("arn:aws:iam::123456789012:user/test-user"),
					},
				}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				ctx = context.WithValue(ctx, ectypes.IAMClientKey, mockIAMClient)
				ctx = context.WithValue(ctx, ectypes.STSClientKey, mockSTSClient)
				return ctx
			},
		},
		{
			Name:        "error",
			Args:        []string{},
			WantErr:     true,
			ErrContains: "failed to get caller identity",
			SetupContext: func(ctx context.Context) context.Context {
				mockEC2Client := mockclient.NewMockEC2Client(t)
				mockIAMClient := mockclient.NewMockIAMClient(t)
				mockSTSClient := mockclient.NewMockSTSClient(t)

				// Mock EC2 DescribeInstances
				mockEC2Client.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-1234567890abcdef0"),
								},
							},
						},
					},
				}, nil)

				// Mock GetCallerIdentity with error
				mockSTSClient.On("GetCallerIdentity", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(nil, errors.New("failed to get caller identity"))

				// Mock IAM ListUsers
				mockIAMClient.On("ListUsers", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
					return true
				}), mock.Anything).Return(&iam.ListUsersOutput{
					Users: []iamtypes.User{
						{
							UserName: aws.String("test-user"),
							Arn:      aws.String("arn:aws:iam::123456789012:user/test-user"),
						},
					},
				}, nil)

				ctx = context.WithValue(ctx, ectypes.EC2ClientKey, mockEC2Client)
				ctx = context.WithValue(ctx, ectypes.IAMClientKey, mockIAMClient)
				ctx = context.WithValue(ctx, ectypes.STSClientKey, mockSTSClient)
				return ctx
			},
		},
	}

	testutil.RunCommandTest(t, NewCheckCredentialsCmd, tests)
}
