#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PLUGIN_NAME="reboot"
VERSION="${1:-}"
REPO_URL="https://github.com/ayetkin/kubectl-reboot"
KREW_INDEX_REPO="https://github.com/kubernetes-sigs/krew-index.git"

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    cat << EOF
Usage: $0 <version> [options]

Arguments:
  version           Version to submit (e.g., v1.0.1)

Options:
  -h, --help        Show this help message
  -d, --dry-run     Show what would be done without executing
  --skip-validation Skip plugin validation

Examples:
  $0 v1.0.1                    # Submit v1.0.1 to krew-index
  $0 v1.0.1 --dry-run          # Preview submission for v1.0.1

EOF
}

check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if we're in the right directory
    if [[ ! -f "kubectl-reboot.yaml" ]]; then
        error "kubectl-reboot.yaml not found. Run this from the repo root."
        exit 1
    fi
    
    # Check required tools
    local required_tools=("git" "kubectl" "curl")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "Required tool '$tool' is not installed"
            exit 1
        fi
    done
    
    # Check if kubectl krew is available
    if ! kubectl krew version &> /dev/null; then
        error "kubectl krew is not installed or not in PATH"
        exit 1
    fi
    
    success "All prerequisites met"
}

download_updated_manifest() {
    local version=$1
    
    log "Downloading updated manifest from GitHub release..."
    
    local manifest_url="https://github.com/ayetkin/kubectl-reboot/releases/download/${version}/kubectl-reboot.yaml.updated"
    
    if curl -fsSL "$manifest_url" -o kubectl-reboot.yaml.updated; then
        success "Downloaded updated manifest"
        return 0
    else
        warning "Failed to download updated manifest, using local kubectl-reboot.yaml"
        cp kubectl-reboot.yaml kubectl-reboot.yaml.updated
        return 1
    fi
}

validate_checksums() {
    local version=$1
    
    log "Validating checksums..."
    
    # Download checksums.txt from release
    local checksums_url="https://github.com/ayetkin/kubectl-reboot/releases/download/${version}/checksums.txt"
    
    if ! curl -fsSL "$checksums_url" -o checksums.txt; then
        error "Failed to download checksums.txt from release"
        return 1
    fi
    
    # Verify all platforms have checksums
    local platforms=("linux-amd64" "linux-arm64" "darwin-amd64" "darwin-arm64")
    
    for platform in "${platforms[@]}"; do
        local filename="kubectl-reboot-${platform}.tar.gz"
        
        if ! grep -q "$filename" checksums.txt; then
            error "Checksum for $filename not found in checksums.txt"
            return 1
        fi
    done
    
    success "All checksums validated"
}

setup_krew_index() {
    log "Setting up krew-index repository..."
    
    if [[ -d "krew-index" ]]; then
        log "krew-index directory exists, updating..."
        cd krew-index
        git fetch origin
        git checkout master
        git reset --hard origin/master
        cd ..
    else
        log "Cloning krew-index repository..."
        git clone "$KREW_INDEX_REPO" krew-index
    fi
    
    success "krew-index repository ready"
}

create_submission_branch() {
    local version=$1
    
    cd krew-index
    
    local branch_name="add-kubectl-reboot-$version"
    
    log "Creating submission branch: $branch_name"
    
    # Check if branch already exists
    if git branch -r | grep -q "origin/$branch_name"; then
        warning "Branch $branch_name already exists on remote"
        read -p "Continue with existing branch? [y/N] " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
        git checkout -B "$branch_name" "origin/$branch_name"
    else
        git checkout -b "$branch_name"
    fi
    
    cd ..
    success "Submission branch created"
}

update_manifest() {
    local version=$1
    
    log "Updating plugin manifest..."
    
    # Copy the updated manifest
    cp kubectl-reboot.yaml.updated "krew-index/plugins/$PLUGIN_NAME.yaml"
    
    # Ensure the version is correct in the manifest
    sed -i.bak "s/version: v[0-9]\+\.[0-9]\+\.[0-9]\+/version: $version/" "krew-index/plugins/$PLUGIN_NAME.yaml"
    rm -f "krew-index/plugins/$PLUGIN_NAME.yaml.bak"
    
    success "Plugin manifest updated"
}

