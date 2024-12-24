package mock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// CreateSnapshot implements EC2Client
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotFunc != nil {
		return m.CreateSnapshotFunc(ctx, params, optFns...)
	}
	return m.CreateSnapshotOutput, nil
}

// DescribeSnapshots implements EC2Client
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.DescribeSnapshotsFunc != nil {
		return m.DescribeSnapshotsFunc(ctx, params, optFns...)
	}

	// Mock snapshot not found for specific ID
	if len(params.SnapshotIds) > 0 && params.SnapshotIds[0] != "snap-123" {
		return nil, fmt.Errorf("snapshot not found: %s", params.SnapshotIds[0])
	}

	return m.DescribeSnapshotsOutput, nil
}
