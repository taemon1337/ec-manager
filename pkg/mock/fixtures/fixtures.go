package fixtures

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

// Common test constants
const (
	TestAMIID      = "ami-123"
	TestInstanceID = "i-123"
	TestVolumeID   = "vol-123"
	TestSubnetID   = "subnet-123"
	TestKeyName    = "test-key"
	TestSnapshotID = "snap-123"
)

// TestInstance returns a test EC2 instance
func TestInstance() ec2types.Instance {
	return ec2types.Instance{
		InstanceId:   aws.String(TestInstanceID),
		ImageId:      aws.String(TestAMIID),
		InstanceType: ec2types.InstanceTypeT2Micro,
		SubnetId:     aws.String(TestSubnetID),
		KeyName:      aws.String(TestKeyName),
		State: &ec2types.InstanceState{
			Name: ec2types.InstanceStateNameRunning,
		},
		LaunchTime: aws.Time(time.Now()),
		Placement: &ec2types.Placement{
			AvailabilityZone: aws.String("us-east-1a"),
		},
		Tags: []ec2types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-instance"),
			},
		},
		BlockDeviceMappings: []ec2types.InstanceBlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2types.EbsInstanceBlockDevice{
					VolumeId: aws.String(TestVolumeID),
				},
			},
		},
	}
}

// TestInstanceStopped returns a test EC2 instance in stopped state
func TestInstanceStopped() ec2types.Instance {
	instance := TestInstance()
	instance.State.Name = ec2types.InstanceStateNameStopped
	return instance
}

// TestInstanceTerminated returns a test EC2 instance in terminated state
func TestInstanceTerminated() ec2types.Instance {
	instance := TestInstance()
	instance.State.Name = ec2types.InstanceStateNameTerminated
	return instance
}

// TestAMI returns a test AMI
func TestAMI() ec2types.Image {
	return ec2types.Image{
		ImageId:      aws.String(TestAMIID),
		Name:         aws.String("test-ami"),
		Description:  aws.String("Test AMI for unit tests"),
		State:        ec2types.ImageStateAvailable,
		Architecture: "x86_64",
		Platform:     "Linux/UNIX",
		Tags: []ec2types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-ami"),
			},
			{
				Key:   aws.String("OS"),
				Value: aws.String("Linux"),
			},
		},
	}
}

// TestAMIPending returns a test AMI in pending state
func TestAMIPending() ec2types.Image {
	ami := TestAMI()
	ami.State = ec2types.ImageStatePending
	return ami
}

// TestAMIFailed returns a test AMI in failed state
func TestAMIFailed() ec2types.Image {
	ami := TestAMI()
	ami.State = ec2types.ImageStateFailed
	return ami
}

// TestRootCmd returns a test root command
func TestRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	return cmd
}

// CommonTestFlags adds common test flags to a command
func CommonTestFlags(cmd *cobra.Command) {
	cmd.Flags().String("instance", "", "Instance ID")
	cmd.Flags().String("ami", "", "AMI ID")
	cmd.Flags().String("profile", "", "AWS profile")
	cmd.Flags().String("region", "", "AWS region")
	cmd.Flags().Bool("mock", false, "Use mock mode")
}

// TestInstanceRunningWaiter returns a test instance running waiter
func TestInstanceRunningWaiter() *ec2.InstanceRunningWaiter {
	return &ec2.InstanceRunningWaiter{}
}

// TestInstanceStoppedWaiter returns a test instance stopped waiter
func TestInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter {
	return &ec2.InstanceStoppedWaiter{}
}

// TestInstanceTerminatedWaiter returns a test instance terminated waiter
func TestInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter {
	return &ec2.InstanceTerminatedWaiter{}
}

// TestVolumeAvailableWaiter returns a test volume available waiter
func TestVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter {
	return &ec2.VolumeAvailableWaiter{}
}

