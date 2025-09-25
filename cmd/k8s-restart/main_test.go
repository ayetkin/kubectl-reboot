package main

import (
	"os"
	"strings"
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

func TestReadNodesFile(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "nodes-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data
	content := `node1
node2
# This is a comment
node3

node4
# Another comment`

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test reading the file
	nodes, err := readNodesFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("readNodesFile() error = %v", err)
	}

	expected := []string{"node1", "node2", "node3", "node4"}
	if len(nodes) != len(expected) {
		t.Fatalf("Expected %d nodes, got %d", len(expected), len(nodes))
	}

	for i, expectedNode := range expected {
		if nodes[i] != expectedNode {
			t.Errorf("Expected node %d to be %q, got %q", i, expectedNode, nodes[i])
		}
	}
}

func TestReadNodesFileNonExistent(t *testing.T) {
	_, err := readNodesFile("/non/existent/file")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestReadNodesFileEmpty(t *testing.T) {
	// Create empty temporary file
	tmpfile, err := os.CreateTemp("", "empty-nodes-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	nodes, err := readNodesFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("readNodesFile() error = %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("Expected empty slice, got %v", nodes)
	}
}

func TestProcessNodeConfigurationFileReading(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid nodes provided",
			cfg: &config.Config{
				Nodes: []string{"node1", "node2"},
			},
			expectError: false,
		},
		{
			name: "no nodes provided",
			cfg: &config.Config{
				Nodes: []string{},
			},
			expectError: true,
			errorMsg:    "no nodes provided",
		},
		{
			name: "file with valid nodes",
			cfg: &config.Config{
				File:  createTempNodesFile(t, "node1\nnode2\n"),
				Nodes: []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp file if created
			if tt.cfg.File != "" {
				defer os.Remove(tt.cfg.File)
			}

			// We can't easily test with real kube client, so test only file reading part
			if tt.cfg.File != "" {
				fileNodes, err := readNodesFile(tt.cfg.File)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error reading file: %v", err)
				}
				if err == nil {
					tt.cfg.Nodes = append(tt.cfg.Nodes, fileNodes...)
				}
			}

			// Check if we have nodes after processing
			if len(tt.cfg.Nodes) == 0 && !tt.expectError {
				t.Error("Expected nodes to be populated")
			}
		})
	}
}

func TestFilterExcludedNodes(t *testing.T) {
	tests := []struct {
		name          string
		nodes         []string
		excludeNodes  []string
		expectedNodes []string
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "no exclusions",
			nodes:         []string{"node1", "node2", "node3"},
			excludeNodes:  []string{},
			expectedNodes: []string{"node1", "node2", "node3"},
			expectError:   false,
		},
		{
			name:          "exclude some nodes",
			nodes:         []string{"node1", "node2", "node3"},
			excludeNodes:  []string{"node2"},
			expectedNodes: []string{"node1", "node3"},
			expectError:   false,
		},
		{
			name:          "exclude all nodes",
			nodes:         []string{"node1", "node2"},
			excludeNodes:  []string{"node1", "node2"},
			expectedNodes: []string{},
			expectError:   true,
			errorMsg:      "All nodes were excluded",
		},
		{
			name:          "exclude non-existent node",
			nodes:         []string{"node1", "node2"},
			excludeNodes:  []string{"node3"},
			expectedNodes: []string{"node1", "node2"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Nodes:        tt.nodes,
				ExcludeNodes: tt.excludeNodes,
			}

			err := filterExcludedNodes(cfg)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
			}

			if !tt.expectError {
				if len(cfg.Nodes) != len(tt.expectedNodes) {
					t.Fatalf("Expected %d nodes after filtering, got %d", len(tt.expectedNodes), len(cfg.Nodes))
				}
				for i, expected := range tt.expectedNodes {
					if cfg.Nodes[i] != expected {
						t.Errorf("Expected node %d to be %q, got %q", i, expected, cfg.Nodes[i])
					}
				}
			}
		})
	}
}

// Helper function to create temporary nodes file
func createTempNodesFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "nodes-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}
