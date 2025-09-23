package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ayetkin/kubectl-reboot/internal/config"
	"github.com/ayetkin/kubectl-reboot/internal/kube"
	sshpkg "github.com/ayetkin/kubectl-reboot/internal/ssh"
	"github.com/charmbracelet/log"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetTimeFormat("")
	log.SetReportTimestamp(false)
	log.SetFormatter(log.TextFormatter)
}

func main() {
	cfg := config.Parse()
	kclient, err := kube.New(cfg.KubeconfigPath, cfg.KubeContext, log.Default())
	if err != nil {
		log.Fatalf("kube client: %v", err)
	}

	if cfg.AllNodes {
		nodes, err := kclient.ListNodeNames(cfg.ExcludeControlPlane)
		if err != nil {
			log.Fatalf("list nodes: %v", err)
		}
		cfg.Nodes = nodes
	} else if cfg.File != "" {
		fileNodes, err := readNodesFile(cfg.File)
		if err != nil {
			log.Fatalf("nodes file: %v", err)
		}
		cfg.Nodes = append(cfg.Nodes, fileNodes...)
	}

	if len(cfg.Nodes) == 0 {
		log.Fatal("no nodes provided")
	}
	if len(cfg.ExcludeNodes) > 0 {
		original := append([]string(nil), cfg.Nodes...)
		exset := map[string]struct{}{}
		for _, e := range cfg.ExcludeNodes {
			exset[e] = struct{}{}
		}
		filtered := make([]string, 0, len(cfg.Nodes))
		excluded := make([]string, 0)
		for _, n := range cfg.Nodes {
			if _, skip := exset[n]; skip {
				excluded = append(excluded, n)
				continue
			}
			filtered = append(filtered, n)
		}
		cfg.Nodes = filtered
		if len(excluded) > 0 {
			excludedList := "    " + strings.Join(excluded, "\n    ")
			log.Info("🚫 Excluded nodes", "count", len(excluded), "nodes", excludedList)
		} else {
			log.Infof("⚠️  --exclude-nodes provided but none matched the target node list")
		}

		missing := make([]string, 0)
		origSet := map[string]struct{}{}
		for _, n := range original {
			origSet[n] = struct{}{}
		}
		for _, want := range cfg.ExcludeNodes {
			if _, ok := origSet[want]; !ok {
				missing = append(missing, want)
			}
		}
		if len(missing) > 0 {
			missingList := "    " + strings.Join(missing, "\n    ")
			log.Warn("❓ Exclude nodes not found in target set", "missing", missingList)
		}
		if len(cfg.Nodes) == 0 {
			log.Fatal("❌ All nodes were excluded - no nodes to process")
		}
	}

	log.Info("🚀 Starting k8s-restart operation")

	// Format nodes list
	nodesList := strings.Join(cfg.Nodes, "\n    ")
	log.Info("📋 Target nodes", "count", len(cfg.Nodes), "nodes", "    "+nodesList)
	log.Info("🔧 Drain arguments", "args", cfg.DrainArgs)
	log.Info("🔑 SSH options", "opts", cfg.SSHOpts)
	if cfg.SSHIdentityFile != "" {
		log.Info("🗝️  SSH identity file", "path", cfg.SSHIdentityFile)
	}
	log.Info("🔄 Require reboot verification", "enabled", !cfg.AllowUncordonWithoutReboot)
	if cfg.AllNodes {
		log.Info("🌐 Processing all nodes", "exclude_control_plane", cfg.ExcludeControlPlane)
	}
	if cfg.DryRun {
		log.Info("🧪 DRY-RUN mode enabled - no actual changes will be made")
	}

	sshRunner := &sshpkg.Runner{DryRun: cfg.DryRun, Opts: cfg.SSHOpts, Key: cfg.SSHIdentityFile}

	log.Info("⏳ Initial wait before starting operations", "seconds", 5)
	time.Sleep(5 * time.Second)

	var failures []string
	for _, node := range cfg.Nodes {
		if err := processNode(cfg, kclient, sshRunner, node); err != nil {
			log.Error("❌ Node processing failed", "node", node, "error", err)
			failures = append(failures, node)
		}
	}
	if len(failures) > 0 {
		failuresList := "    " + strings.Join(failures, "\n    ")
		log.Error("💥 Operation failed", "failed_count", len(failures), "failed_nodes", failuresList)
		os.Exit(1)
	}
	log.Info("🎉 All nodes processed successfully! Operation completed.")
}

