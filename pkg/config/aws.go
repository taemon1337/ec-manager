package config

import (
	"os"
	"path/filepath"

	"bufio"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// GetAWSUsername returns the current AWS username from either:
// 1. AWS credentials file
// 2. IAM user info
// 3. STS caller identity
func GetAWSUsername(ctx context.Context) (string, error) {
	// First try to get from credentials file
	if username := getUserFromCredentials(); username != "" {
		return username, nil
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	// Try IAM user info first
	iamClient := iam.NewFromConfig(cfg)
	user, err := iamClient.GetUser(ctx, &iam.GetUserInput{})
	if err == nil && user.User != nil && user.User.UserName != nil {
		return *user.User.UserName, nil
	}

	// Fall back to STS caller identity
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	// Extract username from ARN
	// ARN format: arn:aws:iam::123456789012:user/username
	parts := strings.Split(*identity.Arn, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return "", nil
}

// getUserFromCredentials attempts to read the username from AWS credentials file
func getUserFromCredentials() string {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Read AWS credentials file
	credFile := filepath.Join(home, ".aws", "credentials")
	file, err := os.Open(credFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "username") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
