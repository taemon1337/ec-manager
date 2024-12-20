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
- Automatic AMI version management (latest/previous)

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
  migrate \
  --new-ami ami-xxxxx
```

## CLI Arguments

### Global Flags
- `--enabled-value` (optional, default: "enabled"): Value to match for the "ami-migrate" tag

### Create Command
- `--user` (required): Your user ID for instance ownership
- `--os` (optional, default: "Ubuntu"): OS type (Ubuntu or RHEL9)
- `--size` (optional, default: "large"): Instance size (small, medium, large, xlarge)
- `--name` (optional): Instance name (default: randomly generated)

Example:
```bash
# Create a default Ubuntu instance
ami-migrate create --user johndoe

# Create a custom RHEL instance
ami-migrate create --user johndoe --os RHEL9 --size xlarge --name my-instance
```

### List Command
- `--user` (required): Your user ID to list your instances

Example:
```bash
ami-migrate list --user johndoe
```

Output:
```
Found 2 instance(s):

Instance: happy-penguin-123 (i-1234567890abcdef0)
  OS:           Ubuntu
  Size:         t3.large
  State:        running
  Launch Time:  2024-12-19T20:00:00Z
  Private IP:   10.0.0.100
  Public IP:    54.123.45.67
  Current AMI:  ami-0abc123def456
  Latest AMI:   ami-0xyz789uvw123 (migration available)

Instance: clever-falcon-456 (i-0987654321fedcba0)
  OS:           RHEL9
  Size:         t3.xlarge
  State:        stopped
  Launch Time:  2024-12-18T15:30:00Z
  Private IP:   10.0.0.200
  Current AMI:  ami-0def456abc789
```

### Check Command
- `--user` (required): Your user ID to find your instance (matches Owner tag)

Example:
```bash
ami-migrate check --user johndoe
```

Output:
```
Instance Status for i-1234567890abcdef0:
  OS Type:        RHEL9
  Instance Type:  t3.micro
  State:          running
  Launch Time:    2024-12-19T20:00:00Z
  Private IP:     10.0.0.100
  Public IP:      54.123.45.67

AMI Status:
  Current AMI:    ami-0abc123def456
    Name:         RHEL-9.2-20231201
    Created:      2023-12-01T00:00:00Z
  Latest AMI:     ami-0xyz789uvw123
    Name:         RHEL-9.2-20231219
    Created:      2023-12-19T00:00:00Z

Migration Needed: true

Run 'ami-migrate migrate' to update your instance to the latest AMI.
```

### Migrate Command
- `--new-ami` (required): The ID of the new AMI to upgrade instances to
- `--instance-id` (optional): ID of specific instance to migrate (bypasses tag requirements)

### Backup Command
- `--instance-id` (optional): ID of specific instance to backup (bypasses tag requirements)

### Delete Command
- `--user` (required): Your user ID
- `--instance` (required): Instance ID to delete

Example:
```bash
ami-migrate delete --user johndoe --instance i-1234567890abcdef0
```

Output:
```
About to delete the following instance:

Instance: happy-penguin-123 (i-1234567890abcdef0)
  OS:           Ubuntu
  Size:         t3.large
  State:        running
  Launch Time:  2024-12-19T20:00:00Z
  Private IP:   10.0.0.100
  Public IP:    54.123.45.67
  Current AMI:  ami-0abc123def456

WARNING: This action cannot be undone!
Delete this instance? [y/N]: y

Deleting instance i-1234567890abcdef0...

Instance i-1234567890abcdef0 deletion initiated successfully!
Note: It may take a few minutes for the instance to be fully terminated.

Run 'ami-migrate list --user johndoe' to check instance status.
```

## Instance Management

The tool provides a complete set of commands for managing EC2 instances:

### Creating Instances

Create instances with sensible defaults:
```bash
# Create Ubuntu instance (t3.large)
ami-migrate create --user johndoe

# Create RHEL instance with custom size
ami-migrate create --user johndoe --os RHEL9 --size xlarge