validate_plugin() {
    if [[ "$SKIP_VALIDATION" == "true" ]]; then
        warning "Skipping plugin validation (--skip-validation)"
        return
    fi
    
    log "Validating plugin manifest..."
    
    # Test with linux-amd64 binary
    local test_url="https://github.com/ayetkin/kubectl-reboot/releases/download/${VERSION}/kubectl-reboot-linux-amd64.tar.gz"
    
    # Create a temporary directory for validation
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # Download test binary
    if ! curl -fsSL "$test_url" -o "$temp_dir/kubectl-reboot-linux-amd64.tar.gz"; then
        error "Failed to download test binary for validation"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Validate with krew
    if kubectl krew install --manifest="krew-index/plugins/$PLUGIN_NAME.yaml" --archive="$temp_dir/kubectl-reboot-linux-amd64.tar.gz"; then
        success "Plugin validation successful"
        # Uninstall after validation
        kubectl krew uninstall reboot 2>/dev/null || true
    else
        error "Plugin validation failed"
        rm -rf "$temp_dir"
        return 1
    fi
    
    rm -rf "$temp_dir"
}

create_commit() {
    local version=$1
    
    cd krew-index
    
    log "Creating commit..."
    
    # Check if there are changes to commit
    if git diff --staged --quiet && git diff --quiet; then
        if git ls-files --others --exclude-standard | grep -q .; then
            git add "plugins/$PLUGIN_NAME.yaml"
        else
            warning "No changes to commit"
            cd ..
            return
        fi
    else
        git add "plugins/$PLUGIN_NAME.yaml"
    fi
    
    # Create commit message
    local commit_msg="Add kubectl-reboot $version

kubectl-reboot is a kubectl plugin that safely reboots Kubernetes nodes
by draining pods, rebooting via SSH, verifying the reboot, and uncordoning
the nodes.

Key Features:
- 🔄 Safe node restart workflow (cordon → drain → reboot → uncordon)
- 🚀 SSH integration with customizable reboot commands
- 🔍 Reboot verification via Boot ID monitoring
- 🌐 Cluster-wide operations with node filtering
- 🧪 Dry-run mode for safe planning
- ⚡ Extensive configuration options

Use cases:
- Rolling node updates and maintenance
- Kernel updates requiring reboots
- Hardware troubleshooting
- Cluster maintenance automation

Source: $REPO_URL
Release: $REPO_URL/releases/tag/$version"
    
    git commit -m "$commit_msg"
    
    cd ..
    success "Commit created"
}

cleanup() {
    rm -f checksums.txt kubectl-reboot.yaml.updated
}

main() {
    local VERSION=""
    local DRY_RUN="false"
    local SKIP_VALIDATION="false"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            --skip-validation)
                SKIP_VALIDATION="true"
                shift
                ;;
            v*.*.*)
                VERSION=$1
                shift
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$VERSION" ]]; then
        error "Version is required"
        usage
        exit 1
    fi
    
    # Export variables for use in functions
    export DRY_RUN SKIP_VALIDATION VERSION
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    log "Starting krew-index submission for kubectl-reboot $VERSION"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        warning "DRY RUN MODE - No changes will be made"
    fi
    
    check_prerequisites
    download_updated_manifest "$VERSION"
    validate_checksums "$VERSION"
    setup_krew_index
    create_submission_branch "$VERSION"
    update_manifest "$VERSION"
    validate_plugin
    
    if [[ "$DRY_RUN" == "false" ]]; then
        create_commit "$VERSION"
    fi
    
    success "Submission prepared!"
    
    echo ""
    log "Next steps:"
    if [[ "$DRY_RUN" == "false" ]]; then
        echo "  1. Review the changes:"
        echo "     cd krew-index && git show"
        echo ""
        echo "  2. Push the branch:"
        echo "     cd krew-index && git push origin add-kubectl-reboot-$VERSION"
        echo ""
        echo "  3. Open a Pull Request:"
        echo "     https://github.com/kubernetes-sigs/krew-index/compare/master...add-kubectl-reboot-$VERSION"
        echo ""
        echo "  4. Wait for maintainer review and approval"
    else
        echo "  Run without --dry-run to create the actual submission"
    fi
    echo ""
    echo "📁 Plugin manifest location: krew-index/plugins/$PLUGIN_NAME.yaml"
    echo "🔗 Plugin source: $REPO_URL"
}

main "$@"
