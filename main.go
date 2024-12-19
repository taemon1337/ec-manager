package main

import (
	"context"
	"flag"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

func main() {
	// Parse CLI arguments
	newAMI := flag.String("new-ami", "", "The ID of the new AMI to upgrade instances to")
	flag.Parse()

	if *newAMI == "" {
		log.Fatal("--new-ami argument is required")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	amiService := ami.NewService(ec2Client)
	ctx := context.Background()

	// Fetch the current "latest" AMI ID
	oldAMI, err := amiService.GetAMIWithTag(ctx, "latest")
	if err != nil {
		log.Fatalf("failed to fetch AMI with 'latest' tag: %v", err)
	}

	// Update AMI tags
	if err := amiService.UpdateAMITags(ctx, oldAMI, *newAMI); err != nil {
		log.Fatalf("failed to update AMI tags: %v", err)
	}

	// Migrate instances from old AMI to new AMI
	if err := amiService.MigrateInstances(ctx, oldAMI, *newAMI); err != nil {
		log.Fatalf("failed to migrate instances: %v", err)
	}

	log.Println("Migration completed successfully")
}
