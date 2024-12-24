package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new EC2 instance",
	Long: `Create a new EC2 instance with specified parameters.
	
	Example:
	  ecman create --user john.doe --os linux --size t2.micro`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user ID from flag
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		// Get OS type from flag
		osType, err := cmd.Flags().GetString("os")
		if err != nil {
			return fmt.Errorf("get os flag: %w", err)
		}

		// Get instance size from flag
		size, err := cmd.Flags().GetString("size")
		if err != nil {
			return fmt.Errorf("get size flag: %w", err)
		}

		// Get instance name from flag
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("get name flag: %w", err)
		}

		// Get EC2 client
		c := client.NewClient()
		ec2Client, err := c.GetEC2Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("get EC2 client: %w", err)
		}

		// Create AMI service
		amiService := ami.NewService(ec2Client)

		// Create instance config
		instanceConfig := ami.InstanceConfig{
			UserID: userID,
			OSType: osType,
			Size:   size,
			Name:   name,
		}

		// Create instance
		instance, err := amiService.CreateInstance(cmd.Context(), instanceConfig)
		if err != nil {
			return fmt.Errorf("create instance: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created instance %s\n", instance.InstanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("user", "", "Your user ID")
	createCmd.Flags().String("os", "", "OS type (linux or windows)")
	createCmd.Flags().String("size", "", "Instance size (e.g. t2.micro)")
	createCmd.Flags().String("name", "", "Instance name (optional, random if not provided)")
	createCmd.MarkFlagRequired("os")
	createCmd.MarkFlagRequired("size")

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

func generateInstanceName() string {
	adjectives := []string{"happy", "clever", "swift", "bright", "agile"}
	nouns := []string{"penguin", "dolphin", "falcon", "tiger", "wolf"}
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s-%s", adj, noun, timestamp)
}
