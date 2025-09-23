package ssh

import (
	"fmt"
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
		Auth:            r.getAuthMethods(),
	}

	// Connect and run command
	client, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		return fmt.Errorf("❌ SSH connection failed to %s: %v", host, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("❌ SSH session failed: %v", err)
	}
	defer session.Close()

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
