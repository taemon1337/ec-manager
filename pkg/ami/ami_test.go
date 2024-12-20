package ami

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ami-migrate/pkg/client"
	apitypes "github.com/taemon1337/ami-migrate/pkg/types"
)

func TestGetAMIWithTag(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *apitypes.MockEC2Client
		tagKey      string
		tagValue    string
		wantAMI     string
		wantErr     bool
		errContains string
	}{
		{
			name: "found AMI with tag",
			mockClient: &apitypes.MockEC2Client{
				DescribeImagesOutput: &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-123"),
							Tags: []types.Tag{
								{
									Key:   aws.String("Status"),
									Value: aws.String("latest"),
								},
							},
						},
					},
				},
			},
			tagKey:   "Status",
			tagValue: "latest",
			wantAMI:  "ami-123",
			wantErr:  false,
		},
		{
			name: "no AMI found",
			mockClient: &apitypes.MockEC2Client{
				DescribeImagesOutput: &ec2.DescribeImagesOutput{
					Images: []types.Image{},
				},
			},
			tagKey:      "Status",
			tagValue:    "latest",
			wantAMI:     "",
			wantErr:     true,
			errContains: "no AMI found",
		},
		{
			name: "error describing images",
			mockClient: &apitypes.MockEC2Client{
				DescribeImagesError: fmt.Errorf("describe images error"),
			},
			tagKey:      "Status",
			tagValue:    "latest",
			wantAMI:     "",
			wantErr:     true,
			errContains: "describe images error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)
			defer client.ResetClient()

			svc := NewService(tt.mockClient)
			ami, err := svc.GetAMIWithTag(context.Background(), tt.tagKey, tt.tagValue)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAMI, ami)
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

			mockClient := &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: tt.instances,
						},
					},
				},
				DescribeInstancesError:   tt.mockError,
				CreateSnapshotOutput:     &ec2.CreateSnapshotOutput{SnapshotId: aws.String("snap-123")},
				TerminateInstancesOutput: &ec2.TerminateInstancesOutput{},
				RunInstancesOutput: &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-new123"),
							State: &types.InstanceState{
								Name: types.InstanceStateNamePending,
							},
						},
					},
				},
				StopInstancesOutput:     &ec2.StopInstancesOutput{},
				StartInstancesOutput:    &ec2.StartInstancesOutput{},
			}

			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(mockClient)
			defer client.ResetClient()

			s := NewService(mockClient)
			err := s.MigrateInstances(ctx, "enabled")

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
			mockClient := &apitypes.MockEC2Client{
				CreateTagsError: tt.mockError,
			}

			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(mockClient)
			defer client.ResetClient()

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
		mockClient    *apitypes.MockEC2Client
		expectedError bool
	}{
		{
			name:       "successful_migration",
			instanceID: "i-123",
			mockClient: &apitypes.MockEC2Client{
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
											Key:   aws.String("Owner"),
											Value: aws.String("testuser"),
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
							ImageId: aws.String("ami-new"),
							Name:    aws.String("RHEL9"),
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
			},
			expectedError: false,
		},
		{
			name:       "wrong_AMI",
			instanceID: "i-123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									ImageId:          aws.String("ami-different"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
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
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)
			defer client.ResetClient()

			svc := NewService(tt.mockClient)
			err := svc.MigrateInstance(context.Background(), tt.instanceID)
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
		mockClient    *apitypes.MockEC2Client
		expectedError bool
	}{
		{
			name:       "successful_backup",
			instanceID: "i-123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
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
				CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				},
			},
			expectedError: false,
		},
		{
			name:       "instance_not_found",
			instanceID: "i-123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			expectedError: true,
		},
		{
			name:       "snapshot_creation_fails",
			instanceID: "i-123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
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
				CreateSnapshotError: fmt.Errorf("failed to create snapshot"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)
			defer client.ResetClient()

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

func TestCheckMigrationStatus(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockClient    *apitypes.MockEC2Client
		expectedError bool
	}{
		{
			name:   "instance_needs_migration",
			userID: "user123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									ImageId:          aws.String("ami-old"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									Tags: []types.Tag{
										{
											Key:   aws.String("Owner"),
											Value: aws.String("user123"),
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
							ImageId: aws.String("ami-new"),
							Name:    aws.String("RHEL9"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:   "instance_up_to_date",
			userID: "user123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									ImageId:          aws.String("ami-new"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									Tags: []types.Tag{
										{
											Key:   aws.String("Owner"),
											Value: aws.String("user123"),
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
							ImageId: aws.String("ami-new"),
							Name:    aws.String("RHEL9"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:   "no_instance_found",
			userID: "user123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			expectedError: true,
		},
		{
			name:   "ami_lookup_error",
			userID: "user123",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:       aws.String("i-123"),
									PlatformDetails: aws.String("Red Hat Enterprise Linux"),
									ImageId:          aws.String("ami-old"),
									InstanceType:    types.InstanceTypeT2Micro,
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									Tags: []types.Tag{
										{
											Key:   aws.String("Owner"),
											Value: aws.String("user123"),
										},
									},
								},
							},
						},
					},
				},
				DescribeImagesError: fmt.Errorf("failed to lookup AMI"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)
			defer client.ResetClient()

			svc := NewService(tt.mockClient)
			_, err := svc.CheckMigrationStatus(context.Background(), tt.userID)
			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestListUserInstances(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockInstances  []types.Instance
		expectedCount  int
		expectedError  bool
		mockError      error
	}{
		{
			name:   "success with multiple instances",
			userID: "testuser",
			mockInstances: []types.Instance{
				{
					InstanceId:   aws.String("i-123"),
					InstanceType: types.InstanceTypeT3Large,
					State:       &types.InstanceState{Name: types.InstanceStateNameRunning},
					LaunchTime:        aws.Time(time.Now()),
					PrivateIpAddress: aws.String("10.0.0.1"),
					PublicIpAddress:  aws.String("54.123.45.67"),
					Tags: []types.Tag{
						{Key: aws.String("Name"), Value: aws.String("test-1")},
						{Key: aws.String("Owner"), Value: aws.String("testuser")},
					},
				},
				{
					InstanceId:   aws.String("i-456"),
					InstanceType: types.InstanceTypeT3Xlarge,
					State:       &types.InstanceState{Name: types.InstanceStateNameStopped},
					LaunchTime:        aws.Time(time.Now()),
					PrivateIpAddress: aws.String("10.0.0.2"),
					PublicIpAddress:  aws.String("54.123.45.68"),
					Tags: []types.Tag{
						{Key: aws.String("Name"), Value: aws.String("test-2")},
						{Key: aws.String("Owner"), Value: aws.String("testuser")},
					},
				},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:           "success with no instances",
			userID:         "testuser",
			mockInstances:  []types.Instance{},
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name:           "aws error",
			userID:         "testuser",
			mockError:      fmt.Errorf("aws error"),
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2 := &apitypes.MockEC2Client{}
			if tt.mockError != nil {
				mockEC2.DescribeInstancesError = tt.mockError
			} else {
				mockEC2.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: tt.mockInstances,
						},
					},
				}
			}

			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(mockEC2)
			defer client.ResetClient()

			service := &Service{client: mockEC2}
			instances, err := service.ListUserInstances(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(instances))

			if len(tt.mockInstances) > 0 {
				assert.Equal(t, aws.ToString(tt.mockInstances[0].InstanceId), instances[0].InstanceID)
			}
		})
	}
}

