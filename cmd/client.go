package cmd

import "github.com/taemon1337/ami-migrate/pkg/ami"

// ec2Client is used to allow mocking in tests
var ec2Client ami.EC2ClientAPI
