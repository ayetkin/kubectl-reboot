package main

import (
	"testing"

	"github.com/ayetkin/kubectl-reboot/internal/config"
)

func TestBuildSSHHost(t *testing.T) {
	tests := []struct {
		name     string
		template string
		node     string
		user     string
		expected string
	}{
		{
			name:     "basic hostname",
			template: "%s",
			node:     "node1",
			user:     "",
			expected: "node1",
		},
		{
			name:     "hostname with domain",
			template: "%s.example.com",
			node:     "node1",
			user:     "",
			expected: "node1.example.com",
		},
		{
			name:     "hostname with user",
			template: "%s",
			node:     "node1",
			user:     "ubuntu",
			expected: "ubuntu@node1",
		},
		{
			name:     "hostname with domain and user",
			template: "%s.example.com",
			node:     "node1",
			user:     "ubuntu",
			expected: "ubuntu@node1.example.com",
		},
		{
			name:     "template already contains user",
			template: "admin@%s.example.com",
			node:     "node1",
			user:     "ubuntu",
			expected: "admin@node1.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SSHHostTemplate: tt.template,
				SSHUser:         tt.user,
			}
			result := buildSSHHost(cfg, tt.node)
			if result != tt.expected {
				t.Errorf("buildSSHHost() = %v, want %v", result, tt.expected)
			}
		})
	}
}
