package manager

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig validation failed: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				MaxMsgSize:        1024,
				KeepAliveTime:     time.Second,
				KeepAliveTimeout:  time.Second,
				MaxReconnectDelay: time.Second,
				MinConnectTimeout: time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid MaxMsgSize",
			config: &Config{
				MaxMsgSize:        0,
				KeepAliveTime:     time.Second,
				KeepAliveTimeout:  time.Second,
				MaxReconnectDelay: time.Second,
				MinConnectTimeout: time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid KeepAliveTime",
			config: &Config{
				MaxMsgSize:        1024,
				KeepAliveTime:     0,
				KeepAliveTimeout:  time.Second,
				MaxReconnectDelay: time.Second,
				MinConnectTimeout: time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConnectionManager(t *testing.T) {
	// Test with nil config and nil metrics
	cm, err := NewConnectionManager(nil, nil)
	if err != nil {
		t.Fatalf("NewConnectionManager with nil config failed: %v", err)
	}
	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}
	cm.Close()

	// Test with valid config
	cfg := DefaultConfig()
	cm, err = NewConnectionManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewConnectionManager failed: %v", err)
	}
	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}
	cm.Close()

	// Test with invalid config
	invalidCfg := &Config{
		MaxMsgSize: 0, // Invalid
	}
	cm, err = NewConnectionManager(invalidCfg, nil)
	if err == nil {
		t.Fatal("NewConnectionManager should fail with invalid config")
	}
}

func TestConnectionManager_GetConnectionsCount(t *testing.T) {
	cm, err := NewConnectionManager(nil, nil)
	if err != nil {
		t.Fatalf("NewConnectionManager failed: %v", err)
	}
	defer cm.Close()

	count := cm.GetConnectionsCount()
	if count != 0 {
		t.Errorf("Expected 0 connections, got %d", count)
	}
}

func TestConnectionManager_HealthCheck(t *testing.T) {
	cm, err := NewConnectionManager(nil, nil)
	if err != nil {
		t.Fatalf("NewConnectionManager failed: %v", err)
	}
	defer cm.Close()

	ctx := context.Background()
	health := cm.HealthCheck(ctx)
	if health == nil {
		t.Fatal("HealthCheck returned nil")
	}
	if len(health) != 0 {
		t.Errorf("Expected empty health map, got %d entries", len(health))
	}
}
