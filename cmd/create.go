package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var (
	imageID      string
	instanceType string
	keyName      string
	subnetID     string
	userData     string
	useLatestAmi bool
)

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new EC2 instance",
	Long:  "Create a new EC2 instance with specified configuration",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !useLatestAmi && imageID == "" {
			return fmt.Errorf("either --image or --latest flag must be specified")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		// If --latest flag is set, find the latest AMI
		if useLatestAmi {
			filters := []types.Filter{
				{
					Name:   aws.String("tag:ami-migrate"),
					Values: []string{"latest"},
				},
			}

			output, err := awsClient.GetEC2Client().DescribeImages(ctx, &ec2.DescribeImagesInput{
				Filters: filters,
			})
			if err != nil {
				return fmt.Errorf("failed to list AMIs: %w", err)
			}

			if len(output.Images) == 0 {
				return fmt.Errorf("no AMIs found with tag ami-migrate=latest")
			}

			// Find the newest AMI
			var latestAmi types.Image
			for _, image := range output.Images {
				if latestAmi.CreationDate == nil || *image.CreationDate > *latestAmi.CreationDate {
					latestAmi = image
				}
			}
			imageID = *latestAmi.ImageId
			fmt.Printf("Using latest AMI: %s (created: %s)\n", imageID, *latestAmi.CreationDate)
		}

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

func init() {
	rootCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringVar(&imageID, "image", "", "AMI ID to use for the instance")
	CreateCmd.Flags().StringVar(&instanceType, "type", "t2.micro", "Instance type")
	CreateCmd.Flags().StringVar(&keyName, "key", "", "SSH key name")
	CreateCmd.Flags().StringVar(&subnetID, "subnet", "", "Subnet ID")
	CreateCmd.Flags().StringVar(&userData, "userdata", "", "User data script")
	CreateCmd.Flags().BoolVar(&useLatestAmi, "latest", false, "Use the latest AMI with tag ami-migrate=latest")

	CreateCmd.MarkFlagRequired("key")
	CreateCmd.MarkFlagRequired("subnet")
}
