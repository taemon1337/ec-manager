package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new EC2 instance",
	Long:  "Create a new EC2 instance with specified configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		cfg := ami.InstanceConfig{
			ImageID:      imageID,
			InstanceType: instanceType,
			KeyName:      keyName,
			SubnetID:     subnetID,
			UserData:     userData,
		}

		instanceID, err := amiService.CreateInstance(ctx, cfg)
		if err != nil {
			return err
		}

		fmt.Printf("Created instance %s with:\n", instanceID)
		fmt.Printf("  Image ID: %s\n", imageID)
		fmt.Printf("  Instance Type: %s\n", instanceType)
		fmt.Printf("  Key Name: %s\n", keyName)
		fmt.Printf("  Subnet ID: %s\n", subnetID)
		if userData != "" {
			fmt.Println("  User Data: [provided]")
		}

		return nil
	},
}

var (
	imageID      string
	instanceType string
	keyName      string
	subnetID     string
	userData     string
)

func init() {
	rootCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringVarP(&imageID, "image", "i", "", "AMI ID to use")
	CreateCmd.Flags().StringVarP(&instanceType, "type", "t", "t2.micro", "Instance type")
	CreateCmd.Flags().StringVarP(&keyName, "key", "k", "", "Key pair name")
	CreateCmd.Flags().StringVarP(&subnetID, "subnet", "s", "", "Subnet ID")
	CreateCmd.Flags().StringVarP(&userData, "userdata", "u", "", "User data script")

	CreateCmd.MarkFlagRequired("image")
	CreateCmd.MarkFlagRequired("key")
	CreateCmd.MarkFlagRequired("subnet")
}