# Create instance with specific name
ami-migrate create --user johndoe --name web-server-1
```

Instance sizes map to AWS instance types:
- small: t3.small
- medium: t3.medium
- large: t3.large (default)
- xlarge: t3.xlarge

### Listing Instances

View all your instances and their status:
```bash
ami-migrate list --user johndoe
```

The list command shows:
- Instance name and ID
- OS type and size
- Current state (running/stopped)
- IP addresses
- Launch time
- Current and latest AMI IDs
- Migration status

### Checking Migration Status

Check if instances need migration:
```bash
ami-migrate check --user johndoe
```

The check command provides:
- Current AMI details
- Latest available AMI
- OS version information
- Migration recommendation

### Deleting Instances

Delete instances safely:
```bash
ami-migrate delete --user johndoe --instance i-1234567890abcdef0
```

The delete command:
1. Verifies instance ownership
2. Shows instance details
3. Requires confirmation
4. Initiates termination
5. Provides status updates

## How It Works

1. **AMI Version Management**:
   - The tool tracks AMI versions using tags
   - When a new AMI is specified, the old "latest" AMI is marked as "previous"
   - The new AMI becomes the "latest" version
   - If no AMI is marked as "latest", the new AMI is used as both old and new

2. **Instance Selection**:
   - Instances are selected for migration if they have `ami-migrate=enabled`
   - Running instances require both `ami-migrate=enabled` and `ami-migrate-if-running=enabled`
   - Stopped instances only require `ami-migrate=enabled`

3. **Migration Process**:
   - Each instance's volumes are snapshotted before migration
   - The instance is stopped if running
   - A new instance is created with the target AMI
   - All tags are copied to the new instance
   - The old instance is terminated
   - If the instance was running, it is started again
   - Comprehensive error handling and status tracking

## Instance Tagging

Three tags control the instance behavior:

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

3. Owner Tag (Required for check command):
```
Key: Owner
Value: <your-user-id>  # Used to identify your instance
```

Tag Combinations and Behavior:
- Running instances:
  - Requires BOTH `ami-migrate=enabled` AND `ami-migrate-if-running=enabled`
  - Will be skipped if missing either tag
- Stopped instances:
  - Only requires `ami-migrate=enabled`
  - Will be migrated regardless of `ami-migrate-if-running` tag

This ensures that:
1. Running instances are only migrated when explicitly allowed via both tags
2. Stopped instances can be safely migrated with just the main migration tag

## Migration Status Tracking

The tool tracks migration status using the following tags:

1. Status Tag:
```
Key: ami-migrate-status
Value: [status]  # One of: skipped, in-progress, failed, warning, completed
```

2. Message Tag:
```
Key: ami-migrate-message
Value: [detailed message]  # Explains the current status
```

Status Values:
- `skipped`: Instance was not migrated (e.g., running instance without required tags)
- `in-progress`: Migration has started
- `failed`: Migration failed (error message in ami-migrate-message)
- `warning`: Migration partially successful (e.g., migrated but failed to start)
- `completed`: Migration completed successfully

These tags provide a clear audit trail of the migration process and help identify any issues that need attention.

## Development

### Testing

The codebase includes comprehensive tests:
- Unit tests for all service functions
- Mock EC2 client for AWS operations
- Test coverage for error cases
- Integration test examples

Run tests:
```bash
go test ./... -v
```

### Project Structure

```
ami-migrate/
├── cmd/               # CLI commands
│   ├── backup.go      # Volume snapshot backup
│   ├── check.go       # Migration status check
│   ├── create.go      # Instance creation
│   ├── delete.go      # Instance deletion
│   ├── list.go        # Instance listing
│   ├── migrate.go     # AMI migration
│   └── root.go        # Root command and flags
├── pkg/
│   └── ami/          # Core functionality
│       ├── ami.go     # AWS operations
│       ├── ami_test.go # Unit tests
│       └── mock_ec2.go # Mock AWS client
```

### Adding New Features

When adding new features:
1. Add service functions in `pkg/ami/ami.go`
2. Add unit tests in `pkg/ami/ami_test.go`
3. Update mock client if needed
4. Create CLI command in `cmd/`
5. Update documentation

## Usage

The tool provides five main commands:

### 1. Check

Check the status of your instance and determine if a migration is needed:

```bash
ami-migrate check --user johndoe
```

### 2. Create

Create a new instance:

```bash
ami-migrate create --user johndoe
```

### 3. List

List your instances:

```bash
ami-migrate list --user johndoe
```

### 4. Migrate

Migrate instances to a new AMI version:

```bash
# Migrate instances by tag
ami-migrate migrate --new-ami ami-xxxxx

# Migrate specific instance
ami-migrate migrate --new-ami ami-xxxxx --instance-id i-xxxxx
```

Optional flags:
- `--enabled-value`: Value to match for the ami-migrate tag (default: "enabled")
- `--instance-id`: ID of specific instance to migrate (bypasses tag requirements)

### 5. Backup

Create snapshots of all volumes attached to instances:

```bash
# Backup instances by tag
ami-migrate backup

# Backup specific instance
ami-migrate backup --instance-id i-xxxxx
```

Optional flags:
- `--enabled-value`: Value to match for the ami-migrate tag (default: "enabled")
- `--instance-id`: ID of specific instance to backup (bypasses tag requirements)

The backup command will:
1. Find all instances with the ami-migrate tag (or use specified instance)
2. Create snapshots of all attached volumes
3. Tag snapshots with instance and device information

### 6. Delete

Delete an instance:

```bash
ami-migrate delete --user johndoe --instance i-1234567890abcdef0
```

### CI/CD Integration

For GitLab CI, add this to your `.gitlab-ci.yml`:

```yaml
ami-migrate:
  image: golang:1.21-alpine
  script:
    - go install github.com/taemon1337/ami-migrate@latest
    - ami-migrate migrate --new-ami $NEW_AMI_ID
  rules:
    - if: $CI_COMMIT_TAG  # Only run on tags
```

Make sure to set these environment variables in GitLab:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION`
- `NEW_AMI_ID`

## AWS Configuration

When running the containerized version, mount your AWS credentials:

```bash
docker run --rm \
  -v ~/.aws:/root/.aws:ro \
  ami-migrate:latest \
  migrate \
  --new-ami ami-xxxxx
```

## License

MIT License