// TestSnapshot returns a test snapshot
func TestSnapshot() ec2types.Snapshot {
	return ec2types.Snapshot{
		SnapshotId: aws.String(TestSnapshotID),
		VolumeId:   aws.String(TestVolumeID),
		State:      ec2types.SnapshotStateCompleted,
		Tags: []ec2types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-snapshot"),
			},
			{
				Key:   aws.String("ami-migrate-device"),
				Value: aws.String("/dev/xvdf"),
			},
		},
	}
}

// TestListAMIs returns a list of test AMIs
func TestListAMIs() []ec2types.Image {
	return []ec2types.Image{
		{
			ImageId:      aws.String("ami-123"),
			Name:         aws.String("test-ami-1"),
			Description:  aws.String("Test AMI 1 for unit tests"),
			State:        ec2types.ImageStateAvailable,
			Architecture: ec2types.ArchitectureValuesX8664,
			Platform:     "Linux/UNIX",
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-ami-1"),
				},
				{
					Key:   aws.String("OS"),
					Value: aws.String("Linux"),
				},
			},
		},
		{
			ImageId:      aws.String("ami-456"),
			Name:         aws.String("test-ami-2"),
			Description:  aws.String("Test AMI 2 for unit tests"),
			State:        ec2types.ImageStateAvailable,
			Architecture: ec2types.ArchitectureValuesX8664,
			Platform:     "Windows",
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-ami-2"),
				},
				{
					Key:   aws.String("OS"),
					Value: aws.String("Windows"),
				},
			},
		},
	}
}

// TestListInstances returns a list of test instances
func TestListInstances() []ec2types.Instance {
	return []ec2types.Instance{
		{
			InstanceId:   aws.String("i-123"),
			ImageId:      aws.String("ami-123"),
			InstanceType: ec2types.InstanceTypeT2Micro,
			SubnetId:     aws.String("subnet-123"),
			KeyName:      aws.String("test-key-1"),
			State: &ec2types.InstanceState{
				Name: ec2types.InstanceStateNameRunning,
			},
			LaunchTime: aws.Time(time.Now()),
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-instance-1"),
				},
			},
		},
		{
			InstanceId:   aws.String("i-456"),
			ImageId:      aws.String("ami-456"),
			InstanceType: ec2types.InstanceTypeT2Small,
			SubnetId:     aws.String("subnet-456"),
			KeyName:      aws.String("test-key-2"),
			State: &ec2types.InstanceState{
				Name: ec2types.InstanceStateNameStopped,
			},
			LaunchTime: aws.Time(time.Now()),
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-instance-2"),
				},
			},
		},
	}
}

// TestListSubnets returns a list of test subnets
func TestListSubnets() []ec2types.Subnet {
	return []ec2types.Subnet{
		{
			SubnetId:         aws.String("subnet-123"),
			VpcId:            aws.String("vpc-123"),
			CidrBlock:        aws.String("10.0.1.0/24"),
			AvailabilityZone: aws.String("us-east-1a"),
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-subnet-1"),
				},
			},
		},
		{
			SubnetId:         aws.String("subnet-456"),
			VpcId:            aws.String("vpc-123"),
			CidrBlock:        aws.String("10.0.2.0/24"),
			AvailabilityZone: aws.String("us-east-1b"),
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-subnet-2"),
				},
			},
		},
	}
}

// TestListKeyPairs returns a list of test key pairs
func TestListKeyPairs() []ec2types.KeyPairInfo {
	return []ec2types.KeyPairInfo{
		{
			KeyName:        aws.String("test-key-1"),
			KeyFingerprint: aws.String("12:34:56:78:90:ab:cd:ef"),
			KeyType:        ec2types.KeyTypeEd25519,
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-key-1"),
				},
			},
		},
		{
			KeyName:        aws.String("test-key-2"),
			KeyFingerprint: aws.String("98:76:54:32:10:fe:dc:ba"),
			KeyType:        ec2types.KeyTypeRsa,
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-key-2"),
				},
			},
		},
	}
}
