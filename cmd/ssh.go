package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var (
	sshInstanceID string
	sshKeyPath    string
	sshUser       string
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into an EC2 instance",
	Long:  `SSH into an EC2 instance using its instance ID and the specified key pair.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		// Get instance details
		instance, err := amiService.GetInstance(ctx, sshInstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance details: %w", err)
		}

		if instance.PublicIpAddress == nil {
			return fmt.Errorf("instance %s does not have a public IP address", sshInstanceID)
		}

		// Prepare SSH command
		sshCmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			fmt.Sprintf("%s@%s", sshUser, *instance.PublicIpAddress),
		)

		// Connect SSH session to current terminal
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		// Execute SSH command
		return sshCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.Flags().StringVarP(&sshInstanceID, "instance", "i", "", "Instance ID to SSH into")
	sshCmd.Flags().StringVarP(&sshKeyPath, "key", "k", "", "Path to SSH private key file")
	sshCmd.Flags().StringVarP(&sshUser, "user", "u", "ec2-user", "SSH user (default: ec2-user)")

	if err := sshCmd.MarkFlagRequired("instance"); err != nil {
		panic(err)
	}
	if err := sshCmd.MarkFlagRequired("key"); err != nil {
		panic(err)
	}
}
