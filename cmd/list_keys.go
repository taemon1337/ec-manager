package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// listKeysCmd represents the list keys command
var listKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List SSH key pairs",
	Long:  `List all SSH key pairs in your AWS account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		keys, err := amiService.ListKeyPairs(ctx)
		if err != nil {
			return fmt.Errorf("failed to list key pairs: %w", err)
		}

		if len(keys) == 0 {
			fmt.Println("No key pairs found")
			return nil
		}

		for _, key := range keys {
			fmt.Printf("Key Name: %s\n", *key.KeyName)
			if key.KeyFingerprint != nil {
				fmt.Printf("  Fingerprint: %s\n", *key.KeyFingerprint)
			}
			if len(key.Tags) > 0 {
				fmt.Println("  Tags:")
				for _, tag := range key.Tags {
					if tag.Key != nil && tag.Value != nil {
						fmt.Printf("    %s: %s\n", *tag.Key, *tag.Value)
					}
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listKeysCmd)
}
