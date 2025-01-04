# EC2 Manager (ecman)

[![Go Report Card](https://goreportcard.com/badge/github.com/taemon1337/ec-manager)](https://goreportcard.com/report/github.com/taemon1337/ec-manager)
[![GoDoc](https://godoc.org/github.com/taemon1337/ec-manager?status.svg)](https://godoc.org/github.com/taemon1337/ec-manager)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive CLI tool for managing AWS EC2 instances. This tool helps automate common EC2 management tasks including instance creation, AMI migrations, backups, and lifecycle management.

## Quick Start

1. Install the tool:
```bash
docker pull ec-manager:latest
```

2. Run your first command:
```bash
# List your instances
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  list instances

# Create a new instance
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  create --type t2.micro --key my-key --subnet subnet-xxx

# Migrate an instance to a new AMI
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  migrate --instance-id i-1234567890abcdef0 --new-ami ami-xxxxx
```

## Core Features

- Complete EC2 instance lifecycle management
- Automatic AMI version tracking and migration
- Safe instance migration with volume snapshots
- Selective migration based on instance state and tags
- Comprehensive status tracking and error handling
- Mock mode for testing and development
- Automatic user detection from AWS credentials

## Available Commands

### Instance Management
- `backup`: Backup an EC2 instance
  - `-i, --instance-id`: Instance ID to backup (required)

- `create`: Create a new EC2 instance
  - `--key`: SSH key name (required)
  - `--subnet`: Subnet ID (required)
  - `--ami`: AMI ID
  - `--type`: Instance type (default: t2.micro)
  - `--name`: Instance name

- `delete`: Delete an EC2 instance
  - `-i, --instance`: Instance ID to delete (required)

- `restore`: Restore an instance from a snapshot or version
  - `-i, --instance-id`: Instance ID to restore (required)
  - `-s, --snapshot`: Snapshot ID to restore from (optional if using --version)
  - `-v, --version`: Version to restore to (optional if using --snapshot)

### Instance State Management
- `start`: Start an EC2 instance
  - `-i, --instance`: Instance ID to start (required)

- `stop`: Stop an EC2 instance
  - `-i, --instance`: Instance ID to stop (required)

- `restart`: Restart an EC2 instance
  - `-i, --instance`: Instance ID to restart (required)

### AMI Management
- `check migrate`: Check instances that need AMI migration
  - `-i, --check-instance-id`: Instance ID to check for migration
  - `-a, --check-target-ami`: New AMI ID to migrate to

- `migrate`: Migrate an EC2 instance to a new AMI
  - `-i, --instance-id`: Instance to migrate
  - `-a, --new-ami`: New AMI ID to migrate to
  - `-e, --enabled`: Migrate all enabled instances
  - `-v, --version`: Version to migrate to

### Resource Listing
- `list instances`: List all EC2 instances
- `list amis`: List available AMIs in your account
- `list keys`: List available SSH key pairs
- `list subnets`: List available VPC subnets

### Authentication and Access
- `check credentials`: Verify AWS credentials and permissions

- `ssh`: SSH into an EC2 instance
  - `-i, --instance`: Instance ID to SSH into (required)
  - `-k, --key`: Path to SSH private key file (required)
  - `-u, --user`: SSH user (default: ec2-user)

## Global Flags

Available for all commands:
- `--mock`: Enable mock mode for testing
- `--log-level`: Set log level (debug, info, warn, error)
- `--region`: AWS region to use
- `--profile`: AWS profile to use

## Development

### Mock Mode

The tool supports a mock mode for testing and development. Enable it with the `--mock` flag:

```bash
ecman list --mock
```

This will use mock clients instead of making real AWS API calls.

### Building from Source

```bash
# Build the binary
make build

# Run tests
make test

# Run with coverage
make cover
```

## License

MIT License - see LICENSE file for details.