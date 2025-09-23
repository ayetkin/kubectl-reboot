#!/bin/bash
set -euo pipefail

# This script helps submit the plugin to the krew-index repository

PLUGIN_NAME="reboot"
VERSION="${1:-}"
REPO_URL="https://github.com/ayetkin/kubectl-reboot"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

echo "🚀 Preparing to submit kubectl-reboot $VERSION to krew-index"

# Check if required files exist
if [ ! -f "kubectl-reboot.yaml" ]; then
    echo "❌ kubectl-reboot.yaml not found"
    exit 1
fi

if [ ! -d "dist" ]; then
    echo "❌ dist directory not found. Run 'make package' first"
    exit 1
fi

# Clone krew-index if it doesn't exist
if [ ! -d "krew-index" ]; then
    echo "📦 Cloning krew-index repository..."
    git clone https://github.com/kubernetes-sigs/krew-index.git
fi

cd krew-index

# Update krew-index
echo "🔄 Updating krew-index..."
git checkout master
git pull origin master

# Create a new branch
BRANCH_NAME="add-kubectl-reboot-$VERSION"
git checkout -b "$BRANCH_NAME"

# Copy plugin manifest
echo "📋 Copying plugin manifest..."
cp "../kubectl-reboot.yaml" "plugins/$PLUGIN_NAME.yaml"

# Update the manifest with actual checksums from dist/checksums.txt
if [ -f "../dist/checksums.txt" ]; then
    echo "🔑 Updating checksums in manifest..."
    
    # Read checksums and update the manifest
    while IFS= read -r line; do
        checksum=$(echo "$line" | awk '{print $1}')
        filename=$(echo "$line" | awk '{print $2}' | sed 's|.*/||')
        
        case "$filename" in
            kubectl-reboot-linux-amd64.tar.gz)
                sed -i.bak "s/TO_BE_FILLED/$checksum/" "plugins/$PLUGIN_NAME.yaml"
                ;;
            kubectl-reboot-linux-arm64.tar.gz)
                sed -i.bak "s/TO_BE_FILLED/$checksum/" "plugins/$PLUGIN_NAME.yaml"
                ;;
            kubectl-reboot-darwin-amd64.tar.gz)
                sed -i.bak "s/TO_BE_FILLED/$checksum/" "plugins/$PLUGIN_NAME.yaml"
                ;;
            kubectl-reboot-darwin-arm64.tar.gz)
                sed -i.bak "s/TO_BE_FILLED/$checksum/" "plugins/$PLUGIN_NAME.yaml"
                ;;
            kubectl-reboot-windows-amd64.zip)
                sed -i.bak "s/TO_BE_FILLED/$checksum/" "plugins/$PLUGIN_NAME.yaml"
                ;;
        esac
    done < "../dist/checksums.txt"
    
    # Clean up backup files
    rm -f "plugins/$PLUGIN_NAME.yaml.bak"
else
    echo "⚠️  No checksums.txt found. You'll need to update the manifest manually."
fi

# Validate the plugin
echo "✅ Validating plugin manifest..."
kubectl krew install --manifest="plugins/$PLUGIN_NAME.yaml" --archive="../dist/kubectl-reboot-linux-amd64.tar.gz" || {
    echo "❌ Plugin validation failed"
    exit 1
}

# Uninstall after validation
kubectl krew uninstall reboot 2>/dev/null || true

# Commit changes
echo "📝 Committing changes..."
git add "plugins/$PLUGIN_NAME.yaml"
git commit -m "Add kubectl-reboot $VERSION

kubectl-reboot is a kubectl plugin that safely reboots Kubernetes nodes
by draining pods, rebooting via SSH, verifying the reboot, and uncordoning
the nodes.

Features:
- Safe node restart workflow
- SSH integration for rebooting
- Reboot verification via Boot ID monitoring
- Cluster-wide operations
- Dry-run mode
- Extensive configuration options

Source: $REPO_URL
"

echo "🎉 Ready to submit! Next steps:"
echo ""
echo "1. Push the branch:"
echo "   cd krew-index"
echo "   git push origin $BRANCH_NAME"
echo ""
echo "2. Open a Pull Request at:"
echo "   https://github.com/kubernetes-sigs/krew-index/compare/master...$BRANCH_NAME"
echo ""
echo "3. Wait for the krew-index maintainers to review and merge"
echo ""
echo "📁 The plugin manifest is located at:"
echo "   krew-index/plugins/$PLUGIN_NAME.yaml"
