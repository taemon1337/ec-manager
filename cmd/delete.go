package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EC2 instance",
	Long:  "Delete an EC2 instance by its ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())
		
		state, err := amiService.DeleteInstance(ctx, deleteInstanceID)
		if err != nil {
			return err
		}

		fmt.Printf("Instance %s is being terminated (current state: %s)\n", deleteInstanceID, state)
		return nil
	},
}

var deleteInstanceID string

func init() {
	rootCmd.AddCommand(DeleteCmd)

	DeleteCmd.Flags().StringVarP(&deleteInstanceID, "instance", "i", "", "Instance ID to delete")
	if err := DeleteCmd.MarkFlagRequired("instance"); err != nil {
		panic(err)
	}
}
