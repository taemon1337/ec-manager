package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("with mock mode", func(t *testing.T) {
		client, err := NewClient(true, "", "us-east-1")
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	// Skip tests that require AWS credentials
	t.Run("with default config", func(t *testing.T) {
		t.Skip("Skipping test that requires AWS credentials")
	})
}

func TestLoadAWSConfig(t *testing.T) {
	// Skip tests that require AWS credentials
	t.Run("with default config", func(t *testing.T) {
		t.Skip("Skipping test that requires AWS credentials")
	})

	t.Run("with profile", func(t *testing.T) {
		t.Skip("Skipping test that requires AWS credentials")
	})
}

func TestGetEC2Client(t *testing.T) {
	t.Run("with mock mode", func(t *testing.T) {
		client, err := NewClient(true, "", "us-east-1")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		ec2Client := client.GetEC2Client()
		assert.NotNil(t, ec2Client)
	})

	// Skip tests that require AWS credentials
	t.Run("with real client", func(t *testing.T) {
		t.Skip("Skipping test that requires AWS credentials")
	})
}

func TestGetAMIService(t *testing.T) {
	t.Run("with mock mode", func(t *testing.T) {
		client, err := NewClient(true, "", "us-east-1")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		amiService := client.GetAMIService()
		assert.NotNil(t, amiService)
	})

	// Skip tests that require AWS credentials
	t.Run("with real client", func(t *testing.T) {
		t.Skip("Skipping test that requires AWS credentials")
	})
}
