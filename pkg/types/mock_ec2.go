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
	InstanceStates map[string]types.InstanceStateName
}

// NewMockEC2Client creates a new mock EC2 client
func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{
		InstanceStates: make(map[string]types.InstanceStateName),
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
		// Update the instance states in the output based on our tracked states
		for i, reservation := range m.DescribeInstancesOutput.Reservations {
			for j, instance := range reservation.Instances {
				if instance.InstanceId != nil {
					if state, exists := m.InstanceStates[*instance.InstanceId]; exists {
						m.DescribeInstancesOutput.Reservations[i].Instances[j].State = &types.InstanceState{
							Name: state,
						}
					}
				}
			}
		}
		return m.DescribeInstancesOutput, nil
	}

	// Default behavior
	instances := make([]types.Instance, 0)
	if params.InstanceIds != nil {
		for _, id := range params.InstanceIds {
			state := types.InstanceStateNameRunning
			if s, exists := m.InstanceStates[id]; exists {
				state = s
			}
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

// DescribeImages implements EC2ClientAPI
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.DescribeImagesError != nil {
		return nil, m.DescribeImagesError
	}
	if m.DescribeImagesOutput != nil {
		return m.DescribeImagesOutput, nil
	}

	return &ec2.DescribeImagesOutput{
		Images: m.Images,
	}, nil
}

// RunInstances mocks the RunInstances operation
func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.RunInstancesError != nil {
		return nil, m.RunInstancesError
	}

	if m.RunInstancesOutput != nil {
		// Update instance state to running for all instances
		for _, instance := range m.RunInstancesOutput.Instances {
			if instance.InstanceId != nil {
				m.setInstanceStateWithLock(*instance.InstanceId, types.InstanceStateNameRunning)
			}
		}
		return m.RunInstancesOutput, nil
	}

	// Default behavior
	instanceID := "i-456"
	m.setInstanceStateWithLock(instanceID, types.InstanceStateNameRunning)
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String(instanceID),
				ImageId:    params.ImageId,
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil
}

// StopInstances mocks the StopInstances operation
func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.StopInstancesError != nil {
		return nil, m.StopInstancesError
	}

	if m.StopInstancesOutput != nil {
		// Update instance state to stopped
		for _, instanceID := range params.InstanceIds {
			m.setInstanceStateWithLock(instanceID, types.InstanceStateNameStopped)
		}
		return m.StopInstancesOutput, nil
	}

	// Default behavior if no output is set
	stoppingInstances := make([]types.InstanceStateChange, 0, len(params.InstanceIds))
	for _, instanceID := range params.InstanceIds {
		m.setInstanceStateWithLock(instanceID, types.InstanceStateNameStopped)
		stoppingInstances = append(stoppingInstances, types.InstanceStateChange{
			CurrentState: &types.InstanceState{
				Name: types.InstanceStateNameStopped,
			},
			InstanceId: aws.String(instanceID),
			PreviousState: &types.InstanceState{
				Name: types.InstanceStateNameRunning,
			},
		})
	}

	return &ec2.StopInstancesOutput{
		StoppingInstances: stoppingInstances,
	}, nil
}

// StartInstances implements EC2ClientAPI
func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.StartInstancesError != nil {
		return nil, m.StartInstancesError
	}

	// Update instance states
	for _, id := range params.InstanceIds {
		m.setInstanceStateWithLock(id, types.InstanceStateNameRunning)
	}

	if m.StartInstancesOutput != nil {
		return m.StartInstancesOutput, nil
	}

	// Create default response
	startingInstances := make([]types.InstanceStateChange, 0, len(params.InstanceIds))
	for _, id := range params.InstanceIds {
		startingInstances = append(startingInstances, types.InstanceStateChange{
			CurrentState: &types.InstanceState{
				Name: types.InstanceStateNameRunning,
			},
			InstanceId: aws.String(id),
			PreviousState: &types.InstanceState{
				Name: types.InstanceStateNameStopped,
			},
		})
	}

	return &ec2.StartInstancesOutput{
		StartingInstances: startingInstances,
	}, nil
}

