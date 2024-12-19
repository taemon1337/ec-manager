package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

func main() {
	// Parse command line arguments
	newAMI := flag.String("new-ami", "", "ID of the new AMI to migrate to")
	latestTag := flag.String("latest-tag", "latest", "Tag value for the current latest AMI")
	enabledValue := flag.String("enabled-value", "enabled", "Value to match for the ami-migrate tag")
	flag.Parse()

	if *newAMI == "" {
		log.Fatal("--new-ami is required")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create AMI service
	amiService := ami.NewService(ec2Client)

	// Get current latest AMI
	oldAMI, err := amiService.GetAMIWithTag(context.Background(), "Status", *latestTag)
	if err != nil {
		log.Fatalf("Failed to get current AMI: %v", err)
	}

	// Update AMI tags
	if err := amiService.TagAMI(context.Background(), oldAMI, "Status", "previous"); err != nil {
		log.Printf("Warning: Failed to update old AMI tags: %v", err)
	}
	if err := amiService.TagAMI(context.Background(), *newAMI, "Status", *latestTag); err != nil {
		log.Printf("Warning: Failed to update new AMI tags: %v", err)
	}

	// Migrate instances
	fmt.Printf("Starting migration from AMI %s to %s\n", oldAMI, *newAMI)
	fmt.Printf("Will migrate instances with tag 'ami-migrate=%s'\n", *enabledValue)
	fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

	if err := amiService.MigrateInstances(context.Background(), oldAMI, *newAMI, *enabledValue); err != nil {
		log.Fatalf("Failed to migrate instances: %v", err)
	}

	fmt.Println("Migration completed successfully")
	os.Exit(0)
}
