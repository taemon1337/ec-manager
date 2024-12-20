package ami

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	apitypes "github.com/taemon1337/ec-manager/pkg/types"
)

func TestGetAMIWithTag(t *testing.T) {
	// Initialize test logger with debug level
	testutil.InitTestLogger(t)
	logger.Init(logger.LogLevel("debug"))

	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		tagKey      string
		tagValue    string
		wantAMI     string
		wantErr     bool
		errContains string
	}{
		{
			name: "found AMI with tag",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
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
				}
			},
			tagKey:   "Status",
			tagValue: "latest",
			wantAMI:  "ami-123",
			wantErr:  false,
		},
		{
			name: "no AMI found",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}
			},
			tagKey:      "Status",
			tagValue:    "latest",
			wantAMI:     "",
			wantErr:     true,
			errContains: "no AMI found with tag Status=latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Set mock client
			if err := client.SetEC2Client(mockClient); err != nil {
				t.Fatal(err)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
			gotAMI, err := svc.GetAMIWithTag(context.Background(), tt.tagKey, tt.tagValue)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAMI, gotAMI)
			}
		})
	}
}

func TestTagAMI(t *testing.T) {
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		amiID       string
		tagKey      string
		tagValue    string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful tag",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}
			},
			amiID:    "ami-123",
			tagKey:   "Status",
			tagValue: "latest",
			wantErr:  false,
		},
		{
			name: "error tagging",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.CreateTagsError = fmt.Errorf("failed to tag AMI")
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
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
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
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		instanceID  string
		newAMI      string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful migration",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
				m.StopInstancesOutput = &ec2.StopInstancesOutput{
					StoppingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameStopped,
							},
							InstanceId: aws.String("i-123"),
						},
					},
				}
				m.StartInstancesOutput = &ec2.StartInstancesOutput{
					StartingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
							InstanceId: aws.String("i-123"),
						},
					},
				}
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-456"),
							ImageId:    aws.String("ami-new"),
							State: &types.InstanceState{
								Name: types.InstanceStateNamePending,
							},
						},
					},
				}
				m.TerminateInstancesOutput = &ec2.TerminateInstancesOutput{
					TerminatingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameShuttingDown,
							},
							InstanceId: aws.String("i-123"),
						},
					},
				}
			},
			instanceID: "i-123",
			newAMI:     "ami-new",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
			instanceID:  "i-nonexistent",
			newAMI:     "ami-new",
			wantErr:    true,
			errContains: "instance not found",
		},
		{
			name: "stop instance error",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
				m.StopInstancesError = fmt.Errorf("failed to stop instance")
			},
			instanceID:  "i-123",
			newAMI:     "ami-new",
			wantErr:    true,
			errContains: "failed to stop instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Set mock client
			if err := client.SetEC2Client(mockClient); err != nil {
				t.Fatal(err)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
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
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		instanceID  string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful backup",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
				m.CreateSnapshotOutput = &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				}
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}
			},
			instanceID: "i-123",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
			instanceID:   "i-123",
			wantErr:     true,
			errContains: "instance not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
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
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		userID      string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful list",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
			},
			userID:  "user123",
			wantErr: false,
		},
		{
			name: "no instances found",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
			userID:      "user123",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
			instances, err := svc.ListUserInstances(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if mockClient.DescribeInstancesOutput != nil && len(mockClient.DescribeInstancesOutput.Reservations) > 0 {
					assert.NotEmpty(t, instances)
				} else {
					assert.Empty(t, instances)
				}
			}
		})
	}
}

func TestCreateInstance(t *testing.T) {
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		config      InstanceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "successful create",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-123"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
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
				}
			},
			config: InstanceConfig{
				Name:   "test-instance",
				OSType: "linux",
				Size:   "small",
				UserID: "user123",
			},
			wantErr: false,
		},
		{
			name: "invalid size",
			setupMock: func(m *apitypes.MockEC2Client) {
			},
			config: InstanceConfig{
				Name:   "test-instance",
				OSType: "linux",
				Size:   "invalid",
				UserID: "user123",
			},
			wantErr:     true,
			errContains: "get latest AMI: no AMI found for OS type: linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
			instance, err := svc.CreateInstance(context.Background(), tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
			}
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	// Initialize test logger
	testutil.InitTestLogger(t)
	tests := []struct {
		name        string
		setupMock   func(*apitypes.MockEC2Client)
		userID      string
		instanceID  string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful delete",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
				m.TerminateInstancesOutput = &ec2.TerminateInstancesOutput{}
			},
			userID:     "user123",
			instanceID: "i-123",
			wantErr:    false,
		},
		{
			name: "instance not found",
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
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
			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Create service with mock client
			svc := NewService(mockClient)

			// Run test
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
