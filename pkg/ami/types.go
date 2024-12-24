package ami

import (
	"time"
)

// MigrationStatus represents the migration status of an instance
type MigrationStatus struct {
	InstanceID   string
	OSType       string
	InstanceType string
	State        string
	LaunchTime   time.Time
	PrivateIP    string
	PublicIP     string
	CurrentAMI   *AMIDetails
	LatestAMI    *AMIDetails
	NeedsMigrate bool
}

// AMIDetails represents details about an AMI
type AMIDetails struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
