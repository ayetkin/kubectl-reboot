#!/bin/bash

# Extract changelog for a specific version from CHANGELOG.md
# Usage: ./extract-changelog.sh <version>

set -euo pipefail

VERSION=${1:-}

if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Remove 'v' prefix if present for matching in changelog
CLEAN_VERSION=${VERSION#v}

# Check if CHANGELOG.md exists
if [[ ! -f "CHANGELOG.md" ]]; then
    echo "Initial release"
    exit 0
fi

# Extract content between version headers
# This will match ## [1.0.0] or ## [1.0.0] - 2024-09-24
START_PATTERN="^## \[${CLEAN_VERSION}\]"
END_PATTERN="^## \["

# Find the start line number
START_LINE=$(grep -n "$START_PATTERN" CHANGELOG.md | head -1 | cut -d: -f1 || echo "")

if [[ -z "$START_LINE" ]]; then
    echo "Version $VERSION not found in CHANGELOG.md"
    echo "Initial release"
    exit 0
fi

# Find the next version section (end line)
END_LINE=$(tail -n +$((START_LINE + 1)) CHANGELOG.md | grep -n "$END_PATTERN" | head -1 | cut -d: -f1 || echo "")

if [[ -n "$END_LINE" ]]; then
    # Calculate actual end line number
    END_LINE=$((START_LINE + END_LINE))
    # Extract content between start and end, excluding the headers
    sed -n "$((START_LINE + 1)),$((END_LINE - 1))p" CHANGELOG.md
else
    # No next version found, extract until end of file
    sed -n "$((START_LINE + 1)),\$p" CHANGELOG.md
fi | sed '/^$/d' | sed '/^\[/d' # Remove empty lines and reference links