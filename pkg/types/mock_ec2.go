package types

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MockEC2Client is a mock implementation of EC2ClientAPI
type MockEC2Client struct {
	sync.Mutex
	// Output and error fields for each operation
	DescribeInstancesOutput *ec2.DescribeInstancesOutput
	DescribeInstancesError  error
	DescribeImagesOutput   *ec2.DescribeImagesOutput
	DescribeImagesError    error
	RunInstancesOutput     *ec2.RunInstancesOutput
	RunInstancesError      error
	StopInstancesOutput    *ec2.StopInstancesOutput
	StopInstancesError     error
	StartInstancesOutput   *ec2.StartInstancesOutput
	StartInstancesError    error
	CreateTagsOutput       *ec2.CreateTagsOutput
	CreateTagsError        error
	TerminateInstancesOutput *ec2.TerminateInstancesOutput
	TerminateInstancesError  error
	CreateSnapshotOutput    *ec2.CreateSnapshotOutput
	CreateSnapshotError     error
	DescribeSnapshotsOutput *ec2.DescribeSnapshotsOutput
	DescribeSnapshotsError  error
	CreateVolumeOutput      *ec2.CreateVolumeOutput
	CreateVolumeError       error
	DescribeVolumesOutput   *ec2.DescribeVolumesOutput
	DescribeVolumesError    error
	AttachVolumeOutput      *ec2.AttachVolumeOutput
	AttachVolumeError       error

	// Data fields for convenience
	Images    []types.Image
	Instances []types.Instance
	Instance  *types.Instance
	Snapshots []types.Snapshot
	Volumes   []types.Volume

	// Track instance states for waiters
	instanceStates map[string]types.InstanceStateName
}

// NewMockEC2Client creates a new mock EC2 client
func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{
		instanceStates: make(map[string]types.InstanceStateName),
	}
}

// DescribeInstances implements EC2ClientAPI
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.DescribeInstancesError != nil {
		return nil, m.DescribeInstancesError
	}
	if m.DescribeInstancesOutput != nil {
		return m.DescribeInstancesOutput, nil
	}

	// If we're looking for specific instances (waiter case)
	if len(params.InstanceIds) > 0 {
		instances := make([]types.Instance, 0, len(params.InstanceIds))
		for _, id := range params.InstanceIds {
			if state, exists := m.instanceStates[id]; exists {
				instances = append(instances, types.Instance{
					InstanceId: aws.String(id),
					State: &types.InstanceState{
						Name: state,
					},
				})
			}
		}
		return &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: instances,
				},
			},
		}, nil
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: m.Instances,
			},
		},
	}, nil
}

// DescribeImages implements EC2ClientAPI
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesError != nil {
		return nil, m.DescribeImagesError
	}
	if m.DescribeImagesOutput != nil {
		return m.DescribeImagesOutput, nil
	}
	// Use the data fields if output is not set
	if len(m.Images) > 0 {
		return &ec2.DescribeImagesOutput{
			Images: m.Images,
		}, nil
	}
	return &ec2.DescribeImagesOutput{}, nil
}

// RunInstances implements EC2ClientAPI
func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesError != nil {
		return nil, m.RunInstancesError
	}
	return m.RunInstancesOutput, nil
}

// StopInstances implements EC2ClientAPI
func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.StopInstancesError != nil {
		return nil, m.StopInstancesError
	}
	if m.StopInstancesOutput != nil {
		return m.StopInstancesOutput, nil
	}

	// Update instance state in the mock data
	if len(params.InstanceIds) > 0 {
		m.setInstanceStateWithLock(params.InstanceIds[0], types.InstanceStateNameStopped)
	}

	return &ec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
				InstanceId: aws.String(params.InstanceIds[0]),
			},
		},
	}, nil
}

// StartInstances implements EC2ClientAPI
func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.StartInstancesError != nil {
		return nil, m.StartInstancesError
	}
	if m.StartInstancesOutput != nil {
		return m.StartInstancesOutput, nil
	}

	// Update instance state in the mock data
	if len(params.InstanceIds) > 0 {
		m.setInstanceStateWithLock(params.InstanceIds[0], types.InstanceStateNameRunning)
	}

	return &ec2.StartInstancesOutput{
		StartingInstances: []types.InstanceStateChange{
			{
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
				InstanceId: aws.String(params.InstanceIds[0]),
			},
		},
	}, nil
}

// CreateTags implements EC2ClientAPI
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsError != nil {
		return nil, m.CreateTagsError
	}
	return m.CreateTagsOutput, nil
}

// TerminateInstances implements EC2ClientAPI
func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.TerminateInstancesError != nil {
		return nil, m.TerminateInstancesError
	}
	if m.TerminateInstancesOutput != nil {
		return m.TerminateInstancesOutput, nil
	}

	// Update instance state in the mock data
	if len(params.InstanceIds) > 0 {
		m.setInstanceStateWithLock(params.InstanceIds[0], types.InstanceStateNameTerminated)
	}

	return &ec2.TerminateInstancesOutput{
		TerminatingInstances: []types.InstanceStateChange{
			{
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameTerminated,
				},
				InstanceId: aws.String(params.InstanceIds[0]),
			},
		},
	}, nil
}

// CreateSnapshot implements EC2ClientAPI
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotError != nil {
		return nil, m.CreateSnapshotError
	}
	return m.CreateSnapshotOutput, nil
}

// DescribeSnapshots implements EC2ClientAPI
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.DescribeSnapshotsError != nil {
		return nil, m.DescribeSnapshotsError
	}
	if m.DescribeSnapshotsOutput != nil {
		return m.DescribeSnapshotsOutput, nil
	}
	// Use the data fields if output is not set
	if len(m.Snapshots) > 0 {
		return &ec2.DescribeSnapshotsOutput{
			Snapshots: m.Snapshots,
		}, nil
	}
	return &ec2.DescribeSnapshotsOutput{}, nil
}

// CreateVolume implements EC2ClientAPI
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeError != nil {
		return nil, m.CreateVolumeError
	}
	return m.CreateVolumeOutput, nil
}

// DescribeVolumes implements EC2ClientAPI
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.DescribeVolumesError != nil {
		return nil, m.DescribeVolumesError
	}
	if m.DescribeVolumesOutput != nil {
		return m.DescribeVolumesOutput, nil
	}
	// Use the data fields if output is not set
	if len(m.Volumes) > 0 {
		return &ec2.DescribeVolumesOutput{
			Volumes: m.Volumes,
		}, nil
	}
	return &ec2.DescribeVolumesOutput{}, nil
}

// AttachVolume implements EC2ClientAPI
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeError != nil {
		return nil, m.AttachVolumeError
	}
	return m.AttachVolumeOutput, nil
}

// setInstanceStateWithLock sets the state of an instance in both the main instance list and the waiter state map
func (m *MockEC2Client) setInstanceStateWithLock(instanceID string, state types.InstanceStateName) {
	// Update instance in the main list
	for i, instance := range m.Instances {
		if aws.ToString(instance.InstanceId) == instanceID {
			m.Instances[i].State = &types.InstanceState{
				Name: state,
			}
			break
		}
	}

	// Update state for waiters
	m.instanceStates[instanceID] = state
}

// SetInstanceState sets the state of an instance in both the main instance list and the waiter state map
func (m *MockEC2Client) SetInstanceState(instanceID string, state types.InstanceStateName) {
	m.Lock()
	defer m.Unlock()
	m.setInstanceStateWithLock(instanceID, state)
}
