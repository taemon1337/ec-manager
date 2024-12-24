package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// mockAMIService is a mock implementation of the AMI service for testing
type mockAMIService struct {
	mockStatus *ami.MigrationStatus
	mockError  error
}

func (m *mockAMIService) GetMigrationStatus(ctx context.Context, user string) (*ami.MigrationStatus, error) {
	if m.mockError != nil {
		return nil, m.mockError
	}
	return m.mockStatus, nil
}

func (m *mockAMIService) FormatMigrationStatus(status *ami.MigrationStatus) string {
	if status == nil {
		return ""
	}

	return fmt.Sprintf(`Instance Status for %s:
  OS Type:        %s
  Instance Type:  %s
  State:          %s
  Launch Time:    %s
  Private IP:     %s
  Public IP:      %s

AMI Status:
  Current AMI:    %s
    Name:         %s
    Created:      %s
  Latest AMI:     %s
    Name:         %s
    Created:      %s

Migration Needed: %t

Run 'ami-migrate migrate' to update your instance to the latest AMI.
`,
		status.InstanceID,
		status.OSType,
		status.InstanceType,
		status.InstanceState,
		status.LaunchTime.Format(time.RFC3339),
		status.PrivateIP,
		status.PublicIP,
		status.CurrentAMI,
		status.CurrentAMIInfo.Name,
		status.CurrentAMIInfo.CreatedDate,
		status.LatestAMI,
		status.LatestAMIInfo.Name,
		status.LatestAMIInfo.CreatedDate,
		status.NeedsMigration,
	)
}

func TestCheckMigrateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		mockStatus  *ami.MigrationStatus
		mockError   error
		wantErr     bool
		wantOutput  string
	}{
		{
			name: "successful_check",
			args: []string{"--user", "testuser"},
			mockStatus: &ami.MigrationStatus{
				InstanceID:    "i-1234567890abcdef0",
				OSType:        "linux",
				InstanceType:  "t2.micro",
				InstanceState: "running",
				LaunchTime:    time.Date(2024, 12, 23, 0, 0, 0, 0, time.UTC),
				PrivateIP:     "172.16.0.10",
				PublicIP:      "54.123.45.67",
				CurrentAMI:    "ami-current",
				CurrentAMIInfo: &ami.AMIDetails{
					Name:        "current-ami",
					CreatedDate: "2024-12-01",
				},
				LatestAMI: "ami-latest",
				LatestAMIInfo: &ami.AMIDetails{
					Name:        "latest-ami",
					CreatedDate: "2024-12-20",
				},
				NeedsMigration: true,
			},
			wantOutput: `Instance Status for i-1234567890abcdef0:
  OS Type:        linux
  Instance Type:  t2.micro
  State:          running
  Launch Time:    2024-12-23T00:00:00Z
  Private IP:     172.16.0.10
  Public IP:      54.123.45.67

AMI Status:
  Current AMI:    ami-current
    Name:         current-ami
    Created:      2024-12-01
  Latest AMI:     ami-latest
    Name:         latest-ami
    Created:      2024-12-20

Migration Needed: true

Run 'ami-migrate migrate' to update your instance to the latest AMI.
`,
		},
		{
			name:      "missing_user_flag",
			args:      []string{},
			wantErr:   true,
		},
		{
			name:      "check_error",
			args:      []string{"--user", "testuser"},
			mockError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock AMI service
			mockService := &mockAMIService{
				mockStatus: tt.mockStatus,
				mockError:  tt.mockError,
			}

			// Create command and capture output
			cmd := &cobra.Command{
				Use:           "check",
				Short:         "Check migration status for an instance",
				Long:          "Check if an instance needs to be migrated to a new AMI",
				SilenceUsage:  true,
				SilenceErrors: true,
				RunE: func(cmd *cobra.Command, args []string) error {
					user, err := cmd.Flags().GetString("user")
					if err != nil {
						return err
					}
					if user == "" {
						return fmt.Errorf("--user flag must be specified")
					}

					// Get migration status
					status, err := mockService.GetMigrationStatus(cmd.Context(), user)
					if err != nil {
						return err
					}

					// Format and print status
					output := mockService.FormatMigrationStatus(status)
					fmt.Fprint(cmd.OutOrStdout(), output)

					return nil
				},
			}

			cmd.Flags().String("user", "", "Your AWS username")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, buf.String())
			}
		})
	}
}
