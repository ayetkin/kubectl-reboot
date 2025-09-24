# kubectl-reboot

[![Go Report Card](https://goreportcard.com/badge/github.com/ayetkin/kubectl-reboot)](https://goreportcard.com/report/github.com/ayetkin/kubectl-reboot)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A kubectl plugin that safely restarts Kubernetes nodes by draining pods, rebooting via SSH, verifying the reboot, and uncordoning the nodes.

## Features

- 🔄 **Safe Node Restart**: Automatically cordon, drain, reboot, and uncordon nodes
- 🚀 **SSH Integration**: Reboot nodes via SSH with customizable commands
- 🔍 **Reboot Verification**: Verify successful reboots by monitoring Boot ID changes
- 🌐 **Cluster-wide Operations**: Restart all nodes or specific subsets
- 🧪 **Dry-run Mode**: Preview operations without making changes
- ⚡ **Flexible Configuration**: Extensive customization options
- 📋 **Rich Logging**: Detailed, emoji-rich logging for better visibility

## Installation

### Option 1: Install via Krew (Recommended)

[Krew](https://krew.sigs.k8s.io/) is the plugin manager for kubectl command-line tool.

If you haven't installed Krew yet, follow the [official installation guide](https://krew.sigs.k8s.io/docs/user-guide/setup/install/).

Once Krew is installed, install kubectl-reboot:

```bash
kubectl krew install reboot
```

Verify the installation:

```bash
kubectl reboot --help
```

### Option 2: Manual Installation

#### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/ayetkin/kubectl-reboot/releases).

**Linux/macOS:**
```bash
# Download for your platform
curl -LO https://github.com/ayetkin/kubectl-reboot/releases/latest/download/kubectl-reboot-linux-amd64.tar.gz

# Extract
tar -xzf kubectl-reboot-linux-amd64.tar.gz

# Move to PATH
sudo mv kubectl-reboot /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/kubectl-reboot
```

#### Build from Source

```bash
git clone https://github.com/ayetkin/kubectl-reboot.git
cd kubectl-reboot
make build
sudo cp bin/kubectl-reboot /usr/local/bin/
```

## Usage

### Basic Examples

```bash
# Restart a single node
kubectl reboot node1

# Restart multiple nodes
kubectl reboot node1 node2 node3

# Restart all worker nodes (excluding control plane)
kubectl reboot --all --exclude-control-plane

# Dry run to see what would happen
kubectl reboot --all --exclude-control-plane --dry-run

# Restart nodes from a file
kubectl reboot --file nodes.txt

# Exclude specific nodes
kubectl reboot --all --exclude-nodes node1,node2
```

### Advanced Examples

```bash
# Custom SSH configuration
kubectl reboot --ssh-user ubuntu --ssh-opts "-i ~/.ssh/my-key" node1

# Custom reboot command
kubectl reboot --reboot-cmd "sudo shutdown -r now" node1

# Custom timeouts
kubectl reboot --timeout-ready 300 --timeout-bootid 600 node1

# Custom SSH host template (useful for cloud providers)
kubectl reboot --ssh-host-template "%s.us-west-2.compute.internal" node1

# Allow uncordon without reboot verification
kubectl reboot --allow-uncordon-without-reboot node1
```

## Configuration Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--all` | | `false` | Restart all nodes in the cluster |
| `--exclude-control-plane` | | `false` | Exclude control plane nodes when using --all |
| `--exclude-nodes` | | | Comma-separated node names to exclude |
| `--file` | `-f` | | Read node names from file (one per line) |
| `--ssh-user` | `-u` | `root` | SSH username |
| `--ssh-opts` | | See below | SSH connection options |
| `--ssh-host-template` | | `%s` | SSH host template (e.g., %s.example.com) |
| `--reboot-cmd` | | See below | Command to execute for reboot |
| `--timeout-ready` | | `180` | Timeout waiting for node to become ready (seconds) |
| `--timeout-bootid` | | `300` | Timeout waiting for boot ID change (seconds) |
| `--poll-interval` | | `10` | Polling interval (seconds) |
| `--allow-uncordon-without-reboot` | | `false` | Allow uncordon even if reboot verification fails |
| `--dry-run` | | `false` | Show what would be done without executing |
| `--context` | | | Kubeconfig context to use |
| `--kubeconfig` | | `$KUBECONFIG` | Path to kubeconfig file |

### Default Values

- **SSH Options**: `-o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=10`
- **Reboot Command**: `sudo systemctl reboot || sudo reboot`
- **Drain Arguments**: `--ignore-daemonsets --grace-period=30 --timeout=10m --delete-emptydir-data`

## How It Works

1. **Cordon**: Mark the node as unschedulable to prevent new pods
2. **Drain**: Evict all non-system pods from the node
3. **Reboot**: Execute reboot command via SSH
4. **Wait**: Monitor Boot ID change to verify reboot completion
5. **Ready**: Wait for the node to become ready
6. **Uncordon**: Mark the node as schedulable again

## Prerequisites

- Kubernetes cluster with SSH access to nodes
- kubectl configured and authenticated
- SSH access to target nodes (key-based authentication recommended)
- Appropriate RBAC permissions for node operations

### Required RBAC Permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubectl-reboot
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "patch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "delete"]
- apiGroups: ["policy"]
  resources: ["poddisruptionbudgets"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]
  resources: ["daemonsets", "replicasets"]
  verbs: ["get", "list"]
```

## Cloud Provider Examples

### AWS EKS

```bash
# Using private IPs with bastion host
kubectl reboot --ssh-user ec2-user \
  --ssh-opts "-i ~/.ssh/eks-key.pem -o ProxyCommand='ssh -i ~/.ssh/bastion.pem ec2-user@bastion-host -W %h:%p'" \
  ip-10-0-1-100

# Using public DNS names
kubectl reboot --ssh-user ec2-user \
  --ssh-host-template "%s.us-west-2.compute.amazonaws.com" \
  ip-10-0-1-100
```

### Google GKE

```bash
# Using gcloud compute ssh wrapper
kubectl reboot --ssh-user $USER \
  --reboot-cmd "gcloud compute instances reset \$(hostname) --zone=us-central1-a" \
  gke-cluster-default-pool-12345678-abcd
```

### Azure AKS

```bash
kubectl reboot --ssh-user azureuser \
  --ssh-host-template "%s.cloudapp.azure.com" \
  aks-nodepool1-12345678-vmss000000
```

## Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   ```bash
   # Test SSH connectivity first
   ssh -o StrictHostKeyChecking=no -o BatchMode=yes user@node
   
   # Check SSH key permissions
   chmod 600 ~/.ssh/your-key.pem
   ```

2. **Boot ID Not Changing**
   ```bash
   # Use flag to skip boot verification if needed
   kubectl reboot --allow-uncordon-without-reboot node1
   ```

3. **Pod Eviction Timeout**
   ```bash
   # Check for PodDisruptionBudgets that might block eviction
   kubectl get pdb --all-namespaces
   ```

4. **RBAC Permission Denied**
   ```bash
   # Check your permissions
   kubectl auth can-i get nodes
   kubectl auth can-i patch nodes
   kubectl auth can-i delete pods
   ```

### Logs and Debugging

The plugin provides detailed logging with emojis for better visibility:

- 🚀 Operation start
- 📋 Configuration details  
- ✅ Successful operations
- ⚠️  Warnings
- ❌ Errors
- 🧪 Dry-run operations

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make release

# Run tests
make test

# Format and vet code
make check
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Run `make check` to ensure code quality
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Security Considerations

- Use key-based SSH authentication instead of passwords
- Limit SSH access to specific users and source IPs
- Consider using SSH bastion hosts for additional security
- Review and understand the reboot commands being executed
- Test in non-production environments first

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [kubectl](https://kubernetes.io/docs/reference/kubectl/) - The Kubernetes command-line tool
- [Krew](https://krew.sigs.k8s.io/) - The kubectl plugin manager
- [Kubernetes](https://kubernetes.io/) - The container orchestration platform
