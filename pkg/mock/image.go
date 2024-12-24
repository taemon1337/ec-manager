package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// CreateImage implements EC2Client
func (m *MockEC2Client) CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, params, optFns...)
	}
	return m.CreateImageOutput, nil
}

// DescribeImages implements EC2Client
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesFunc != nil {
		return m.DescribeImagesFunc(ctx, params, optFns...)
	}
	return m.DescribeImagesOutput, nil
}

// CreateTags implements EC2Client
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsFunc != nil {
		return m.CreateTagsFunc(ctx, params, optFns...)
	}
	return m.CreateTagsOutput, nil
}
