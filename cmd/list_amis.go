package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

// listAmisCmd represents the list AMIs command
var listAmisCmd = &cobra.Command{
	Use:   "amis",
	Short: "List AMIs",
	Long:  `List all AMIs created by this project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get AMIs with project tags
		images, err := awsClient.ListImages([]types.Filter{
			{
				Name:   aws.String("tag:Project"),
				Values: []string{"ec-manager"},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to list AMIs: %w", err)
		}

		if len(images) == 0 {
			fmt.Println("No AMIs found")
			return nil
		}

		for _, image := range images {
			fmt.Printf("AMI ID: %s\n", *image.ImageId)
			if image.Name != nil {
				fmt.Printf("  Name: %s\n", *image.Name)
			}
			if image.Description != nil {
				fmt.Printf("  Description: %s\n", *image.Description)
			}
			if image.CreationDate != nil {
				fmt.Printf("  Created: %s\n", *image.CreationDate)
			}
			if image.State != "" {
				fmt.Printf("  State: %s\n", image.State)
			}
			// Print tags
			fmt.Println("  Tags:")
			for _, tag := range image.Tags {
				if tag.Key != nil && tag.Value != nil {
					fmt.Printf("    %s: %s\n", *tag.Key, *tag.Value)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listAmisCmd)
}
