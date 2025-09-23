package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Nodes                      []string
	File                       string
	SSHUser                    string
	SSHIdentityFile            string
	SSHOpts                    string
	SSHHostTemplate            string
	RebootCmd                  string
	DrainArgs                  string
	TimeoutReadySeconds        int
	PollIntervalSeconds        int
	TimeoutBootIDSeconds       int
	AllowUncordonWithoutReboot bool
	KubeContext                string
	KubeconfigPath             string
	DryRun                     bool
	AllNodes                   bool
	ExcludeControlPlane        bool
	ExcludeNodes               []string // new
}

const (
	DefaultSSHOpts       = "-o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=10"
	DefaultRebootCmd     = "sudo systemctl reboot || sudo reboot"
	DefaultDrainArgs     = "--ignore-daemonsets --grace-period=30 --timeout=10m --delete-emptydir-data"
	DefaultReadyTimeout  = 180
	DefaultPollInterval  = 10
	DefaultBootIDTimeout = 300
)

func Parse() *Config {
	cfg := &Config{}
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	defaultKubeconfig := os.Getenv("KUBECONFIG")
	fs.StringVar(&cfg.File, "file", "", "read node names from file (one per line)")
	fs.StringVar(&cfg.File, "f", "", "read node names from file (one per line)")
	fs.StringVar(&cfg.SSHUser, "ssh-user", "", "SSH username")
	fs.StringVar(&cfg.SSHUser, "u", "", "SSH username")
	fs.StringVar(&cfg.SSHIdentityFile, "i", "", "SSH private key file")
	fs.StringVar(&cfg.SSHOpts, "ssh-opts", DefaultSSHOpts, "SSH options")
	fs.StringVar(&cfg.SSHHostTemplate, "ssh-host-template", "%s", "SSH host template (e.g., %s.example.com)")
	fs.StringVar(&cfg.RebootCmd, "reboot-cmd", DefaultRebootCmd, "reboot command to execute")
	fs.StringVar(&cfg.DrainArgs, "drain-args", DefaultDrainArgs, "kubectl drain arguments")
	fs.IntVar(&cfg.TimeoutReadySeconds, "timeout-ready", DefaultReadyTimeout, "timeout waiting for node to become ready (seconds)")
	fs.IntVar(&cfg.PollIntervalSeconds, "poll-interval", DefaultPollInterval, "polling interval (seconds)")
	fs.IntVar(&cfg.TimeoutBootIDSeconds, "timeout-bootid", DefaultBootIDTimeout, "timeout waiting for boot ID change (seconds)")
	fs.BoolVar(&cfg.AllowUncordonWithoutReboot, "allow-uncordon-without-reboot", false, "allow uncordon even if reboot verification fails")
	fs.BoolVar(&cfg.AllNodes, "all", false, "restart all nodes in the cluster")
	fs.BoolVar(&cfg.ExcludeControlPlane, "exclude-control-plane", false, "exclude control plane nodes when using --all")
	fs.StringVar(&cfg.KubeContext, "context", "", "kubeconfig context to use")
	fs.StringVar(&cfg.KubeconfigPath, "kubeconfig", defaultKubeconfig, "path to kubeconfig file")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "show what would be done without executing")
	var excludeNodesRaw string
	fs.StringVar(&excludeNodesRaw, "exclude-nodes", "", "comma-separated node names to exclude (e.g. node1,node2)")

	for _, a := range os.Args[1:] {
		if a == "-h" || a == "--help" {
			fmt.Fprintf(os.Stderr, `k8s-restart - Kubernetes Node Restart Tool

DESCRIPTION:
    Safely restart Kubernetes nodes by draining pods, rebooting via SSH,
    verifying the reboot, and uncordoning the nodes.

USAGE:
    k8s-restart [OPTIONS] [NODE_NAMES...]

EXAMPLES:
    # Restart specific nodes
    k8s-restart node1 node2

    # Restart all worker nodes (dry-run)
    k8s-restart --all --exclude-control-plane --dry-run

    # Restart nodes from file
    k8s-restart -f nodes.txt

    # Custom SSH settings
    k8s-restart -u myuser -i ~/.ssh/mykey node1

OPTIONS:
`)
			fs.PrintDefaults()
			os.Exit(0)
		}
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	cfg.Nodes = fs.Args()
	if excludeNodesRaw != "" {
		for _, p := range strings.Split(excludeNodesRaw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.ExcludeNodes = append(cfg.ExcludeNodes, p)
			}
		}
	}
	return cfg
}