func TestCreateInstance(t *testing.T) {
	tests := []struct {
		name          string
		config        InstanceConfig
		mockInstance  types.Instance
		expectedError bool
		mockError     error
	}{
		{
			name: "success ubuntu instance",
			config: InstanceConfig{
				Name:   "test-instance",
				OSType: "Ubuntu",
				Size:   "large",
				UserID: "testuser",
			},
			mockInstance: types.Instance{
				InstanceId:   aws.String("i-123"),
				InstanceType: types.InstanceTypeT3Large,
				State:       &types.InstanceState{Name: types.InstanceStateNamePending},
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String("test-instance")},
					{Key: aws.String("Owner"), Value: aws.String("testuser")},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid size",
			config: InstanceConfig{
				Name:   "test-instance",
				OSType: "Ubuntu",
				Size:   "invalid",
				UserID: "testuser",
			},
			expectedError: true,
		},
		{
			name: "aws error",
			config: InstanceConfig{
				Name:   "test-instance",
				OSType: "Ubuntu",
				Size:   "large",
				UserID: "testuser",
			},
			mockError:     fmt.Errorf("aws error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2 := &apitypes.MockEC2Client{}
			if tt.mockError != nil {
				mockEC2.RunInstancesError = tt.mockError
			} else {
				mockEC2.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{tt.mockInstance},
				}
			}

			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(mockEC2)
			defer client.ResetClient()

			service := &Service{client: mockEC2}
			summary, err := service.CreateInstance(context.Background(), tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.config.Name, summary.Name)
			assert.Equal(t, tt.config.OSType, summary.OSType)
			assert.Equal(t, string(tt.mockInstance.InstanceType), summary.Size)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	tests := []struct {
		name          string
		mockClient    *apitypes.MockEC2Client
		userID        string
		instanceID    string
		expectedError bool
	}{
		{
			name: "success",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									Tags: []types.Tag{
										{
											Key:   aws.String("Owner"),
											Value: aws.String("testuser"),
										},
									},
								},
							},
						},
					},
				},
			},
			userID:        "testuser",
			instanceID:    "i-123",
			expectedError: false,
		},
		{
			name: "instance not found",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			userID:        "testuser",
			instanceID:    "i-123",
			expectedError: true,
		},
		{
			name: "error describing instances",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesError: fmt.Errorf("API error"),
			},
			userID:        "testuser",
			instanceID:    "i-123",
			expectedError: true,
		},
		{
			name: "error terminating instance",
			mockClient: &apitypes.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									Tags: []types.Tag{
										{
											Key:   aws.String("Owner"),
											Value: aws.String("testuser"),
										},
									},
								},
							},
						},
					},
				},
				TerminateInstancesError: fmt.Errorf("API error"),
			},
			userID:        "testuser",
			instanceID:    "i-123",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)
			defer client.ResetClient()

			svc := NewService(tt.mockClient)
			err := svc.DeleteInstance(context.Background(), tt.userID, tt.instanceID)
			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
