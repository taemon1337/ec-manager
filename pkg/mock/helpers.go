package mock

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	"github.com/taemon1337/ec-manager/pkg/mock/fixtures"
)

// SetupSuccessfulAMICreation sets up mock expectations for a successful AMI creation
func SetupSuccessfulAMICreation(m *MockEC2Client, instanceID, name, description string) {
	// Mock instance lookup
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					fixtures.TestInstance(),
				},
			},
		},
	}, nil).Once()

	// Mock image creation
	m.On("CreateImage", mock.Anything, &ec2.CreateImageInput{
		InstanceId:  aws.String(instanceID),
		Name:        aws.String(name),
		Description: aws.String(description),
	}).Return(&ec2.CreateImageOutput{
		ImageId: aws.String(fixtures.TestAMIID),
	}, nil).Once()

	// Mock tagging
	m.On("CreateTags", mock.Anything, &ec2.CreateTagsInput{
		Resources: []string{fixtures.TestAMIID},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}).Return(&ec2.CreateTagsOutput{}, nil).Once()
}

// SetupFailedAMICreation sets up mock expectations for a failed AMI creation
func SetupFailedAMICreation(m *MockEC2Client, instanceID, name, description, errorMsg string) {
	// Mock instance lookup
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					fixtures.TestInstance(),
				},
			},
		},
	}, nil).Once()

	// Mock image creation failure
	m.On("CreateImage", mock.Anything, &ec2.CreateImageInput{
		InstanceId:  aws.String(instanceID),
		Name:        aws.String(name),
		Description: aws.String(description),
	}).Return(nil, errors.New(errorMsg)).Once()
}

// SetupInstanceNotFound sets up mock expectations for when an instance is not found
func SetupInstanceNotFound(m *MockEC2Client, instanceID string) {
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{},
	}, nil).Once()
}

// SetupSuccessfulInstanceStart sets up mock expectations for a successful instance start
func SetupSuccessfulInstanceStart(m *MockEC2Client, instanceID string) {
	m.On("StartInstances", mock.Anything, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.StartInstancesOutput{
		StartingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNamePending,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceRunningWaiter").Return(&ec2.InstanceRunningWaiter{}).Once()
}

// SetupSuccessfulInstanceStop sets up mock expectations for a successful instance stop
func SetupSuccessfulInstanceStop(m *MockEC2Client, instanceID string) {
	m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopping,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceStoppedWaiter").Return(&ec2.InstanceStoppedWaiter{}).Once()
}

// SetupSuccessfulBackup sets up mock expectations for a successful backup
func SetupSuccessfulBackup(m *MockEC2Client, instanceID string) {
	// Mock instance lookup
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					fixtures.TestInstance(),
				},
			},
		},
	}, nil).Once()

	// Mock snapshot creation
	m.On("CreateSnapshot", mock.Anything, &ec2.CreateSnapshotInput{
		VolumeId: aws.String("vol-123test"),
	}).Return(&ec2.CreateSnapshotOutput{
		SnapshotId: aws.String("snap-123test"),
	}, nil).Once()

	m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopping,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceStoppedWaiter").Return(&ec2.InstanceStoppedWaiter{}).Once()
}

