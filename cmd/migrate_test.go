package cmd

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/taemon1337/ami-migrate/pkg/ami/mocks"
	"github.com/taemon1337/ami-migrate/pkg/client"
)

func TestMigrateCmd(t *testing.T) {
	// Reset client state before test
	client.ResetClient()
	defer client.ResetClient()

	// Initialize mock EC2 client for tests
	mockClient := &mocks.MockEC2Client{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:       aws.String("i-123"),
							ImageId:          aws.String("ami-old"),
							PlatformDetails: aws.String("Red Hat Enterprise Linux"),
							InstanceType:    types.InstanceTypeT2Micro,
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
							Tags: []types.Tag{
								{
									Key:   aws.String("ami-migrate"),
									Value: aws.String("enabled"),
								},
							},
						},
					},
				},
			},
		},
		DescribeImagesOutput: &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId: aws.String("ami-123"),
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate"),
							Value: aws.String("latest"),
						},
						{
							Key:   aws.String("OS"),
							Value: aws.String("Red Hat Enterprise Linux"),
						},
					},
					Name:        aws.String("RHEL-9"),
					Description: aws.String("Red Hat Enterprise Linux 9"),
				},
			},
		},
		RunInstancesOutput: &ec2.RunInstancesOutput{
			Instances: []types.Instance{
				{
					InstanceId: aws.String("i-456"),
				},
			},
		},
		StopInstancesOutput: &ec2.StopInstancesOutput{
			StoppingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameStopped,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
		StartInstancesOutput: &ec2.StartInstancesOutput{
			StartingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
		CreateTagsOutput: &ec2.CreateTagsOutput{},
		TerminateInstancesOutput: &ec2.TerminateInstancesOutput{
			TerminatingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameTerminated,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
	}

	// Set the mock client
	client.SetEC2Client(mockClient)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errContains string
	}{
		{
			name:    "no args",
			args:    []string{"migrate"},
			wantErr: false,
		},
		{
			name:    "with instance ID",
			args:    []string{"migrate", "--instance-id", "i-123"},
			wantErr: false,
		},
		{
			name:    "with enabled value",
			args:    []string{"migrate", "--enabled", "enabled"},
			wantErr: false,
		},
		{
			name:        "no new AMI",
			args:        []string{"migrate"},
			wantErr:     true,
			errContains: "--new-ami is required",
		},
		{
			name:    "with new AMI",
			args:    []string{"migrate", "--new-ami", "ami-123", "--instance-id", "i-123"},
			wantErr: false,
		},
		{
			name:        "instance not found",
			args:        []string{"migrate", "--new-ami", "ami-123", "--instance-id", "i-nonexistent"},
			wantErr:     true,
			errContains: "instance not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("migrate command error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
					t.Errorf("migrate command error = %v, want it to contain %v", err, tt.errContains)
				}
			}
		})
	}
}
