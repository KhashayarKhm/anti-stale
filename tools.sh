#!/bin/bash

set -e
set -u

# Helper function to display usage
help() {
    echo "Usage: $0 {help|build|release} [vX.Y.Z]"
    echo "Commands:"
    echo "  help        Shows this message"
    echo "  build       Builds the project"
    echo "  release     Releases a specific version (e.g., v0.2.0)"
    exit 0
}

# Build the project
build() {
    local version=`git describe --tags`
    go build main.go -o ./bin/anti-stale -ldflags "-X cmd.cmd.version=$version"
}

# Release a specific version
release() {
    local next_version=$1
    if [ -z "$next_version" ]; then
        echo "Error: version is required (e.g., $0 release v0.2.0)"
        exit 1
    fi
    if ! echo "$next_version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
        echo "Error: version must be in SemVer format (e.g., v0.2.0)"
        exit 1
    fi
    echo "Releasing version $next_version"
    branch_name=`git branch --show-current`
    if [ $branch_name == "master" ]; then
        echo "do the rest process"
        echo "$next_version" > ./VERSION
        git tag "$next_version"
        git push origin "$next_version"
    else
        echo "Error: you can only use release command on master branch"
    fi
}

# Main script
case "${1:-help}" in
    help)
        help
        ;;
    build)
        build
        ;;
    release)
        release "${2:-}"
        ;;
    *)
        echo "Unknown command: $1"
        help
        exit 1
        ;;
esac
