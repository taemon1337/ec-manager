package ami

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	localmock "github.com/taemon1337/ec-manager/pkg/mock"
)

// AMITestSuite defines a test suite for AMI-related tests
type AMITestSuite struct {
	suite.Suite
	mockClient *localmock.MockEC2Client
	service    *Service
}

// SetupTest runs before each test
func (s *AMITestSuite) SetupTest() {
	s.mockClient = localmock.NewMockEC2Client()
	s.service = NewService(s.mockClient)
}

// TearDownTest runs after each test
func (s *AMITestSuite) TearDownTest() {
	s.mockClient.AssertExpectations(s.T())
}

// TestAMISuite runs the test suite
func TestAMISuite(t *testing.T) {
	suite.Run(t, new(AMITestSuite))
}

func (s *AMITestSuite) TestCreateAMI() {
	tests := []struct {
		name          string
		instanceID    string
		amiName       string
		description   string
		mockSetup     func()
		expectedError error
		assertMock    func()
	}{
		{
			name:        "success",
			instanceID:  "i-123",
			amiName:     "test-ami",
			description: "test description",
			mockSetup: func() {
				// Mock DescribeInstances
				s.mockClient.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
								},
							},
						},
					},
				}
				s.mockClient.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(s.mockClient.DescribeInstancesOutput, nil).Once()

				// Mock CreateImage
				s.mockClient.CreateImageOutput = &ec2.CreateImageOutput{
					ImageId: aws.String("ami-123"),
				}
				s.mockClient.On("CreateImage", mock.Anything, &ec2.CreateImageInput{
					InstanceId:  aws.String("i-123"),
					Name:        aws.String("test-ami"),
					Description: aws.String("test description"),
				}).Return(s.mockClient.CreateImageOutput, nil).Once()

				// Mock DescribeImages
				s.mockClient.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-123"),
						},
					},
				}
				s.mockClient.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-123"},
				}).Return(s.mockClient.DescribeImagesOutput, nil).Once()

				// Mock CreateTags
				s.mockClient.CreateTagsOutput = &ec2.CreateTagsOutput{}
				s.mockClient.On("CreateTags", mock.Anything, &ec2.CreateTagsInput{
					Resources: []string{"ami-123"},
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-ami"),
						},
					},
				}).Return(s.mockClient.CreateTagsOutput, nil).Once()
			},
		},
		{
			name:        "instance_not_found",
			instanceID:  "i-123",
			amiName:     "test-ami",
			description: "test description",
			mockSetup: func() {
				// Mock DescribeInstances to return empty reservations
				s.mockClient.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
				s.mockClient.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(s.mockClient.DescribeInstancesOutput, nil).Once()
			},
			expectedError: ErrInstanceNotFound,
		},
		{
			name:        "create_image_error",
			instanceID:  "i-123",
			amiName:     "test-ami",
			description: "test description",
			mockSetup: func() {
				// Mock DescribeInstances to return a valid instance
				s.mockClient.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
								},
							},
						},
					},
				}
				s.mockClient.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(s.mockClient.DescribeInstancesOutput, nil).Once()

				// Mock CreateImage to return an error
				s.mockClient.CreateImageOutput = nil
				s.mockClient.On("CreateImage", mock.Anything, &ec2.CreateImageInput{
					InstanceId:  aws.String("i-123"),
					Name:        aws.String("test-ami"),
					Description: aws.String("test description"),
				}).Return(s.mockClient.CreateImageOutput, ErrCreateImageFailed).Once()
			},
			expectedError: ErrCreateImageFailed,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockSetup()

			_, err := s.service.CreateAMI(context.Background(), tt.instanceID, tt.amiName, tt.description)
			if tt.expectedError != nil {
				if err == nil {
					s.Fail("Expected error but got nil")
					return
				}
				s.ErrorIs(err, tt.expectedError)
			} else {
				s.NoError(err)
			}
			if tt.assertMock != nil {
				tt.assertMock()
			}
		})
	}
}

func (s *AMITestSuite) TestFindAMI() {
	tests := []struct {
		name      string
		amiName   string
		setupMock func(*localmock.MockEC2Client)
		want      *types.Image
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-123"),
						},
					},
				}
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					Filters: []types.Filter{
						{
							Name: aws.String("name"),
							Values: []string{
								"test-ami",
							},
						},
					},
				}).Return(m.DescribeImagesOutput, nil).Once()
			},
			want: &types.Image{
				ImageId: aws.String("ami-123"),
			},
			wantErr: false,
		},
		{
			name:    "ami_not_found",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					Filters: []types.Filter{
						{
							Name: aws.String("name"),
							Values: []string{
								"test-ami",
							},
						},
					},
				}).Return(m.DescribeImagesOutput, nil).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "AMI not found",
		},
		{
			name:    "describe_images_failure",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					Filters: []types.Filter{
						{
							Name: aws.String("name"),
							Values: []string{
								"test-ami",
							},
						},
					},
				}).Return(nil, fmt.Errorf("failed to describe images")).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "AMI not found",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock(s.mockClient)
			ami, err := s.service.FindAMI(context.Background(), tt.amiName)
			if tt.wantErr {
				s.Error(err)
				if tt.errMsg != "" {
					s.Contains(err.Error(), tt.errMsg)
				}
			} else {
				s.NoError(err)
				if tt.want != nil {
					s.Equal(tt.want.ImageId, ami.ImageId)
				}
			}
		})
	}
}

