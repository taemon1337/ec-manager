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
	Long: `Create a new EC2 instance with specified parameters.
Requires:
- OS type (linux or windows)
- Instance size (e.g. t2.micro)
- Instance name
- Your user ID (for ownership)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		osType, err := cmd.Flags().GetString("os")
		if err != nil {
			return fmt.Errorf("get os flag: %w", err)
		}

		size, err := cmd.Flags().GetString("size")
		if err != nil {
			return fmt.Errorf("get size flag: %w", err)
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("get name flag: %w", err)
		}

		if name == "" {
			name = generateInstanceName()
		}

		if userID == "" || osType == "" || size == "" {
			return fmt.Errorf("--user, --os, and --size flags are required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		amiService := ami.NewService(ec2Client)

		config := ami.InstanceConfig{
			Name:   name,
			OSType: osType,
			Size:   size,
			UserID: userID,
		}

		instance, err := amiService.CreateInstance(cmd.Context(), config)
		if err != nil {
			return fmt.Errorf("create instance: %w", err)
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
	createCmd.MarkFlagRequired("user")
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
