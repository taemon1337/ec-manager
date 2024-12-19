# AMI Migration Tool

A Go-based tool for managing AWS AMI migrations. This tool helps automate the process of updating EC2 instances to use a new AMI while maintaining proper versioning and backups.

## Features

- Automatic AMI version tracking using tags
- Parallel instance migration
- Volume snapshots for backup
- Graceful error handling and status tracking
- Clean separation of concerns with pkg structure
- Docker-based build and test environment
- Selective instance migration based on tags
- Migration of instances regardless of current AMI
- Smart instance state handling

## Prerequisites

- Docker
- Make

No need to install Go locally - all Go operations are performed within Docker containers!

## Quick Start

1. Build the project:
```bash
make build
```

2. Run tests:
```bash
make test
```

3. Build Docker image:
```bash
make docker-build
```

4. Run the tool:
```bash
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ami-migrate:latest \
  --new-ami ami-xxxxx \
  --latest-tag latest \
  --enabled-value enabled
```

## CLI Arguments

- `--new-ami` (required): The ID of the new AMI to upgrade instances to
- `--latest-tag` (optional, default: "latest"): The tag used to track the current AMI version
- `--enabled-value` (optional, default: "enabled"): Value to match for the "ami-migrate" tag

## How It Works

1. **AMI Version Management**:
   - The tool tracks AMI versions using tags
   - When a new AMI is specified, the old "latest" AMI is archived
   - The new AMI becomes the "latest" version

2. **Instance Selection**:
   - Instances are selected for migration if they have `ami-migrate=enabled`
   - Running instances are migrated immediately
   - Stopped instances are handled based on additional tags

3. **Instance State Handling**:
   - Running instances: Migrated directly
   - Stopped instances with `ami-migrate-if-running=enabled`: 
     1. Instance is started
     2. Migration is performed
     3. Instance is returned to its original state
   - Stopped instances without the if-running tag: Skipped
   - Failed migrations are tagged for tracking

4. **Migration Process**:
   - Migrations run in parallel for faster updates
   - Each instance's volumes are snapshotted before migration
   - Instance state is preserved throughout the process
   - Comprehensive error handling and status tracking

## Instance Tagging

Two tags control the migration behavior:

1. Main Migration Tag:
```
Key: ami-migrate
Value: enabled  # or your custom value specified with --enabled-value
```

2. Optional State Control Tag:
```
Key: ami-migrate-if-running
Value: enabled
```

Tag Combinations and Behavior:
- `ami-migrate=enabled` only:
  - Running instances: Will be migrated
  - Stopped instances: Will be skipped
- `ami-migrate=enabled` AND `ami-migrate-if-running=enabled`:
  - Running instances: Will be migrated
  - Stopped instances: Will be started, migrated, then stopped again

## Development

### Available Make Commands

- `make all` - Clean, build, and test
- `make build` - Build the binary
- `make clean` - Clean build artifacts
- `make test` - Run tests in Docker
- `make lint` - Run linter in Docker
- `make docker-build` - Build Docker image
- `make docker-test` - Run tests in Docker
- `make init` - Initialize go.mod (if needed)
- `make shell` - Open an interactive shell in the container

### Project Structure

```
ami-migrate/
├── Dockerfile          # Multi-stage Docker build
├── Makefile           # Build automation
├── main.go            # Entry point
├── pkg/
│   └── ami/           # AMI management package
│       ├── ami.go     # Core AMI operations
│       └── ami_test.go # Unit tests
```

### Adding New Features

1. Add new functionality to the appropriate package in `pkg/`
2. Write tests for new functionality
3. Update documentation as needed
4. Run tests using `make test`
5. Build and test the Docker image using `make docker-build`

## AWS Configuration

When running the containerized version, mount your AWS credentials:

```bash
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ami-migrate:latest \
  --new-ami ami-xxxxx \
  --latest-tag latest \
  --enabled-value enabled
```

## License

MIT License