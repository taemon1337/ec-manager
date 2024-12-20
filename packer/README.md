# AMI Builder with Packer

This directory contains Packer configurations for building custom AMIs for RHEL 9 and Ubuntu.

## Features

- Containerized Packer builds
- Base configuration for both RHEL 9 and Ubuntu
- Common system optimizations
- Security hardening
- CloudWatch agent installation
- Network performance tuning

## Prerequisites

- Docker
- AWS credentials configured (`~/.aws/credentials` or environment variables)
- Make

## Quick Start

1. Build the Packer Docker image:
```bash
make docker-build
```

2. Initialize Packer plugins:
```bash
make init
```

3. Validate templates:
```bash
make validate
```

4. Build AMIs:
```bash
# Build all AMIs
make all

# Build specific AMI
make rhel9
# or
make ubuntu
```

## Configuration

### Variables

Both RHEL 9 and Ubuntu configurations support the following variables:

- `aws_region` (default: us-east-1)
- `source_ami` (defaults to latest official AMI in us-east-1)
- `instance_type` (default: t3.micro)
- `ami_name_prefix` (default: rhel9-custom or ubuntu-custom)

Override variables using the -var flag:
```bash
docker run --rm -it \
  -v $(PWD):/workspace \
  -v ~/.aws:/root/.aws:ro \
  ami-migrate-packer:latest \
  build -var="aws_region=us-west-2" rhel9/rhel9.pkr.hcl
```

### RHEL Subscription

For RHEL builds, you can provide Red Hat subscription credentials:
```bash
docker run --rm -it \
  -v $(PWD):/workspace \
  -v ~/.aws:/root/.aws:ro \
  -e RHSM_USER=your-username \
  -e RHSM_PASS=your-password \
  ami-migrate-packer:latest \
  build rhel9/rhel9.pkr.hcl
```

## Base Configuration

All AMIs include:

- AWS CLI v2
- CloudWatch agent
- Common utilities (curl, wget, git, etc.)
- Network performance optimizations
- Security configurations
- Automatic updates
- UTC timezone

### RHEL 9 Specific

- EPEL repository
- Tuned profile for virtual machines
- SELinux configuration
- Chronyd for time synchronization

### Ubuntu Specific

- Unattended upgrades
- AppArmor configuration
- Chronyd for time synchronization

## Output

After a successful build:

1. AMI is created with tags:
   - Name: [prefix]-[timestamp]
   - OS: RHEL9/Ubuntu
   - Status: latest
   - BuildDate: YYYY-MM-DD
   - BuildTime: HH:MM:SS

2. A manifest.json file is created in the build directory with AMI details

## Integration with ami-migrate

The AMIs built with these configurations are compatible with the ami-migrate tool. They are automatically tagged with "Status=latest" which allows ami-migrate to track and manage AMI versions.
