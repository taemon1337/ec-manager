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
	"github.com/taemon1337/ec-manager/pkg/config"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

type mockEC2Client struct {
	DescribeInstancesFunc func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImagesFunc   func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTagsFunc       func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstancesFunc     func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	StopInstancesFunc    func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstancesFunc   func(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	AttachVolumeFunc    func(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshotFunc  func(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	TerminateInstancesFunc func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	CreateVolumeFunc    func(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.DescribeInstancesOutput{}, nil
}

func (m *mockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesFunc != nil {
		return m.DescribeImagesFunc(ctx, params, optFns...)
	}
	return &ec2.DescribeImagesOutput{}, nil
}

func (m *mockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsFunc != nil {
		return m.CreateTagsFunc(ctx, params, optFns...)
	}
	return &ec2.CreateTagsOutput{}, nil
}

func (m *mockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesFunc != nil {
		return m.RunInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.RunInstancesOutput{}, nil
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.StopInstancesFunc != nil {
		return m.StopInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.StopInstancesOutput{}, nil
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.StartInstancesFunc != nil {
		return m.StartInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.StartInstancesOutput{}, nil
}

func (m *mockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeFunc != nil {
		return m.AttachVolumeFunc(ctx, params, optFns...)
	}
	return &ec2.AttachVolumeOutput{}, nil
}

func (m *mockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotFunc != nil {
		return m.CreateSnapshotFunc(ctx, params, optFns...)
	}
	return &ec2.CreateSnapshotOutput{}, nil
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.TerminateInstancesFunc != nil {
		return m.TerminateInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.TerminateInstancesOutput{}, nil
}

func (m *mockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(ctx, params, optFns...)
	}
	return &ec2.CreateVolumeOutput{}, nil
}

type mockWaiter struct{}

func (w *mockWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return nil
}

func setupTest(t *testing.T) (*Service, *ecTypes.MockEC2Client) {
	mockClient := ecTypes.NewMockEC2Client()
	svc := NewService(mockClient)
	// Override the waiter to return immediately
	instanceStateWaiter = &mockWaiter{}
	return svc, mockClient
}

func TestGetAMIWithTag(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		tagKey      string
		tagValue    string
		wantErr     bool
		errContains string
	}{
		{
			name: "found AMI with tag",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{
							{
								ImageId: aws.String("ami-123"),
							},
						},
					}, nil
				}
			},
			tagKey:   "Status",
			tagValue: "latest",
			wantErr:  false,
		},
		{
			name: "no AMI found",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{},
					}, nil
				}
			},
			tagKey:      "Status",
			tagValue:    "latest",
			wantErr:     true,
			errContains: "no AMI found with tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockClient := setupTest(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			ami, err := svc.GetAMIWithTag(context.Background(), tt.tagKey, tt.tagValue)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ami)
				assert.NotEmpty(t, ami.ImageId)
			}
		})
	}
}