// SetupSuccessfulMigration sets up mock expectations for a successful migration
func SetupSuccessfulMigration(m *MockEC2Client, instanceID, amiID string) {
	// Mock instance lookup
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					fixtures.TestInstance(),
				},
			},
		},
	}, nil).Once()

	// Mock AMI lookup
	m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{
			fixtures.TestAMI(),
		},
	}, nil).Once()

	// Mock instance creation
	m.On("RunInstances", mock.Anything, mock.MatchedBy(func(input *ec2.RunInstancesInput) bool {
		return *input.ImageId == amiID
	})).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("i-456test"),
			},
		},
	}, nil).Once()

	// Mock instance tagging
	m.On("CreateTags", mock.Anything, mock.MatchedBy(func(input *ec2.CreateTagsInput) bool {
		return input.Resources[0] == "i-456test"
	})).Return(&ec2.CreateTagsOutput{}, nil).Once()

	// Mock instance stop
	m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopping,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceStoppedWaiter").Return(&ec2.InstanceStoppedWaiter{}).Once()

	// Mock instance terminate
	m.On("TerminateInstances", mock.Anything, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.TerminateInstancesOutput{
		TerminatingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameShuttingDown,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceTerminatedWaiter").Return(&ec2.InstanceTerminatedWaiter{}).Once()

	// Mock instance start
	m.On("StartInstances", mock.Anything, &ec2.StartInstancesInput{
		InstanceIds: []string{"i-456test"},
	}).Return(&ec2.StartInstancesOutput{
		StartingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String("i-456test"),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNamePending,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceRunningWaiter").Return(&ec2.InstanceRunningWaiter{}).Once()
}

// SetupSuccessfulRestore sets up mock expectations for a successful restore
func SetupSuccessfulRestore(m *MockEC2Client, instanceID, amiID string) {
	// Mock instance lookup
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					fixtures.TestInstance(),
				},
			},
		},
	}, nil).Once()

	// Mock AMI lookup
	m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{
			fixtures.TestAMI(),
		},
	}, nil).Once()

	// Mock instance creation
	m.On("RunInstances", mock.Anything, mock.MatchedBy(func(input *ec2.RunInstancesInput) bool {
		return *input.ImageId == amiID
	})).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("i-456test"),
			},
		},
	}, nil).Once()

	// Mock instance tagging
	m.On("CreateTags", mock.Anything, mock.MatchedBy(func(input *ec2.CreateTagsInput) bool {
		return input.Resources[0] == "i-456test"
	})).Return(&ec2.CreateTagsOutput{}, nil).Once()

	// Mock instance stop
	m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String(instanceID),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopping,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceStoppedWaiter").Return(&ec2.InstanceStoppedWaiter{}).Once()

	// Mock instance start
	m.On("StartInstances", mock.Anything, &ec2.StartInstancesInput{
		InstanceIds: []string{"i-456test"},
	}).Return(&ec2.StartInstancesOutput{
		StartingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String("i-456test"),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNamePending,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
			},
		},
	}, nil).Once()

	m.On("NewInstanceRunningWaiter").Return(&ec2.InstanceRunningWaiter{}).Once()
}

// SetupAMINotFound sets up mock expectations for when an AMI is not found
func SetupAMINotFound(m *MockEC2Client, amiID string) {
	m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{},
	}, nil).Once()
}

// SetupFailedInstanceStart sets up mock expectations for a failed instance start
func SetupFailedInstanceStart(m *MockEC2Client, instanceID string) {
	m.On("StartInstances", mock.Anything, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(nil, errors.New("failed to start instance")).Once()
}

// SetupFailedInstanceStop sets up mock expectations for a failed instance stop
func SetupFailedInstanceStop(m *MockEC2Client, instanceID string) {
	m.On("StopInstances", mock.Anything, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(nil, errors.New("failed to stop instance")).Once()
}

// SetupFailedSnapshot sets up mock expectations for a failed snapshot creation
func SetupFailedSnapshot(m *MockEC2Client, volumeID string) {
	m.On("CreateSnapshot", mock.Anything, &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeID),
	}).Return(nil, errors.New("failed to create snapshot")).Once()
}

// SetupFailedInstanceLaunch sets up mock expectations for a failed instance launch
func SetupFailedInstanceLaunch(m *MockEC2Client, amiID string) {
	m.On("RunInstances", mock.Anything, mock.MatchedBy(func(input *ec2.RunInstancesInput) bool {
		return *input.ImageId == amiID
	})).Return(nil, errors.New("failed to launch instance")).Once()
}

// NewInstanceRunningWaiter returns a mock running waiter
func NewInstanceRunningWaiter() *ec2.InstanceRunningWaiter {
	return &ec2.InstanceRunningWaiter{}
}

// NewInstanceStoppedWaiter returns a mock stopped waiter
func NewInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter {
	return &ec2.InstanceStoppedWaiter{}
}

// NewInstanceTerminatedWaiter returns a mock terminated waiter
func NewInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter {
	return &ec2.InstanceTerminatedWaiter{}
}

// NewVolumeAvailableWaiter returns a mock volume available waiter
func NewVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter {
	return &ec2.VolumeAvailableWaiter{}
}
