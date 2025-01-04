package ami

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEC2Client struct {
	mock.Mock
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, input, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

func (m *mockEC2Client) CreateImage(ctx context.Context, input *ec2.CreateImageInput, opts ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	args := m.Called(ctx, input, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.CreateImageOutput), args.Error(1)
}

func (m *mockEC2Client) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	args := m.Called(ctx, input, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.CreateTagsOutput), args.Error(1)
}

func TestCreateAMI(t *testing.T) {
	t.Parallel()
	testTimeout := time.Second * 10

	tests := []struct {
		name    string
		setup   func(*mockEC2Client)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(m *mockEC2Client) {
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
								},
							},
						},
					},
				}, nil).Once()

				m.On("CreateImage", mock.Anything, mock.MatchedBy(func(input *ec2.CreateImageInput) bool {
					return aws.ToString(input.InstanceId) == "i-123" && aws.ToString(input.Name) == "test-ami"
				}), mock.Anything).Return(&ec2.CreateImageOutput{
					ImageId: aws.String("ami-123"),
				}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error_describe_instances",
			setup: func(m *mockEC2Client) {
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(nil, errors.New("describe error")).Once()
			},
			wantErr: true,
		},
		{
			name: "error_create_image",
			setup: func(m *mockEC2Client) {
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
								},
							},
						},
					},
				}, nil).Once()

				m.On("CreateImage", mock.Anything, mock.MatchedBy(func(input *ec2.CreateImageInput) bool {
					return aws.ToString(input.InstanceId) == "i-123" && aws.ToString(input.Name) == "test-ami"
				}), mock.Anything).Return(nil, errors.New("create error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			mockClient := &mockEC2Client{}
			if tt.setup != nil {
				tt.setup(mockClient)
			}

			got, err := CreateAMI(ctx, "i-123", "test-ami", "test description", mockClient)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, "ami-123", got)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func CreateAMI(ctx context.Context, instanceID, amiName, description string, client *mockEC2Client) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	output, err := client.DescribeInstances(ctx, input)
	if err != nil {
		return "", err
	}
	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return "", errors.New("instance not found")
	}
	instance := output.Reservations[0].Instances[0]
	createImageInput := &ec2.CreateImageInput{
		InstanceId:  instance.InstanceId,
		Name:        aws.String(amiName),
		Description: aws.String(description),
	}
	createImageOutput, err := client.CreateImage(ctx, createImageInput)
	if err != nil {
		return "", err
	}
	return aws.ToString(createImageOutput.ImageId), nil
}
