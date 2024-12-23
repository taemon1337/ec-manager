package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/types"
)

type mockEC2Client struct {
	types.EC2Client
}

func TestGetEC2Client(t *testing.T) {
	ctx := context.Background()

	t.Run("mock mode", func(t *testing.T) {
		client := NewClient()
		client.SetMockMode(true)
		mockClient := &mockEC2Client{}
		client.SetEC2Client(mockClient)

		ec2Client, err := client.GetEC2Client(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, ec2Client)
	})

	t.Run("no client set in mock mode", func(t *testing.T) {
		client := NewClient()
		client.SetMockMode(true)

		ec2Client, err := client.GetEC2Client(ctx)
		assert.Error(t, err)
		assert.Nil(t, ec2Client)
		assert.Contains(t, err.Error(), "mock client not set")
	})
}

func TestSetMockMode(t *testing.T) {
	client := NewClient()
	mockClient := &mockEC2Client{}

	t.Run("enable mock mode", func(t *testing.T) {
		client.SetMockMode(true)
		client.SetEC2Client(mockClient)
		ec2Client, err := client.GetEC2Client(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, ec2Client)
	})

	t.Run("disable mock mode", func(t *testing.T) {
		client.SetMockMode(false)
		ec2Client, err := client.GetEC2Client(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, ec2Client)
	})
}

func TestLoadAWSConfig(t *testing.T) {
	t.Run("missing credentials", func(t *testing.T) {
		// Clear AWS credentials
		t.Setenv("AWS_ACCESS_KEY_ID", "")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "")

		// Test with missing credentials
		cfg, err := LoadAWSConfig(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})
}