// CreateTags implements EC2ClientAPI
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.CreateTagsError != nil {
		return nil, m.CreateTagsError
	}
	return m.CreateTagsOutput, nil
}

// TerminateInstances mocks the TerminateInstances operation
func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.TerminateInstancesError != nil {
		return nil, m.TerminateInstancesError
	}

	if m.TerminateInstancesOutput != nil {
		// Update instance state to terminated
		for _, instanceID := range params.InstanceIds {
			m.setInstanceStateWithLock(instanceID, types.InstanceStateNameTerminated)
		}
		return m.TerminateInstancesOutput, nil
	}

	// Default behavior
	terminatingInstances := make([]types.InstanceStateChange, 0, len(params.InstanceIds))
	for _, instanceID := range params.InstanceIds {
		m.setInstanceStateWithLock(instanceID, types.InstanceStateNameTerminated)
		terminatingInstances = append(terminatingInstances, types.InstanceStateChange{
			CurrentState: &types.InstanceState{
				Name: types.InstanceStateNameTerminated,
			},
			InstanceId: aws.String(instanceID),
			PreviousState: &types.InstanceState{
				Name: types.InstanceStateNameRunning,
			},
		})
	}

	return &ec2.TerminateInstancesOutput{
		TerminatingInstances: terminatingInstances,
	}, nil
}

// CreateSnapshot implements EC2ClientAPI
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.CreateSnapshotError != nil {
		return nil, m.CreateSnapshotError
	}
	return m.CreateSnapshotOutput, nil
}

// DescribeSnapshots implements EC2ClientAPI
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.DescribeSnapshotsError != nil {
		return nil, m.DescribeSnapshotsError
	}
	if m.DescribeSnapshotsOutput != nil {
		return m.DescribeSnapshotsOutput, nil
	}

	return &ec2.DescribeSnapshotsOutput{
		Snapshots: m.Snapshots,
	}, nil
}

// CreateVolume implements EC2ClientAPI
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.CreateVolumeError != nil {
		return nil, m.CreateVolumeError
	}
	return m.CreateVolumeOutput, nil
}

// DescribeVolumes implements EC2ClientAPI
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.DescribeVolumesError != nil {
		return nil, m.DescribeVolumesError
	}
	if m.DescribeVolumesOutput != nil {
		return m.DescribeVolumesOutput, nil
	}

	return &ec2.DescribeVolumesOutput{
		Volumes: m.Volumes,
	}, nil
}

// AttachVolume implements EC2ClientAPI
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	m.Lock()
	defer m.Unlock()

	if m.AttachVolumeError != nil {
		return nil, m.AttachVolumeError
	}
	return m.AttachVolumeOutput, nil
}

// GetInstanceState returns the current state of an instance
func (m *MockEC2Client) GetInstanceState(instanceID string) types.InstanceStateName {
	m.Lock()
	defer m.Unlock()
	
	if state, exists := m.InstanceStates[instanceID]; exists {
		return state
	}
	return types.InstanceStateNamePending // Default state if not found
}

// setInstanceStateWithLock sets the state of an instance in both the main instance list and the waiter state map
func (m *MockEC2Client) setInstanceStateWithLock(instanceID string, state types.InstanceStateName) {
	// Update instance state map
	m.InstanceStates[instanceID] = state

	// Update instance in the main list
	for i, instance := range m.Instances {
		if aws.ToString(instance.InstanceId) == instanceID {
			m.Instances[i].State = &types.InstanceState{
				Name: state,
			}
			break
		}
	}
}

// SetInstanceState sets the state of an instance in both the main instance list and the waiter state map
func (m *MockEC2Client) SetInstanceState(instanceID string, state types.InstanceStateName) {
	m.Lock()
	defer m.Unlock()
	m.setInstanceStateWithLock(instanceID, state)
}
