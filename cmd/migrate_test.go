package cmd

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func init() {
	// Initialize mock EC2 client for tests
	mockClient := &mockEC2Client{
		images: []types.Image{
			{
				ImageId: aws.String("ami-latest"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Status"),
						Value: aws.String("latest"),
					},
				},
			},
		},
		instances: []types.Instance{
			{
				InstanceId: aws.String("i-123"),
				ImageId:    aws.String("ami-latest"),
				State: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
				Tags: []types.Tag{
					{
						Key:   aws.String("ami-migrate"),
						Value: aws.String("enabled"),
					},
				},
			},
		},
	}
	ec2Client = mockClient
}

func TestMigrateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no new AMI",
			args:        []string{"migrate"},
			wantErr:     true,
			errContains: "--new-ami is required",
		},
		{
			name:    "with new AMI",
			args:    []string{"migrate", "--new-ami", "ami-123"},
			wantErr: false,
		},
		{
			name:    "with instance ID",
			args:    []string{"migrate", "--new-ami", "ami-123", "--instance-id", "i-123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("migrate command error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
					t.Errorf("migrate command error = %v, want it to contain %v", err, tt.errContains)
				}
			}
		})
	}
}
