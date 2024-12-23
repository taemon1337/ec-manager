package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taemon1337/ec-manager/pkg/types"
)

type mockEC2Client struct {
	types.EC2Client
}

func TestGetEC2Client(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "mock_mode",
			setup: func() {
				mockClient := &types.MockEC2Client{}
				SetMockMode(true)
				SetMockClient(mockClient)
			},
		},
		{
			name: "no_client_set_in_mock_mode",
			setup: func() {
				SetMockMode(true)
				SetMockClient(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				SetMockMode(false)
				SetMockClient(nil)
			}()

			tt.setup()

			client := NewClient()
			_, err := client.GetEC2Client(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// mockConfigLoadError is used to simulate AWS config loading errors
type mockConfigLoadError struct {
	err error
}

func (m *mockConfigLoadError) LoadDefaultConfig(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
	if m.err != nil {
		return aws.Config{}, m.err
	}
	return aws.Config{}, nil
}

func TestLoadAWSConfig(t *testing.T) {
	originalConfigLoader := configLoader
	defer func() {
		configLoader = originalConfigLoader
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "missing_credentials",
			setup: func() {
				configLoader = &mockConfigLoadError{
					err: fmt.Errorf("SharedConfigProfile: invalid credentials"),
				}
			},
			wantErr: true,
		},
		{
			name: "successful_config_load",
			setup: func() {
				configLoader = &mockConfigLoadError{}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			_, err := LoadAWSConfig(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
