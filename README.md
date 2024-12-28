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
  list

# Create a new instance
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  create --os linux --size t2.micro

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
- `create`: Create a new EC2 instance
- `delete`: Delete an EC2 instance
- `list`: List your EC2 instances
- `restore`: Restore an instance from a snapshot
- `ssh`: SSH into an EC2 instance

### AMI Management
- `check`: Check various aspects of your AWS resources
- `check migrate`: Check instances that need AMI migration
- `migrate`: Migrate an EC2 instance to a new AMI
- `list amis`: List available AMIs in your account

### Resource Listing
- `list instances`: List all EC2 instances
- `list keys`: List available SSH key pairs
- `list subnets`: List available VPC subnets

### Authentication and Credentials
- `check credentials`: Verify AWS credentials and permissions

### Common Tasks

#### 1. List Your Instances
```bash
# Uses your AWS credentials username
ecman list

# Or specify a different username
ecman list --user johndoe
```

Output shows:
- Instance name and ID
- OS type and instance size
- Current state (running/stopped)
- IP addresses
- Current and latest AMI versions
- Migration status

#### 2. Check Migration Status
```bash
# Check migration status for your instances
ecman check migrate --user johndoe
```

Shows for each instance:
- Current AMI details
- Latest available AMI
- Migration recommendation

#### 3. Create New Instance
```bash
# Create a new Linux instance
ecman create \
  --os linux \
  --size t2.micro \
  --name my-instance \
  --user johndoe
```

Options:
- `--os`: Operating system type (linux or windows)
- `--size`: Instance size (e.g., t2.micro, t2.small)
- `--name`: Custom instance name (optional)
- `--user`: AWS username (optional, defaults to AWS credentials)

#### 4. Migrate Instances
```bash
# Migrate a specific instance
ecman migrate \
  --instance-id i-1234567890abcdef0 \
  --new-ami ami-xxxxx

# Migrate all enabled instances
ecman migrate \
  --enabled \
  --new-ami ami-xxxxx
```

Options:
- `--instance-id`: ID of the instance to migrate
- `--enabled`: Only migrate instances with ami-migrate=enabled tag
- `--new-ami`: ID of the new AMI to migrate to

#### 5. Backup and Restore
```bash
# Backup an instance
ecman backup --instance-id i-1234567890abcdef0

# Restore from a snapshot
ecman restore \
  --instance-id i-1234567890abcdef0 \
  --snapshot-id snap-xxxxx
```

#### 6. SSH into an Instance
```bash
# SSH into a specific instance
ecman ssh \
  --instance i-1234567890abcdef0 \
  --key ~/.ssh/my-key.pem \
  --user ec2-user
```

Options:
- `--instance` or `-i`: ID of the instance to connect to
- `--key` or `-k`: Path to the SSH private key file
- `--user` or `-u`: SSH user (defaults to ec2-user)

## Global Flags

Available for all commands:
- `--mock`: Enable mock mode for testing
- `--log-level`: Set log level (debug, info, warn, error)
- `--timeout`: Timeout for AWS operations (default: 5m)
- `--user`: Your AWS username (defaults to current AWS user)

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