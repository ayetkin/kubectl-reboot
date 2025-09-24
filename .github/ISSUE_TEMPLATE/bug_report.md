---
name: Bug Report
about: Create a report to help us improve kubectl-reboot
title: '[BUG] '
labels: ['bug']
assignees: ''
---

## Bug Description
A clear and concise description of what the bug is.

## Environment
- **kubectl-reboot version**: `kubectl reboot --version`
- **Kubernetes version**: `kubectl version --short`
- **Operating System**: (e.g., Ubuntu 20.04, macOS 13.0, Windows 11)
- **Architecture**: (e.g., amd64, arm64)
- **Installation method**: (Krew, direct download, build from source)

## Node Information
- **Node OS**: (e.g., Ubuntu 20.04, Amazon Linux 2, Windows Server)
- **Container Runtime**: (e.g., containerd, docker)
- **Cloud Provider**: (e.g., AWS EKS, Google GKE, Azure AKS, on-premises)

## SSH Configuration
- **SSH User**: 
- **SSH Options**: 
- **Authentication Method**: (key-based, password, etc.)
- **SSH Host Template**: (if used)

## Command Used
```bash
# Paste the exact kubectl-reboot command you ran
kubectl reboot ...
```

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Error Output
```
# Paste the complete error output or logs here
```

## Steps to Reproduce
1. Set up environment with...
2. Configure SSH access to...
3. Run command `kubectl reboot ...`
4. Observe error...

## Workaround
If you found a workaround, please describe it here.

## Additional Context
Add any other context about the problem here, such as:
- Network configuration
- Firewall rules
- Special cluster setup
- Other kubectl plugins or tools in use

## Logs
If possible, include relevant logs from:
- kubectl-reboot output (use `--verbose` if available)
- Kubernetes events: `kubectl get events`
- Node status: `kubectl describe node <node-name>`

## Screenshots
If applicable, add screenshots to help explain your problem.