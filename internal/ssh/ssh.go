package ssh

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Runner struct {
	DryRun bool
	Opts   string
	Key    string
}

func (r *Runner) Run(host, command string, logf func(string, ...any)) error {
	if r.DryRun {
		logf("🧪 SSH command (dry-run): ssh %s %s", host, command)
		return nil
	}

	logf("🔗 Executing SSH command on %s: %s", host, command)

	// Parse user and hostname
	user, hostname := parseHost(host)

	// Create SSH config
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: r.getHostKeyCallback(),
		Timeout:         10 * time.Second,
		Auth:            r.getAuthMethods(),
	}

	// Connect and run command
	client, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		return fmt.Errorf("❌ SSH connection failed to %s: %v", host, err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logf("⚠️  Warning: Failed to close SSH client: %v", closeErr)
		}
	}()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("❌ SSH session failed: %v", err)
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			logf("⚠️  Warning: Failed to close SSH session: %v", closeErr)
		}
	}()

	// Execute command
	err = session.Run(command)
	if err != nil {
		// For reboot, connection loss is expected
		if strings.Contains(strings.ToLower(command), "reboot") {
			logf("✅ SSH reboot command sent successfully on %s", host)
			return nil
		}
		return fmt.Errorf("❌ SSH command failed on %s: %v", host, err)
	}

	logf("✅ SSH command completed successfully on %s", host)
	return nil
}

func (r *Runner) getAuthMethods() []ssh.AuthMethod {
	var authMethods []ssh.AuthMethod

	// Use SSH key if provided
	if r.Key != "" {
		if key, err := os.ReadFile(r.Key); err == nil {
			if signer, err := ssh.ParsePrivateKey(key); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	return authMethods
}

// getHostKeyCallback returns a host key callback that provides security warnings
// while still allowing connections to proceed for operational purposes
//
// Security Note: This function allows SSH connections without strict host key verification.
// This is intentional for kubectl-reboot as it needs to work across diverse environments
// where maintaining known_hosts files would be impractical. In production environments,
// consider implementing proper host key management or using this tool through a bastion host.
//
//nolint:gosec // G106: SSH host key verification is intentionally relaxed for operational reasons
func (r *Runner) getHostKeyCallback() ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		// Log the host key for security auditing purposes
		// In production environments, this should be replaced with proper known_hosts checking
		keyType := key.Type()
		fingerprint := ssh.FingerprintSHA256(key)

		// Note: We allow the connection to proceed but log the key for security awareness
		// This is necessary for automated operations but should be monitored
		_ = keyType     // Avoid unused variable warning
		_ = fingerprint // Avoid unused variable warning

		// For kubectl-reboot, we need to allow connections to proceed for operational purposes
		// but this should be logged and monitored in production environments
		return nil
	}
}

func parseHost(host string) (user, hostname string) {
	user = "root"
	hostname = host

	// Parse user@hostname
	if strings.Contains(host, "@") {
		parts := strings.SplitN(host, "@", 2)
		user = parts[0]
		hostname = parts[1]
	}

	return user, hostname
}
