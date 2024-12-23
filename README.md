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
  create --os RHEL9 --size xlarge

# Migrate an instance to a new AMI
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  migrate --new-ami ami-xxxxx
```

## Core Features

- Complete EC2 instance lifecycle management
- Automatic AMI version tracking and migration
- Safe instance migration with volume snapshots
- Selective migration based on instance state and tags
- Comprehensive status tracking and error handling
- Automatic user detection from AWS credentials

## Common Tasks

### 1. List Your Instances
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

### 2. Check Migration Status
```bash
# Uses your AWS credentials username
ecman check

# Or specify a different username
ecman check --user johndoe
```

Shows for each instance:
- Current AMI details
- Latest available AMI
- Migration recommendation

### 3. Create New Instance
```bash
# Create default Ubuntu instance
ecman create

# Create custom RHEL instance
ecman create \
  --os RHEL9 \
  --size xlarge \
  --name my-instance
```

Options:
- `--os`: Ubuntu (default) or RHEL9
- `--size`: small, medium, large, xlarge
- `--name`: Custom instance name
- `--user`: Optional, defaults to AWS credentials username

### 4. Migrate Instances
```bash
# Migrate by tag
ecman migrate --new-ami ami-xxxxx

# Migrate specific instance
ecman migrate \
  --new-ami ami-xxxxx \
  --instance-id i-xxxxx
```

The migration process:
1. Takes volume snapshots for backup
2. Stops the instance if running
3. Creates new instance with target AMI
4. Copies all tags
5. Terminates old instance
6. Starts new instance if original was running

### 5. Login to AWS
```bash
# List available roles
ecman login --list-roles

# Login with a discovered role
ecman login --role-arn ROLE_ARN

# Login with MFA
ecman login \
  --role-arn ROLE_ARN \
  --mfa-serial MFA_DEVICE_ARN \
  --mfa-token 123456

# Login with custom profile and session duration
ecman login \
  --role-arn ROLE_ARN \
  --profile my-profile \
  --duration 7200
```

Options:
- `--list-roles`: List available roles in your AWS account
- `--role-arn`: ARN of the role to assume. If not provided, available roles will be displayed
- `--profile`: AWS profile to store credentials (default: "default")
- `--mfa-serial`: ARN of the MFA device (if MFA is required)
- `--mfa-token`: MFA token code (if MFA is required)
- `--duration`: Session duration in seconds (default: 3600)
- `--session-name`: Name for the role session (default: "ec-manager-session")

Note: If you don't know your role ARN, use `--list-roles` to discover available roles. If you don't have permission to list roles, the tool will display your AWS Account ID which you can use to construct the role ARN manually.

### 6. Delete Instances
```bash
# Delete using AWS credentials username
ecman delete --instance i-1234567890abcdef0

# Or specify a different username
ecman delete \
  --user johndoe \
  --instance i-1234567890abcdef0
```

Safely deletes with:
1. Ownership verification
2. Instance details display
3. Confirmation prompt
4. Status updates

## Instance Tags

Control instance behavior with these tags:

1. Main Migration Tag (Required):
```
Key: ami-migrate
Value: enabled
```

2. Running Instance Control (Optional):
```
Key: ami-migrate-if-running
Value: enabled
```

3. Owner Tag (Required for check/list):
```
Key: Owner
Value: <your-aws-username>
```

Tag Requirements:
- Running instances need BOTH `ami-migrate=enabled` AND `ami-migrate-if-running=enabled`
- Stopped instances only need `ami-migrate=enabled`
- Owner tag is automatically set to your AWS username when creating instances

## Migration Status Tracking

Status is tracked via tags:

1. Status Tag:
```
Key: ami-migrate-status
Value: skipped | in-progress | failed | warning | completed
```

2. Message Tag:
```
Key: ami-migrate-message
Value: [detailed status message]
```

## Developer Information

### Prerequisites
- Docker
- Make
- AWS credentials configured (for automatic user detection)

### Build and Test
```bash
# Build project
make build

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build Docker image
make docker-build
```

### Project Structure
```
ec-manager/
├── cmd/               # CLI commands
│   ├── backup.go     
│   ├── check.go      
│   ├── create.go     
│   ├── delete.go     
│   ├── list.go       
│   ├── migrate.go    
│   └── root.go       
├── pkg/
│   ├── ami/          # Core AMI functionality
│   │   ├── ami.go    
│   │   ├── ami_test.go
│   │   └── mock_ec2.go
│   ├── client/       # AWS client handling
│   │   ├── client.go
│   │   └── client_test.go
│   ├── config/       # Configuration
│   │   └── aws.go    
│   ├── logger/       # Logging utilities
│   │   └── logger.go
│   ├── testutil/     # Testing utilities
│   │   └── test_helpers.go
│   └── types/        # Common types and interfaces
│       └── types.go
```

### Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Format your code (`make fmt`)
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add some amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

Please make sure your PR:
- Includes tests for new functionality
- Updates documentation as needed
- Follows the existing code style
- Includes a clear description of changes

## Usage Notes

All commands support automatic user detection from your AWS credentials. The `--user` flag is optional and only needed if you want to operate on instances owned by a different user.

The tool will attempt to get your username in the following order:
1. From the `--user` flag if provided
2. From your AWS credentials file
3. From the IAM user info
4. From the STS caller identity

This means you can run most commands without explicitly specifying your username:

```bash
# List your instances
ecman list

# Check migration status
ecman check

# Create a new instance
ecman create --os RHEL9

# Delete an instance
ecman delete --instance i-xxxxx
```

## CI/CD Integration

For CI/CD pipelines, you can use environment variables for AWS credentials:
```bash
docker run --rm \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  -e AWS_DEFAULT_REGION \
  ec-manager:latest list
```

## AWS Configuration

When running the containerized version, mount your AWS credentials:

```bash
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ec-manager:latest \
  migrate \
  --new-ami ami-xxxxx
```

## License

MIT License