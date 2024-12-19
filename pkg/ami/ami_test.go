package ami

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/taemon1337/ami-migrate/pkg/ami/mocks"
)

type mockEC2Client struct {
	describeImagesFunc     func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	createTagsFunc         func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	describeInstancesFunc  func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	createSnapshotFunc     func(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	terminateInstancesFunc func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	runInstancesFunc       func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
}

func (m *mockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m.describeImagesFunc(ctx, params, optFns...)
}

func (m *mockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.createTagsFunc(ctx, params, optFns...)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesFunc(ctx, params, optFns...)
}

func (m *mockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	return m.createSnapshotFunc(ctx, params, optFns...)
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return m.terminateInstancesFunc(ctx, params, optFns...)
}

func (m *mockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.runInstancesFunc(ctx, params, optFns...)
}

func TestGetAMIWithTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		mock    *mocks.MockEC2Client
		want    string
		wantErr bool
	}{
		{
			name: "successful retrieval",
			tag:  "latest",
			mock: &mocks.MockEC2Client{
				DescribeImagesOutput: &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-123"),
						},
					},
				},
			},
			want:    "ami-123",
			wantErr: false,
		},
		{
			name: "no images found",
			tag:  "latest",
			mock: &mocks.MockEC2Client{
				DescribeImagesOutput: &ec2.DescribeImagesOutput{
					Images: []types.Image{},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "aws error",
			tag:  "latest",
			mock: &mocks.MockEC2Client{
				DescribeImagesError: errors.New("aws error"),
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(tt.mock)
			got, err := s.GetAMIWithTag(context.Background(), tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAMIWithTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAMIWithTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateAMITags(t *testing.T) {
	tests := []struct {
		name    string
		oldAMI  string
		newAMI  string
		mock    *mocks.MockEC2Client
		wantErr bool
	}{
		{
			name:   "successful update",
			oldAMI: "ami-old",
			newAMI: "ami-new",
			mock: &mocks.MockEC2Client{
				CreateTagsOutput: &ec2.CreateTagsOutput{},
			},
			wantErr: false,
		},
		{
			name:   "error updating old AMI",
			oldAMI: "ami-old",
			newAMI: "ami-new",
			mock: &mocks.MockEC2Client{
				CreateTagsError: errors.New("aws error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(tt.mock)
			err := s.UpdateAMITags(context.Background(), tt.oldAMI, tt.newAMI)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAMITags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrateInstances(t *testing.T) {
	tests := []struct {
		name    string
		oldAMI  string
		newAMI  string
		mock    *mocks.MockEC2Client
		wantErr bool
	}{
		{
			name:   "successful migration",
			oldAMI: "ami-old",
			newAMI: "ami-new",
			mock: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
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
				CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				},
				TerminateInstancesOutput: &ec2.TerminateInstancesOutput{},
				RunInstancesOutput: &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-new"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "no instances found",
			oldAMI: "ami-old",
			newAMI: "ami-new",
			mock: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			wantErr: false,
		},
		{
			name:   "error describing instances",
			oldAMI: "ami-old",
			newAMI: "ami-new",
			mock: &mocks.MockEC2Client{
				DescribeInstancesError: errors.New("aws error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(tt.mock)
			err := s.MigrateInstances(context.Background(), tt.oldAMI, tt.newAMI)
			if (err != nil) != tt.wantErr {
				t.Errorf("MigrateInstances() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
