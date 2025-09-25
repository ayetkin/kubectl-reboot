package ssh

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestParseHost(t *testing.T) {
	tests := []struct {
		name             string
		host             string
		expectedUser     string
		expectedHostname string
	}{
		{
			name:             "hostname only",
			host:             "server.example.com",
			expectedUser:     "root",
			expectedHostname: "server.example.com",
		},
		{
			name:             "user and hostname",
			host:             "ubuntu@server.example.com",
			expectedUser:     "ubuntu",
			expectedHostname: "server.example.com",
		},
		{
			name:             "IP address only",
			host:             "192.168.1.100",
			expectedUser:     "root",
			expectedHostname: "192.168.1.100",
		},
		{
			name:             "user and IP address",
			host:             "admin@192.168.1.100",
			expectedUser:     "admin",
			expectedHostname: "192.168.1.100",
		},
		{
			name:             "localhost",
			host:             "localhost",
			expectedUser:     "root",
			expectedHostname: "localhost",
		},
		{
			name:             "user with localhost",
			host:             "testuser@localhost",
			expectedUser:     "testuser",
			expectedHostname: "localhost",
		},
		{
			name:             "empty string",
			host:             "",
			expectedUser:     "root",
			expectedHostname: "",
		},
		{
			name:             "multiple @ symbols",
			host:             "user@domain@server.com",
			expectedUser:     "user",
			expectedHostname: "domain@server.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, hostname := parseHost(tt.host)

			if user != tt.expectedUser {
				t.Errorf("parseHost(%q) user = %q, want %q", tt.host, user, tt.expectedUser)
			}

			if hostname != tt.expectedHostname {
				t.Errorf("parseHost(%q) hostname = %q, want %q", tt.host, hostname, tt.expectedHostname)
			}
		})
	}
}

func TestRunnerDryRun(t *testing.T) {
	runner := &Runner{
		DryRun: true,
		Opts:   "-o StrictHostKeyChecking=no",
		Key:    "",
	}

	// Capture log output
	var loggedMessages []string
	logf := func(format string, args ...any) {
		loggedMessages = append(loggedMessages, format)
	}

	err := runner.Run("testhost", "sudo reboot", logf)

	if err != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", err)
	}

	if len(loggedMessages) == 0 {
		t.Error("Expected log message in dry-run mode")
	}

	// Check if log message contains expected dry-run indicator
	found := false
	for _, msg := range loggedMessages {
		if containsPattern(msg, "dry-run") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected dry-run log message not found")
	}
}

func TestRunnerGetAuthMethods(t *testing.T) {
	tests := []struct {
		name        string
		keyPath     string
		expectEmpty bool
	}{
		{
			name:        "no key provided",
			keyPath:     "",
			expectEmpty: true,
		},
		{
			name:        "non-existent key file",
			keyPath:     "/non/existent/key",
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{
				Key: tt.keyPath,
			}

			authMethods := runner.getAuthMethods()

			if tt.expectEmpty && len(authMethods) != 0 {
				t.Errorf("Expected empty auth methods, got %d methods", len(authMethods))
			}
		})
	}
}

func TestHostKeyCallback(t *testing.T) {
	runner := &Runner{}
	callback := runner.getHostKeyCallback()

	// Create a mock public key (we can use nil since the callback doesn't actually use it)
	// But we need to handle the nil case properly in the callback
	err := callback("test.example.com", nil, &mockPublicKey{})
	if err != nil {
		t.Errorf("Expected host key callback to return nil, got: %v", err)
	}
}

// Mock public key for testing
type mockPublicKey struct{}

func (m *mockPublicKey) Type() string {
	return "test-key-type"
}

func (m *mockPublicKey) Marshal() []byte {
	return []byte("mock-key-data")
}

func (m *mockPublicKey) Verify(data []byte, sig *ssh.Signature) error {
	return nil
}

// Helper function to check if a string contains a pattern (case-insensitive)
func containsPattern(s, pattern string) bool {
	return len(s) >= len(pattern) && findSubstring(s, pattern)
}

func findSubstring(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if toLower(s[i+j]) != toLower(pattern[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
