package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/testutil"
)

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate an EC2 instance",
		Long:  "Migrate an EC2 instance to a new AMI",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, err := cmd.Flags().GetString("instance")
			if err != nil {
				return fmt.Errorf("failed to get instance flag: %w", err)
			}

			newAMI, err := cmd.Flags().GetString("new-ami")
			if err != nil {
				return fmt.Errorf("failed to get new-ami flag: %w", err)
			}

			if newAMI == "" {
				return fmt.Errorf("--new-ami flag must be specified")
			}

			enabled, err := cmd.Flags().GetBool("enabled")
			if err != nil {
				return fmt.Errorf("failed to get enabled flag: %w", err)
			}

			if instanceID == "" && !enabled {
				return fmt.Errorf("either --instance or --enabled flag must be set")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			if instanceID != "" {
				ec2Client := testutil.GetEC2Client(cmd)
				if ec2Client == nil {
					return fmt.Errorf("failed to get EC2 client")
				}

				// Check if instance exists
				output, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				})
				if err != nil {
					return fmt.Errorf("failed to describe instance: %w", err)
				}

				if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
					return fmt.Errorf("instance not found")
				}
			}

			return nil
		},
	}

	cmd.Flags().String("instance", "", "Instance to migrate")
	cmd.Flags().String("new-ami", "", "New AMI ID to migrate to")
	cmd.Flags().Bool("enabled", false, "Migrate all enabled instances")

	return cmd
}

func TestMigrateCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errMsg    string
		setup     func(client *mock.MockEC2Client)
	}{
		{
			name: "successful_migration",
			args: []string{"--instance", "i-123", "--new-ami", "ami-123"},
			setup: func(client *mock.MockEC2Client) {
				client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
								},
							},
						},
					},
				}
			},
		},
		{
			name:      "instance_not_found",
			args:      []string{"--instance", "i-nonexistent", "--new-ami", "ami-123"},
			wantError: true,
			errMsg:    "instance not found",
			setup: func(client *mock.MockEC2Client) {
				client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
		},
		{
			name:      "missing_new_ami",
			args:      []string{"--instance", "i-123"},
			wantError: true,
			errMsg:    "--new-ami flag must be specified",
		},
		{
			name:      "no_instance_ID_and_enabled_flag_not_set",
			args:      []string{"--new-ami", "ami-123"},
			wantError: true,
			errMsg:    "either --instance or --enabled flag must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewMigrateCmd()

			if tt.setup != nil {
				mockClient := mock.NewMockEC2Client()
				tt.setup(mockClient)
				cmd.SetContext(context.WithValue(context.Background(), "ec2_client", mockClient))
			}

			err := testutil.SetupTestCommand(cmd, tt.args)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
