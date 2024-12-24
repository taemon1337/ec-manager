package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// MigrateCmd represents the migrate command
var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate an instance",
	Long:  "Migrate an instance by creating an AMI and launching a new instance from it",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		return amiService.MigrateInstance(ctx, migrateInstanceID, newAMI)
	},
}

var (
	migrateInstanceID string
	newAMI           string
)

func init() {
	rootCmd.AddCommand(MigrateCmd)

	MigrateCmd.Flags().StringVarP(&migrateInstanceID, "instance", "i", "", "Instance ID to migrate")
	MigrateCmd.Flags().StringVar(&newAMI, "new-ami", "", "New AMI ID to migrate to")
	MigrateCmd.MarkFlagRequired("instance")
	MigrateCmd.MarkFlagRequired("new-ami")
}
