#!/bin/bash

# Usage: ./bump-version.sh [major|minor|patch] [os_type]
# Example: ./bump-version.sh patch rhel9

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
VERSION_TYPE=$1
OS_TYPE=$2

if [ -z "$VERSION_TYPE" ] || [ -z "$OS_TYPE" ]; then
    echo "Usage: $0 [major|minor|patch] [os_type]"
    echo "Example: $0 patch rhel9"
    exit 1
fi

VERSION_FILE="$SCRIPT_DIR/../$OS_TYPE/version.txt"
if [ ! -f "$VERSION_FILE" ]; then
    echo "Version file not found: $VERSION_FILE"
    exit 1
fi

# Read current version
CURRENT_VERSION=$(cat "$VERSION_FILE")
MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1)
MINOR=$(echo $CURRENT_VERSION | cut -d. -f2)
PATCH=$(echo $CURRENT_VERSION | cut -d. -f3)

# Bump version according to type
case $VERSION_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Invalid version type. Use major, minor, or patch"
        exit 1
        ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "Bumping $OS_TYPE version: $CURRENT_VERSION -> $NEW_VERSION"
echo "$NEW_VERSION" > "$VERSION_FILE"
