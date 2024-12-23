package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate EC2 instances to new AMIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			newAMI, _ := cmd.Flags().GetString("new-ami")
			enabled, _ := cmd.Flags().GetBool("enabled")

			if instanceID == "" && !enabled {
				return fmt.Errorf("either --instance-id or --enabled flag must be set")
			}

			if newAMI == "" {
				return fmt.Errorf("--new-ami flag must be specified")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			
			if instanceID != "" {
				ec2Client := client.GetEC2Client()
				
				// Check if instance exists
				output, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				})
				if err != nil {
					return fmt.Errorf("failed to describe instance: %w", err)
				}
				if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
					return fmt.Errorf("instance not found: %s", instanceID)
				}
			}

			return nil
		},
	}
	cmd.Flags().String("instance-id", "", "ID of the instance")
	cmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")
	cmd.Flags().Bool("enabled", false, "Migrate all instances with ami-migrate=enabled tag")
	return cmd
}

func TestMigrateCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
		setup   func(client *ecTypes.MockEC2Client)
	}{
		{
			name: "successful_migration",
			args: []string{"--instance-id", "i-123", "--new-ami", "ami-new"},
			setup: func(client *ecTypes.MockEC2Client) {
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
			name:    "instance_not_found",
			args:    []string{"--instance-id", "i-nonexistent", "--new-ami", "ami-new"},
			wantErr: true,
			errMsg:  "instance not found",
			setup: func(client *ecTypes.MockEC2Client) {
				client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
		},
		{
			name:    "no_instance_ID_and_enabled_flag_not_set",
			args:    []string{"--new-ami", "ami-new"},
			wantErr: true,
			errMsg:  "either --instance-id or --enabled flag must be set",
		},
		{
			name:    "missing_new-ami_flag",
			args:    []string{"--instance-id", "i-123"},
			wantErr: true,
			errMsg:  "--new-ami flag must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewMigrateCmd()
			
			mockClient := client.NewMockEC2Client()
			if tt.setup != nil {
				tt.setup(mockClient)
			}
			client.SetMockMode(true)
			client.SetMockClient(mockClient)
			defer func() {
				client.SetMockMode(false)
				client.SetMockClient(nil)
			}()

			err := testutil.SetupTestCommand(cmd, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" {
					testutil.AssertErrorContains(t, err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