func TestTagAMI(t *testing.T) {
	svc, mockClient := setupTest(t)

	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		amiID       string
		tagKey      string
		tagValue    string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful tag",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
					return &ec2.CreateTagsOutput{}, nil
				}
			},
			amiID:    "ami-123",
			tagKey:   "Status",
			tagValue: "latest",
			wantErr:  false,
		},
		{
			name: "error tagging",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
					return nil, fmt.Errorf("failed to tag AMI")
				}
			},
			amiID:       "ami-123",
			tagKey:      "Status",
			tagValue:    "latest",
			wantErr:     true,
			errContains: "failed to tag AMI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			err := svc.TagAMI(context.Background(), tt.amiID, tt.tagKey, tt.tagValue)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMigrateInstance(t *testing.T) {
	svc, mockClient := setupTest(t)

	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		instanceID  string
		newAMI      string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful migration",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										ImageId:    aws.String("ami-old"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameStopped,
										},
										PlatformDetails: aws.String("Linux/UNIX"),
										Tags: []types.Tag{
											{
												Key:   aws.String("OS"),
												Value: aws.String("linux"),
											},
										},
									},
								},
							},
						},
					}, nil
				}
				m.StopInstancesFunc = func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
					return &ec2.StopInstancesOutput{
						StoppingInstances: []types.InstanceStateChange{
							{
								CurrentState: &types.InstanceState{
									Name: types.InstanceStateNameStopped,
								},
								InstanceId: aws.String("i-123"),
							},
						},
					}, nil
				}
				m.StartInstancesFunc = func(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
					return &ec2.StartInstancesOutput{
						StartingInstances: []types.InstanceStateChange{
							{
								CurrentState: &types.InstanceState{
									Name: types.InstanceStateNameRunning,
								},
								InstanceId: aws.String("i-123"),
							},
						},
					}, nil
				}
				m.RunInstancesFunc = func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
					return &ec2.RunInstancesOutput{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("i-456"),
								ImageId:    aws.String("ami-new"),
								State: &types.InstanceState{
									Name: types.InstanceStateNamePending,
								},
							},
						},
					}, nil
				}
				m.TerminateInstancesFunc = func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
					return &ec2.TerminateInstancesOutput{
						TerminatingInstances: []types.InstanceStateChange{
							{
								CurrentState: &types.InstanceState{
									Name: types.InstanceStateNameShuttingDown,
								},
								InstanceId: aws.String("i-123"),
							},
						},
					}, nil
				}
			},
			instanceID: "i-123",
			newAMI:     "ami-new",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{},
					}, nil
				}
			},
			instanceID:  "i-nonexistent",
			newAMI:     "ami-new",
			wantErr:    true,
			errContains: "instance not found",
		},
		{
			name: "stop instance error",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										ImageId:    aws.String("ami-old"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							},
						},
					}, nil
				}
				m.StopInstancesFunc = func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
					return nil, fmt.Errorf("failed to stop instance")
				}
			},
			instanceID:  "i-123",
			newAMI:     "ami-new",
			wantErr:    true,
			errContains: "failed to stop instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			err := svc.MigrateInstance(context.Background(), tt.instanceID, tt.newAMI)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBackupInstance(t *testing.T) {
	svc, mockClient := setupTest(t)

	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		instanceID  string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful backup",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							},
						},
					}, nil
				}
				m.CreateSnapshotFunc = func(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
					return &ec2.CreateSnapshotOutput{
						SnapshotId: aws.String("snap-123"),
					}, nil
				}
				m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
					return &ec2.CreateTagsOutput{}, nil
				}
			},
			instanceID: "i-123",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{},
					}, nil
				}
			},
			instanceID:   "i-123",
			wantErr:     true,
			errContains: "instance not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			err := svc.BackupInstance(context.Background(), tt.instanceID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListUserInstances(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		userID      string
		wantErr     bool
		errContains string
		wantEmpty   bool
	}{
		{
			name: "successful list",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
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
												Key:   aws.String("Name"),
												Value: aws.String("test-instance"),
											},
											{
												Key:   aws.String("Owner"),
												Value: aws.String("user123"),
											},
										},
									},
								},
							},
						},
					}, nil
				}
				m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{
							{
								ImageId: aws.String("ami-123"),
							},
						},
					}, nil
				}
			},
			userID:   "user123",
			wantErr:  false,
		},
		{
			name: "no instances found",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{},
					}, nil
				}
			},
			userID:    "user123",
			wantErr:   false,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockClient := setupTest(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			instances, err := svc.ListUserInstances(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.wantEmpty {
					assert.Empty(t, instances)
				} else {
					assert.NotEmpty(t, instances)
				}
			}
		})
	}
}