func (s *AMITestSuite) TestGetImage() {
	tests := []struct {
		name      string
		imageID   string
		setupMock func(*localmock.MockEC2Client)
		want      *types.Image
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success",
			imageID: "ami-123",
			setupMock: func(m *localmock.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-123"),
							Name:    aws.String("test-ami"),
						},
					},
				}
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-123"},
				}).Return(m.DescribeImagesOutput, nil).Once()
			},
			want: &types.Image{
				ImageId: aws.String("ami-123"),
				Name:    aws.String("test-ami"),
			},
			wantErr: false,
		},
		{
			name:    "ami_not_found",
			imageID: "ami-999",
			setupMock: func(m *localmock.MockEC2Client) {
				m.DescribeImagesOutput = &ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-999"},
				}).Return(m.DescribeImagesOutput, nil).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "AMI not found: ami-999",
		},
		{
			name:    "describe_images_error",
			imageID: "ami-123",
			setupMock: func(m *localmock.MockEC2Client) {
				m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
					ImageIds: []string{"ami-123"},
				}).Return(nil, fmt.Errorf("failed to describe images")).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "AMI not found: ami-123",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock(s.mockClient)
			ami, err := s.service.GetImage(context.Background(), tt.imageID)
			if tt.wantErr {
				s.Error(err)
				if tt.errMsg != "" {
					s.Contains(err.Error(), tt.errMsg)
				}
			} else {
				s.NoError(err)
				s.Equal(tt.want.ImageId, ami.ImageId)
				s.Equal(tt.want.Name, ami.Name)
			}
		})
	}
}

func (s *AMITestSuite) TestLaunchInstance() {
	tests := []struct {
		name      string
		amiID     string
		amiName   string
		setupMock func(*localmock.MockEC2Client)
		want      *types.Instance
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success",
			amiID:   "ami-123",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-123"),
							ImageId:    aws.String("ami-123"),
						},
					},
				}

				m.On("RunInstances", mock.Anything, &ec2.RunInstancesInput{
					ImageId:      aws.String("ami-123"),
					InstanceType: types.InstanceTypeT2Micro,
					MinCount:     aws.Int32(1),
					MaxCount:     aws.Int32(1),
				}).Return(m.RunInstancesOutput, nil).Once()

				// Mock CreateTags
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}
				m.On("CreateTags", mock.Anything, &ec2.CreateTagsInput{
					Resources: []string{"i-123"},
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-ami"),
						},
					},
				}).Return(m.CreateTagsOutput, nil).Once()
			},
			want: &types.Instance{
				InstanceId: aws.String("i-123"),
				ImageId:    aws.String("ami-123"),
			},
			wantErr: false,
		},
		{
			name:    "run_instances_error",
			amiID:   "ami-123",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				m.On("RunInstances", mock.Anything, &ec2.RunInstancesInput{
					ImageId:      aws.String("ami-123"),
					InstanceType: types.InstanceTypeT2Micro,
					MinCount:     aws.Int32(1),
					MaxCount:     aws.Int32(1),
				}).Return(&ec2.RunInstancesOutput{
					Instances: []types.Instance{},
				}, fmt.Errorf("failed to launch instance")).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to launch instance: failed to launch instance",
		},
		{
			name:    "create_tags_error",
			amiID:   "ami-123",
			amiName: "test-ami",
			setupMock: func(m *localmock.MockEC2Client) {
				// Mock RunInstances to succeed
				m.On("RunInstances", mock.Anything, &ec2.RunInstancesInput{
					ImageId:      aws.String("ami-123"),
					InstanceType: types.InstanceTypeT2Micro,
					MinCount:     aws.Int32(1),
					MaxCount:     aws.Int32(1),
				}).Return(&ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-123"),
							ImageId:    aws.String("ami-123"),
						},
					},
				}, nil).Once()

				// Mock CreateTags to fail
				m.On("CreateTags", mock.Anything, &ec2.CreateTagsInput{
					Resources: []string{"i-123"},
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-ami"),
						},
					},
				}).Return(&ec2.CreateTagsOutput{}, fmt.Errorf("failed to create tags")).Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to create tags: failed to create tags",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock(s.mockClient)
			instance, err := s.service.LaunchInstance(context.Background(), tt.amiID, tt.amiName)
			if tt.wantErr {
				s.Error(err)
				if tt.errMsg != "" && err != nil {
					s.T().Logf("Expected error: %q", tt.errMsg)
					s.T().Logf("Actual error: %q", err.Error())
					switch tt.name {
					case "run_instances_error":
						s.True(errors.Is(err, ErrRunInstances))
					case "create_tags_error":
						s.True(errors.Is(err, ErrCreateTags))
					default:
						s.Contains(err.Error(), tt.errMsg)
					}
				}
			} else {
				s.NoError(err)
				s.Equal(tt.want.InstanceId, instance.InstanceId)
				s.Equal(tt.want.ImageId, instance.ImageId)
			}
		})
	}
}
