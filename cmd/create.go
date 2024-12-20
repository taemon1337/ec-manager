package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new EC2 instance",
	Long: `Create a new EC2 instance with the specified configuration.
The instance will be tagged with your user ID and ami-migrate tags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user ID
		userID, err := getUserID(cmd)
		if err != nil {
			return err
		}

		// Get other flags
		osType, err := cmd.Flags().GetString("os")
		if err != nil {
			return err
		}

		size, err := cmd.Flags().GetString("size")
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		// Validate required flags
		if osType == "" || size == "" {
			return fmt.Errorf("--os and --size flags are required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		amiService := ami.NewService(ec2Client)

		// Create instance config
		config := ami.InstanceConfig{
			UserID: userID,
			OSType: osType,
			Size:   size,
			Name:   name,
		}

		// Create instance
		instance, err := amiService.CreateInstance(cmd.Context(), config)
		if err != nil {
			return fmt.Errorf("failed to create instance: %v", err)
		}

		fmt.Println("Instance created successfully:")
		fmt.Print(instance.FormatInstanceSummary())
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
}

func generateInstanceName() string {
	rand.Seed(time.Now().UnixNano())
	adjectives := []string{"happy", "clever", "swift", "bright", "agile"}
	nouns := []string{"penguin", "dolphin", "falcon", "tiger", "wolf"}
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s-%s", adj, noun, timestamp)
}
