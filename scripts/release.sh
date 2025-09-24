#!/bin/bash

# kubectl-reboot Release Script
# This script automates the release process for kubectl-reboot

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="ayetkin/kubectl-reboot"
PLUGIN_NAME="kubectl-reboot"

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
  version           Version to release (e.g., v1.0.1)

Options:
  -h, --help        Show this help message
  -d, --dry-run     Show what would be done without executing
  --skip-tests      Skip running tests
  --skip-build      Skip building binaries
  --draft           Create draft release

Examples:
  $0 v1.0.1                    # Create release v1.0.1
  $0 v1.0.1 --dry-run          # Preview release v1.0.1
  $0 v1.0.1 --draft            # Create draft release

EOF
}

check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]] || [[ ! -f "kubectl-reboot.yaml" ]]; then
        error "This script must be run from the root of the kubectl-reboot repository"
        exit 1
    fi
    
    # Check required tools
    local required_tools=("git" "go" "gh" "make")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "Required tool '$tool' is not installed"
            exit 1
        fi
    done
    
    # Check if we're on main branch
    local current_branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$current_branch" != "main" ]]; then
        warning "You're not on the main branch (current: $current_branch)"
        read -p "Continue anyway? [y/N] " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Check for uncommitted changes
    if [[ -n $(git status --porcelain) ]]; then
        error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
    
    # Check if we're up to date
    git fetch origin
    local local_commit remote_commit
    local_commit=$(git rev-parse HEAD)
    remote_commit=$(git rev-parse origin/main)
    if [[ "$local_commit" != "$remote_commit" ]]; then
        error "Local branch is not up to date with origin/main"
        exit 1
    fi
    
    success "All prerequisites met"
}

validate_version() {
    local version=$1
    
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Version must be in format vX.Y.Z (e.g., v1.0.1)"
        exit 1
    fi
    
    # Check if tag already exists
    if git rev-parse "$version" >/dev/null 2>&1; then
        error "Tag $version already exists"
        exit 1
    fi
    
    log "Version $version is valid"
}

run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        warning "Skipping tests (--skip-tests)"
        return
    fi
    
    log "Running tests..."
    make test
    make vet
    success "Tests passed"
}

build_release() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        warning "Skipping build (--skip-build)"
        return
    fi
    
    log "Building release binaries..."
    make clean
    make package
    success "Release binaries built successfully"
}

update_krew_manifest() {
    local version=$1
    
    log "Updating Krew manifest..."
    
    # Create updated manifest
    cp kubectl-reboot.yaml kubectl-reboot.yaml.bak
    
    # Update version
    sed -i.tmp "s/version: v[0-9]\+\.[0-9]\+\.[0-9]\+/version: $version/" kubectl-reboot.yaml
    
    # Update download URLs
    sed -i.tmp "s|/releases/download/v[0-9]\+\.[0-9]\+\.[0-9]\+/|/releases/download/$version/|g" kubectl-reboot.yaml
    
    # Update checksums (will be filled by GitHub Actions)
    sed -i.tmp 's/sha256: "[^"]*"/sha256: "TO_BE_FILLED_BY_RELEASE_PIPELINE"/g' kubectl-reboot.yaml
    
    # Clean up temp files
    rm -f kubectl-reboot.yaml.tmp
    
    success "Krew manifest updated"
}

create_release() {
    local version=$1
    
    log "Creating git tag..."
    git tag -a "$version" -m "Release $version"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        warning "DRY RUN: Would push tag $version"
        warning "DRY RUN: Would trigger release workflow"
        return
    fi
    
    log "Pushing tag to trigger release..."
    git push origin "$version"
    
    success "Tag $version pushed. GitHub Actions will handle the release."
    
    # Wait a moment and check release status
    log "Waiting for GitHub Actions to start..."
    sleep 5
    
    log "You can monitor the release progress at:"
    echo "  https://github.com/$REPO/actions"
    echo "  https://github.com/$REPO/releases"
}

cleanup() {
    if [[ -f "kubectl-reboot.yaml.bak" ]]; then
        log "Restoring original kubectl-reboot.yaml..."
        mv kubectl-reboot.yaml.bak kubectl-reboot.yaml
    fi
}

main() {
    local version=""
    local DRY_RUN="false"
    local SKIP_TESTS="false"
    local SKIP_BUILD="false"
    local DRAFT="false"
    
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
            --skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            --skip-build)
                SKIP_BUILD="true"
                shift
                ;;
            --draft)
                DRAFT="true"
                shift
                ;;
            v*.*.*)
                version=$1
                shift
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$version" ]]; then
        error "Version is required"
        usage
        exit 1
    fi
    
    # Export variables for use in functions
    export DRY_RUN SKIP_TESTS SKIP_BUILD DRAFT
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    log "Starting release process for $version..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        warning "DRY RUN MODE - No changes will be made"
    fi
    
    validate_version "$version"
    check_prerequisites
    run_tests
    build_release
    update_krew_manifest "$version"
    create_release "$version"
    
    success "Release process completed!"
    
    if [[ "$DRY_RUN" == "false" ]]; then
        echo ""
        log "Next steps:"
        echo "  1. Monitor the GitHub Actions workflow"
        echo "  2. Once the release is created, download kubectl-reboot.yaml.updated"
        echo "  3. Submit the updated manifest to Krew index:"
        echo "     ./scripts/submit-to-krew.sh $version"
        echo ""
        echo "Release URL: https://github.com/$REPO/releases/tag/$version"
    fi
}

main "$@"