# AMI Migration Tool

A Go-based tool for managing AWS AMI migrations. This tool helps automate the process of updating EC2 instances to use a new AMI while maintaining proper versioning and backups.

## Features

- Automatic AMI version tracking using tags
- Parallel instance migration
- Volume snapshots for backup
- Graceful error handling and status tracking
- Clean separation of concerns with pkg structure
- Docker-based build and test environment

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
  ami-migrate:latest --new-ami ami-xxxxx
```

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
  ami-migrate:latest --new-ami ami-xxxxx
```

## License

MIT License