func TestCreateInstance(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		config      InstanceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "successful create",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{
							{
								ImageId: aws.String("ami-123"),
								Tags: []types.Tag{
									{
										Key:   aws.String("OS"),
										Value: aws.String("linux"),
									},
									{
										Key:   aws.String("Status"),
										Value: aws.String("latest"),
									},
								},
							},
						},
					}, nil
				}
				m.RunInstancesFunc = func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
					return &ec2.RunInstancesOutput{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("i-123"),
								State: &types.InstanceState{
									Name: types.InstanceStateNamePending,
								},
								Tags: []types.Tag{
									{
										Key:   aws.String("Name"),
										Value: aws.String("test-instance"),
									},
									{
										Key:   aws.String("Owner"),
										Value: aws.String("user123"),
									},
								},
							},
						},
					}, nil
				}
			},
			config: InstanceConfig{
				Name:   "test-instance",
				Size:   "small",
				OSType: "linux",
				UserID: "user123",
			},
			wantErr: false,
		},
		{
			name: "invalid size",
			setupMock: func(m *ecTypes.MockEC2Client) {},
			config: InstanceConfig{
				Name:   "test-instance",
				Size:   "invalid",
				OSType: "linux",
				UserID: "user123",
			},
			wantErr:     true,
			errContains: "invalid size: invalid. Must be one of: small, medium, large, xlarge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockClient := setupTest(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			instance, err := svc.CreateInstance(context.Background(), tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				assert.Equal(t, "i-123", instance.InstanceID)
				assert.Equal(t, tt.config.Name, instance.Name)
				assert.Equal(t, tt.config.OSType, instance.OSType)
				assert.Equal(t, tt.config.Size, instance.Size)
			}
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	svc, mockClient := setupTest(t)

	tests := []struct {
		name        string
		setupMock   func(*ecTypes.MockEC2Client)
		userID      string
		instanceID  string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful delete",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							},
						},
					}, nil
				}
				m.TerminateInstancesFunc = func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
					return &ec2.TerminateInstancesOutput{}, nil
				}
			},
			userID:     "user123",
			instanceID: "i-123",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{},
					}, nil
				}
			},
			userID:      "user123",
			instanceID:  "i-123",
			wantErr:     true,
			errContains: "instance i-123 not found or not owned by user user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			err := svc.DeleteInstance(context.Background(), tt.userID, tt.instanceID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMigrateInstances(t *testing.T) {
	svc, mockClient := setupTest(t)

	// Mock the waiter to return immediately
	instanceStateWaiter = &mockWaiter{}

	// Set a shorter timeout for tests
	oldTimeout := config.GetTimeout()
	config.SetTimeout(1 * time.Second)
	defer config.SetTimeout(oldTimeout)

	mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
		if params.Filters != nil && len(params.Filters) > 0 && *params.Filters[0].Name == "tag:ami-migrate" {
			// Return instances with ami-migrate=enabled tag
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								InstanceId:   aws.String("i-1234567890abcdef0"),
								InstanceType: types.InstanceTypeT2Micro,
								State: &types.InstanceState{
									Name: types.InstanceStateNameRunning,
								},
								BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
									{
										DeviceName: aws.String("/dev/xvda"),
										Ebs: &types.EbsInstanceBlockDevice{
											VolumeId: aws.String("vol-123"),
										},
									},
								},
								Tags: []types.Tag{
									{
										Key:   aws.String("Name"),
										Value: aws.String("test-instance"),
									},
									{
										Key:   aws.String("OS"),
										Value: aws.String("linux"),
									},
									{
										Key:   aws.String("ami-migrate"),
										Value: aws.String("enabled"),
									},
								},
							},
						},
					},
				},
			}, nil
		}
		// Return instance state when checking for status
		return &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-1234567890abcdef0"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameStopped,
							},
							Tags: []types.Tag{
								{
									Key:   aws.String("OS"),
									Value: aws.String("linux"),
								},
							},
						},
					},
				},
			},
		}, nil
	}

	mockClient.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
		return &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId: aws.String("ami-latest"),
					State:   types.ImageStateAvailable,
				},
			},
		}, nil
	}

	mockClient.StopInstancesFunc = func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
		return &ec2.StopInstancesOutput{
			StoppingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameStopped,
					},
					InstanceId: aws.String("i-1234567890abcdef0"),
				},
			},
		}, nil
	}

	mockClient.CreateImageFunc = func(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
		return &ec2.CreateImageOutput{
			ImageId: aws.String("ami-new"),
		}, nil
	}

	mockClient.RunInstancesFunc = func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
		return &ec2.RunInstancesOutput{
			Instances: []types.Instance{
				{
					InstanceId: aws.String("i-new"),
					State: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
				},
			},
		}, nil
	}

	mockClient.TerminateInstancesFunc = func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
		return &ec2.TerminateInstancesOutput{
			TerminatingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameTerminated,
					},
					InstanceId: aws.String("i-1234567890abcdef0"),
				},
			},
		}, nil
	}

	mockClient.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
		return &ec2.CreateTagsOutput{}, nil
	}

	err := svc.MigrateInstances(context.Background(), "enabled")
	assert.NoError(t, err)
}

func TestAMIService(t *testing.T) {
	mockClient := ecTypes.NewMockEC2Client()
	mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
		return &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-1234567890abcdef0"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
							BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
								{
									DeviceName: aws.String("/dev/xvda"),
									Ebs: &types.EbsInstanceBlockDevice{
										VolumeId: aws.String("vol-1234567890abcdef0"),
									},
								},
							},
						},
					},
				},
			},
		}, nil
	}

	service := NewService(mockClient)
	assert.NotNil(t, service)
}

func TestGetInstance(t *testing.T) {
	svc, mockClient := setupTest(t)

	// Set up mock response
	mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
		return &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-1234567890abcdef0"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				},
			},
		}, nil
	}

	// Test method
	instance, err := svc.getInstance(context.Background(), "i-1234567890abcdef0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, "i-1234567890abcdef0", *instance.InstanceId)
}
