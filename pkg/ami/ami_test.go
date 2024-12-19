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
	describeImagesOutput    *ec2.DescribeImagesOutput
	describeImagesError     error
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	createSnapshotOutput   *ec2.CreateSnapshotOutput
	createSnapshotError    error
	terminateInstancesOutput *ec2.TerminateInstancesOutput
	terminateInstancesError  error
	runInstancesOutput     *ec2.RunInstancesOutput
	runInstancesError      error
	createTagsError        error
	stopInstancesOutput    *ec2.StopInstancesOutput
	stopInstancesError     error
	startInstancesOutput   *ec2.StartInstancesOutput
	startInstancesError    error
	attachVolumeOutput    *ec2.AttachVolumeOutput
	attachVolumeError     error
	createVolumeOutput    *ec2.CreateVolumeOutput
	createVolumeError     error
	describeSnapshotsOutput *ec2.DescribeSnapshotsOutput
	describeSnapshotsError  error
	describeVolumesOutput  *ec2.DescribeVolumesOutput
	describeVolumesError   error
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m.describeImagesOutput, m.describeImagesError
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
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

func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	return m.attachVolumeOutput, m.attachVolumeError
}

func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	return m.createVolumeOutput, m.createVolumeError
}

func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	return m.describeSnapshotsOutput, m.describeSnapshotsError
}

func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	return m.describeVolumesOutput, m.describeVolumesError
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
		shouldMigrate     bool
	}{
		{
			name: "running instance with both tags should migrate",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceType: types.InstanceTypeT2Micro,
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate-if-running"),
							Value: aws.String("enabled"),
						},
						{
							Key:   aws.String("ami-migrate"),
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
			shouldMigrate: true,
		},
		{
			name: "running instance without if-running tag should not migrate",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceType: types.InstanceTypeT2Micro,
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate"),
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
			shouldMigrate: false,
		},
		{
			name: "stopped instance should migrate with only ami-migrate tag",
			instances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					State:        &types.InstanceState{Name: types.InstanceStateNameStopped},
					InstanceType: types.InstanceTypeT2Micro,
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate"),
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
			shouldMigrate: true,
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
				runInstancesOutput: &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-new123"),
							State: &types.InstanceState{
								Name: types.InstanceStateNamePending,
							},
						},
					},
				},
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

func TestMigrateInstance(t *testing.T) {
	tests := []struct {
		name          string
		instanceID    string
		oldAMI        string
		newAMI        string
		mockClient    *MockEC2Client
		expectedError bool
	}{
		{
			name:       "successful_migration",
			instanceID: "i-123",
			oldAMI:     "ami-old",
			newAMI:     "ami-new",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									ImageId:    aws.String("ami-old"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameStopped,
									},
								},
							},
						},
					},
				},
				stopInstancesOutput:  &ec2.StopInstancesOutput{},
				startInstancesOutput: &ec2.StartInstancesOutput{},
				runInstancesOutput: &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-new"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:       "wrong_AMI",
			instanceID: "i-123",
			oldAMI:     "ami-old",
			newAMI:     "ami-new",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									ImageId:    aws.String("ami-different"),
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name:       "instance_not_found",
			instanceID: "i-123",
			oldAMI:     "ami-old",
			newAMI:     "ami-new",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.mockClient)
			err := svc.MigrateInstance(context.Background(), tt.instanceID, tt.oldAMI, tt.newAMI)
			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBackupInstance(t *testing.T) {
	tests := []struct {
		name          string
		instanceID    string
		mockClient    *MockEC2Client
		expectedError bool
	}{
		{
			name:       "successful_backup",
			instanceID: "i-123",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
										{
											DeviceName: aws.String("/dev/sda1"),
											Ebs: &types.EbsInstanceBlockDevice{
												VolumeId: aws.String("vol-123"),
											},
										},
									},
								},
							},
						},
					},
				},
				createSnapshotOutput: &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				},
			},
			expectedError: false,
		},
		{
			name:       "instance_not_found",
			instanceID: "i-123",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			expectedError: true,
		},
		{
			name:       "snapshot_creation_fails",
			instanceID: "i-123",
			mockClient: &MockEC2Client{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
										{
											DeviceName: aws.String("/dev/sda1"),
											Ebs: &types.EbsInstanceBlockDevice{
												VolumeId: aws.String("vol-123"),
											},
										},
									},
								},
							},
						},
					},
				},
				createSnapshotError: fmt.Errorf("failed to create snapshot"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.mockClient)
			err := svc.BackupInstance(context.Background(), tt.instanceID)
			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
