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
	enabledValue := flag.String("enabled-value", "enabled", "Value to match for the ami-migrate tag")
	flag.Parse()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create AMI service
	amiService := ami.NewService(ec2Client)

	// Migrate instances
	fmt.Printf("Starting migration for instances with tag 'ami-migrate=%s'\n", *enabledValue)
	fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

	if err := amiService.MigrateInstances(context.Background(), *enabledValue); err != nil {
		log.Fatalf("Failed to migrate instances: %v", err)
	}

	fmt.Println("Migration completed successfully")
	os.Exit(0)
}
