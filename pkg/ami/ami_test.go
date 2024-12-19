package ami

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type MockEC2Client struct {
	describeImagesOutput  *ec2.DescribeImagesOutput
	describeImagesError   error
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	createSnapshotOutput *ec2.CreateSnapshotOutput
	createSnapshotError  error
	terminateInstancesOutput *ec2.TerminateInstancesOutput
	terminateInstancesError  error
	runInstancesOutput *ec2.RunInstancesOutput
	runInstancesError  error
	createTagsError    error
	stopInstancesOutput *ec2.StopInstancesOutput
	stopInstancesError  error
	startInstancesOutput *ec2.StartInstancesOutput
	startInstancesError  error
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m.describeImagesOutput, m.describeImagesError
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	// For waiters, always return the desired state immediately
	if len(params.InstanceIds) == 1 {
		inst := types.Instance{
			InstanceId: aws.String(params.InstanceIds[0]),
			State: &types.InstanceState{
				Name: types.InstanceStateNameRunning,
			},
		}
		return &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{inst},
				},
			},
		}, nil
	}
	return m.describeInstancesOutput, m.describeInstancesError
}

func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	return m.createSnapshotOutput, m.createSnapshotError
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return m.terminateInstancesOutput, m.terminateInstancesError
}

func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.runInstancesOutput, m.runInstancesError
}

func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return &ec2.CreateTagsOutput{}, m.createTagsError
}

func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	return m.stopInstancesOutput, m.stopInstancesError
}

func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	return m.startInstancesOutput, m.startInstancesError
}

func TestGetAMIWithTag(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     *ec2.DescribeImagesOutput
		mockError      error
		expectedAMI    string
		expectedError  bool
		tagKey         string
		tagValue       string
	}{
		{
			name: "successful retrieval",
			mockOutput: &ec2.DescribeImagesOutput{
				Images: []types.Image{
					{
						ImageId: aws.String("ami-123"),
					},
				},
			},
			mockError:     nil,
			expectedAMI:   "ami-123",
			expectedError: false,
			tagKey:       "Status",
			tagValue:     "latest",
		},
		{
			name:           "no images found",
			mockOutput:     &ec2.DescribeImagesOutput{Images: []types.Image{}},
			mockError:      nil,
			expectedAMI:    "",
			expectedError:  false,
			tagKey:        "Status",
			tagValue:      "latest",
		},
		{
			name:           "aws error",
			mockOutput:     nil,
			mockError:      fmt.Errorf("AWS API error"),
			expectedAMI:    "",
			expectedError:  true,
			tagKey:        "Status",
			tagValue:      "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockEC2Client{
				describeImagesOutput: tt.mockOutput,
				describeImagesError:  tt.mockError,
			}

			s := NewService(mockClient)
			ami, err := s.GetAMIWithTag(context.Background(), tt.tagKey, tt.tagValue)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ami != tt.expectedAMI {
				t.Errorf("expected AMI %s but got %s", tt.expectedAMI, ami)
			}
		})
	}
}

func TestMigrateInstances(t *testing.T) {
	tests := []struct {
		name              string
		instances         []types.Instance
		mockError         error
		expectedError     bool
		shouldStart       bool
		shouldStop        bool
	}{
		{
			name: "successful migration of running instance",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceType: types.InstanceTypeT2Micro,
					BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
						{
							Ebs: &types.EbsInstanceBlockDevice{
								VolumeId: aws.String("vol-123"),
							},
						},
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			shouldStart:   false,
			shouldStop:    false,
		},
		{
			name: "successful migration of stopped instance with if-running tag",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameStopped},
					InstanceType: types.InstanceTypeT2Micro,
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate-if-running"),
							Value: aws.String("enabled"),
						},
					},
					BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
						{
							Ebs: &types.EbsInstanceBlockDevice{
								VolumeId: aws.String("vol-123"),
							},
						},
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			shouldStart:   true,
			shouldStop:    true,
		},
		{
			name: "skip stopped instance without if-running tag",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameStopped},
					InstanceType: types.InstanceTypeT2Micro,
				},
			},
			mockError:     nil,
			expectedError: false,
			shouldStart:   false,
			shouldStop:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with timeout to prevent hanging
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mockClient := &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: tt.instances,
						},
					},
				},
				describeInstancesError:   tt.mockError,
				createSnapshotOutput:     &ec2.CreateSnapshotOutput{SnapshotId: aws.String("snap-123")},
				terminateInstancesOutput: &ec2.TerminateInstancesOutput{},
				runInstancesOutput:      &ec2.RunInstancesOutput{},
				stopInstancesOutput:     &ec2.StopInstancesOutput{},
				startInstancesOutput:    &ec2.StartInstancesOutput{},
			}

			s := NewService(mockClient)
			err := s.MigrateInstances(ctx, "ami-old", "ami-new", "enabled")

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTagAMI(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError bool
		amiID         string
		tagKey        string
		tagValue      string
	}{
		{
			name:          "successful tag",
			mockError:     nil,
			expectedError: false,
			amiID:        "ami-123",
			tagKey:       "Status",
			tagValue:     "latest",
		},
		{
			name:          "aws error",
			mockError:     fmt.Errorf("AWS API error"),
			expectedError: true,
			amiID:        "ami-123",
			tagKey:       "Status",
			tagValue:     "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockEC2Client{
				createTagsError: tt.mockError,
			}

			s := NewService(mockClient)
			err := s.TagAMI(context.Background(), tt.amiID, tt.tagKey, tt.tagValue)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
