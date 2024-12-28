package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/mock"
)

func NewRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore an EC2 instance from a snapshot",
		Long:  "Create a new EC2 instance from a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			instanceID, err := cmd.Flags().GetString("instance")
			if err != nil {
				return err
			}

			if instanceID == "" {
				return fmt.Errorf("instance ID must be set")
			}

			amiID, err := cmd.Flags().GetString("ami")
			if err != nil {
				return err
			}

			if amiID == "" {
				return fmt.Errorf("AMI ID must be set")
			}

			mockClient, ok := ctx.Value(mock.EC2ClientKey).(*mock.MockEC2Client)
			if !ok {
				return fmt.Errorf("failed to get EC2 client")
			}

			// Get instance details
			output, err := mockClient.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				return err
			}

			if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
				return fmt.Errorf("instance not found")
			}

			// Get AMI details
			amiOutput, err := mockClient.DescribeImages(ctx, &ec2.DescribeImagesInput{
				ImageIds: []string{amiID},
			})
			if err != nil {
				return fmt.Errorf("failed to describe AMI: %v", err)
			}

			if len(amiOutput.Images) == 0 {
				return fmt.Errorf("AMI not found")
			}

			// Create new instance
			newInstance, err := mockClient.RunInstances(ctx, &ec2.RunInstancesInput{
				ImageId:      aws.String(amiID),
				InstanceType: output.Reservations[0].Instances[0].InstanceType,
				MinCount:     aws.Int32(1),
				MaxCount:     aws.Int32(1),
				SubnetId:     output.Reservations[0].Instances[0].SubnetId,
			})
			if err != nil {
				return err
			}

			// Create tags for new instance
			_, err = mockClient.CreateTags(ctx, &ec2.CreateTagsInput{
				Resources: []string{*newInstance.Instances[0].InstanceId},
				Tags:      output.Reservations[0].Instances[0].Tags,
			})
			if err != nil {
				return err
			}

			// Stop original instance
			_, err = mockClient.StopInstances(ctx, &ec2.StopInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				return err
			}

			// Start new instance
			_, err = mockClient.StartInstances(ctx, &ec2.StartInstancesInput{
				InstanceIds: []string{*newInstance.Instances[0].InstanceId},
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringP("instance", "i", "", "Instance to restore")
	cmd.Flags().StringP("ami", "a", "", "AMI to restore from")
	return cmd
}

func TestRestoreCmd(t *testing.T) {
	mockClient := mock.NewMockEC2Client()

	// Setup test instance
	instance := types.Instance{
		InstanceId:   aws.String("i-123"),
		InstanceType: types.InstanceTypeT2Micro,
		SubnetId:     aws.String("subnet-123"),
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

	// Setup test AMI
	image := types.Image{
		ImageId:      aws.String("ami-123"),
		Name:        aws.String("test-ami"),
		Description: aws.String("Test AMI"),
		State:       types.ImageStateAvailable,
	}

	tests := []struct {
		name       string
		instance   string
		ami        string
		setupMock  func(*mock.MockEC2Client)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "successful restore",
			instance: "i-123",
			ami:      "ami-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{instance},
						},
					},
				}, nil)

				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-123"},
				}).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{image},
				}, nil)

				m.On("RunInstances", mock.Anything, mock.MatchedBy(func(input *ec2.RunInstancesInput) bool {
					return *input.ImageId == "ami-123" && *input.SubnetId == "subnet-123"
				})).Return(&ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-456"),
						},
					},
				}, nil)

				m.On("CreateTags", mock.Anything, mock.MatchedBy(func(input *ec2.CreateTagsInput) bool {
					return len(input.Resources) == 1 && input.Resources[0] == "i-456"
				})).Return(&ec2.CreateTagsOutput{}, nil)

				m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.StopInstancesOutput{}, nil)

				m.On("StartInstances", mock.Anything, &ec2.StartInstancesInput{
					InstanceIds: []string{"i-456"},
				}).Return(&ec2.StartInstancesOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:     "instance not found",
			instance: "i-999",
			ami:      "ami-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-999"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}, nil)
			},
			wantErr:    true,
			wantErrMsg: "instance not found",
		},
		{
			name:     "ami not found",
			instance: "i-123",
			ami:      "ami-999",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{instance},
						},
					},
				}, nil)

				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-999"},
				}).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}, nil)
			},
			wantErr:    true,
			wantErrMsg: "AMI not found",
		},
		{
			name:      "missing instance ID",
			ami:       "ami-123",
			setupMock: func(m *mock.MockEC2Client) {},
			wantErr:   true,
			wantErrMsg: "instance ID must be set",
		},
		{
			name:      "missing ami ID",
			instance:  "i-123",
			setupMock: func(m *mock.MockEC2Client) {},
			wantErr:   true,
			wantErrMsg: "AMI ID must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockClient)

			// Reset command and flags for each test
			cmd := NewRestoreCmd()

			// Set up command arguments
			var args []string
			if tt.instance != "" {
				args = append(args, "--instance", tt.instance)
			}
			if tt.ami != "" {
				args = append(args, "--ami", tt.ami)
			}
			cmd.SetArgs(args)

			ctx := context.WithValue(context.Background(), mock.EC2ClientKey, mockClient)
			cmd.SetContext(ctx)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
