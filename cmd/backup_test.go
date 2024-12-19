package cmd

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func init() {
	// Initialize mock EC2 client
	ec2Client = &mockEC2Client{
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
}

func TestBackupCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{"backup"},
			wantErr: false,
		},
		{
			name:    "with instance ID",
			args:    []string{"backup", "--instance-id", "i-123"},
			wantErr: false,
		},
		{
			name:    "with enabled value",
			args:    []string{"backup", "--instance-id", "i-123"},
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
				t.Errorf("backup command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
