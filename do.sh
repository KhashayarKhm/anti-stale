#!/bin/bash

set -e
set -u

TIMESTAMP=$(date +%Y-%m-%d)

### Helper function to display usage
help() {
    echo "Usage: $0 {help|build|release} [vX.Y.Z]"
    echo "Commands:"
    echo "  help            Shows this message"
    echo "  build [-a]      Builds the project"
    echo "                  -a: Build for all platforms (cross-platform)"
    echo "  release         Releases a specific version (e.g., v0.2.0)"
    exit 0
}

### Build the project
build() {
    if [[ "${1:-}" == "-a" ]]; then
        build_cross_platform
        return
    fi
    echo "Building for current platform..."
    local version=$(git describe --tags 2> /dev/null || echo "v0.0.0")
    go build -o ./bin/anti-stale -ldflags "-X github.com/KhashayarKhm/anti-stale/cmd.version=$version" ./main.go
    echo "Build completed: ./bin/anti-stale"
}

### Build the project for all platforms and create checksums
build_cross_platform() {
    local version=$(git describe --tags 2> /dev/null || echo "v0.0.0")
    local ldflags="-X github.com/KhashayarKhm/anti-stale/cmd.version=$version"

    # Create build directory
    mkdir -p ./bin

    echo "Building for multiple platforms..."

    local file_path="./bin"

    # Linux AMD64
    echo "Building for Linux AMD64..."
    local linux_amd64_file_name="anti-stale-$version-linux-amd64"
    local linux_amd64_file_path="$file_path/$linux_amd64_file_name"
    GOOS=linux GOARCH=amd64 go build -ldflags "$ldflags" -o $linux_amd64_file_path ./main.go
    sha256sum $linux_amd64_file_path > "$file_path/$linux_amd64_file_name-checksum.txt"

    # Linux ARM64
    echo "Building for Linux ARM64..."
    local linux_arm64_file_name="anti-stale-$version-linux-arm64"
    local linux_arm64_file_path="$file_path/$linux_arm64_file_name"
    GOOS=linux GOARCH=arm64 go build -ldflags "$ldflags" -o $linux_arm64_file_path ./main.go
    sha256sum $linux_arm64_file_path > "$file_path/$linux_arm64_file_name-checksum.txt"

    # macOS ARM64 (Apple Silicon)
    echo "Building for macOS ARM64..."
    local darwin_arm64_file_name="anti-stale-$version-darwin-arm64"
    local darwin_arm64_file_path="$file_path/$darwin_arm64_file_name"
    GOOS=darwin GOARCH=arm64 go build -ldflags "$ldflags" -o $darwin_arm64_file_path ./main.go
    sha256sum $darwin_arm64_file_path > "$file_path/$darwin_arm64_file_name-checksum.txt"

    # Windows AMD64
    echo "Building for Windows AMD64..."
    local windows_amd64_file_name="anti-stale-$version-windows-amd64.exe"
    local windows_amd64_file_path="$file_path/$windows_amd64_file_name"
    GOOS=windows GOARCH=amd64 go build -ldflags "$ldflags" -o $windows_amd64_file_path ./main.go
    sha256sum $windows_amd64_file_path > "$file_path/$windows_amd64_file_name-checksum.txt"

    echo "Cross-platform builds completed in ./bin/"
    ls -la ./bin/
}

### Release a specific version
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
    branch_name=$(git branch --show-current)
    if [ $branch_name == "master" ]; then
        if grep -q "## \[Unreleased\]" CHANGELOG.md; then
            sed -i.bak "s/## \[Unreleased\]/## [$next_version] - $TIMESTAMP/" CHANGELOG.md
            read -p "Do you want to edit/preview the CHANGELOG.md before committing? (Y/n): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                ${EDITOR:-nano} CHANGELOG.md
            fi
        else
            echo "Error: no \"Unreleased\" section in the CHANGELOG.md"
        fi

        git add CHANGELOG.md
        git commit -m "chore(release): prepare $next_version"
        git tag "$next_version"
        git push origin "$next_version"

        read -p "Push to remote now? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            git push origin master
            git push origin "$next_version"
            echo "Pushed to remote"
        fi
    else
        echo "Error: you can only use release command on master branch"
    fi
}

### Main script
case "${1:-help}" in
    help)
        help
        ;;
    build)
        build "${2:-}"
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