func processNode(cfg *config.Config, kc *kube.Client, ssh *sshpkg.Runner, nodeName string) error {
	ctx := context.Background()
	log.Info("⏳ Starting node restart process", "node", nodeName)
	nd, err := kc.GetNode(ctx, nodeName)
	if err != nil {
		return err
	}

	if !nd.Spec.Unschedulable {
		if cfg.DryRun {
			log.Info("🧪 DRY-RUN: Would cordon node", "node", nodeName)
		} else if err := kc.Cordon(ctx, nodeName); err != nil {
			return fmt.Errorf("cordon: %w", err)
		}
		log.Info("✅ Node cordoned - scheduling disabled", "node", nodeName)
	} else {
		log.Info("✅ Node already cordoned", "node", nodeName)
	}

	log.Info("⏳ Starting pod eviction process", "node", nodeName)
	if err := kc.EvictPods(ctx, nodeName, time.Duration(cfg.PollIntervalSeconds)*time.Second, 10*time.Minute, cfg.DryRun); err != nil {
		return fmt.Errorf("evict: %w", err)
	}
	log.Info("✅ Pod eviction completed successfully", "node", nodeName)

	bootBefore := nd.Status.NodeInfo.BootID

	log.Info("🔄 Initiating system reboot", "node", nodeName)
	sshHost := buildSSHHost(cfg, nodeName)
	if err := ssh.Run(sshHost, cfg.RebootCmd, log.Infof); err != nil {
		log.Warn("⚠️ SSH reboot command failed", "node", nodeName, "error", err)
	} else {
		log.Info("✅ Reboot command sent successfully", "node", nodeName)
	}

	if !cfg.DryRun {
		if bootBefore != "" {
			log.Info("⏳ Waiting for node reboot", "node", nodeName, "timeout_seconds", cfg.TimeoutBootIDSeconds)
			changed := kc.WaitForBootIDChange(ctx, nodeName, bootBefore, time.Duration(cfg.TimeoutBootIDSeconds)*time.Second, time.Duration(cfg.PollIntervalSeconds)*time.Second)
			if !changed {
				if !cfg.AllowUncordonWithoutReboot {
					return fmt.Errorf("❌ Boot ID unchanged - reboot may have failed")
				}
				log.Warn("⚠️ Boot ID unchanged, but proceeding due to flag", "node", nodeName, "flag", "allow-uncordon-without-reboot")
			} else {
				log.Info("✅ Reboot confirmed", "node", nodeName)
			}
		}

		log.Info("⏳ Waiting for node to become ready", "node", nodeName, "timeout_seconds", cfg.TimeoutReadySeconds)
		if !kc.WaitForCondition(ctx, nodeName, kube.IsNodeReady, time.Duration(cfg.TimeoutReadySeconds)*time.Second, time.Duration(cfg.PollIntervalSeconds)*time.Second) {
			return fmt.Errorf("❌ Node failed to become ready within timeout")
		}
		log.Info("✅ Node is ready", "node", nodeName)
	} else {
		log.Info("🧪 DRY-RUN: Skipping wait phases", "node", nodeName, "phases", "boot ID, ready")
	}

	// Uncordon
	if cfg.DryRun {
		log.Info("🧪 DRY-RUN: Would uncordon node", "node", nodeName)
	} else if err := kc.Uncordon(ctx, nodeName); err != nil {
		return fmt.Errorf("uncordon: %w", err)
	}
	log.Info("🎉 Node restart process completed successfully", "node", nodeName)
	return nil
}

func buildSSHHost(cfg *config.Config, node string) string {
	host := fmt.Sprintf(cfg.SSHHostTemplate, node)
	if cfg.SSHUser != "" && !strings.Contains(host, "@") {
		host = cfg.SSHUser + "@" + host
	}
	return host
}

func readNodesFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var nodes []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nodes = append(nodes, line)
	}
	return nodes, sc.Err()
}
