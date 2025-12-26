#!/usr/bin/env bash
set -e

# bcd installer script
# This script builds and installs bcd and sets up shell integration

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} ${BOLD}$1${NC}"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1" >&2
    exit 1
}

# Detect shell
detect_shell() {
    local shell_name
    shell_name=$(basename "$SHELL")

    case "$shell_name" in
        bash)
            echo "bash"
            ;;
        zsh)
            echo "zsh"
            ;;
        fish)
            echo "fish"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.21 or later from https://golang.org"
    fi

    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    info "Found Go version: $go_version"
}

# Build the binary
build() {
    info "Building bcd-bin..."
    if ! go build -o bcd-bin ./cmd/bcd; then
        error "Build failed"
    fi
    info "Build successful"
}

# Install binary
install_binary() {
    local install_dir="$HOME/.local/bin"

    # Check if user wants system-wide install
    if [[ "$1" == "--system" ]]; then
        install_dir="/usr/local/bin"
        if [[ ! -w "$install_dir" ]]; then
            info "Installing to $install_dir (requires sudo)"
            sudo mv bcd-bin "$install_dir/bcd-bin"
            sudo chmod +x "$install_dir/bcd-bin"
        else
            mv bcd-bin "$install_dir/bcd-bin"
            chmod +x "$install_dir/bcd-bin"
        fi
    else
        mkdir -p "$install_dir"
        mv bcd-bin "$install_dir/bcd-bin"
        chmod +x "$install_dir/bcd-bin"
    fi

    info "Installed bcd-bin to $install_dir"
}

# Setup shell integration
setup_shell_integration() {
    local shell_type
    shell_type=$(detect_shell)

    case "$shell_type" in
        bash)
            setup_bash
            ;;
        zsh)
            setup_zsh
            ;;
        fish)
            setup_fish
            ;;
        *)
            warn "Could not detect shell type. Please manually add shell integration."
            echo "  See README.md for instructions."
            return
            ;;
    esac
}

setup_bash() {
    local rc_file="$HOME/.bashrc"
    local integration_marker="# bcd shell integration"
    local path_export='export PATH="$HOME/.local/bin:$PATH"'
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local integration_file="$script_dir/scripts/bcd.bash"
    local needs_source=false

    # Check if bcd function already exists
    if grep -q "$integration_marker" "$rc_file" 2>/dev/null; then
        warn "bcd shell integration already exists in $rc_file"
        echo "  Skipping function installation."
    else
        info "Adding Bash integration to $rc_file"
        if [[ ! -f "$integration_file" ]]; then
            error "Integration file not found: $integration_file"
        fi
        echo "" >> "$rc_file"
        cat "$integration_file" >> "$rc_file"
        needs_source=true
    fi

    # Check if PATH export is needed
    if grep -q 'PATH.*\.local/bin' "$rc_file" 2>/dev/null; then
        info "~/.local/bin is already in PATH in $rc_file"
    else
        info "Adding ~/.local/bin to PATH in $rc_file"
        echo "$path_export" >> "$rc_file"
        needs_source=true
    fi

    if [ "$needs_source" = true ]; then
        echo ""
        info "Bash integration added!"
        echo "  Run: source ~/.bashrc"
    fi
}

setup_zsh() {
    local rc_file="$HOME/.zshrc"
    local integration_marker="# bcd shell integration"
    local path_export='export PATH="$HOME/.local/bin:$PATH"'
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local integration_file="$script_dir/scripts/bcd.zsh"
    local needs_source=false

    # Check if bcd function already exists
    if grep -q "$integration_marker" "$rc_file" 2>/dev/null; then
        warn "bcd shell integration already exists in $rc_file"
        echo "  Skipping function installation."
    else
        info "Adding Zsh integration to $rc_file"
        if [[ ! -f "$integration_file" ]]; then
            error "Integration file not found: $integration_file"
        fi
        echo "" >> "$rc_file"
        cat "$integration_file" >> "$rc_file"
        needs_source=true
    fi

    # Check if PATH export is needed
    if grep -q 'PATH.*\.local/bin' "$rc_file" 2>/dev/null; then
        info "~/.local/bin is already in PATH in $rc_file"
    else
        info "Adding ~/.local/bin to PATH in $rc_file"
        echo "$path_export" >> "$rc_file"
        needs_source=true
    fi

    if [ "$needs_source" = true ]; then
        echo ""
        info "Zsh integration added!"
        echo "  Run: source ~/.zshrc"
    fi
}

setup_fish() {
    local func_dir="$HOME/.config/fish/functions"
    local func_file="$func_dir/bcd.fish"
    local config_file="$HOME/.config/fish/config.fish"
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local integration_file="$script_dir/scripts/bcd.fish"
    local path_added=false
    local func_added=false

    mkdir -p "$func_dir"
    mkdir -p "$HOME/.config/fish"

    # Check if bcd function already exists
    if [[ -f "$func_file" ]]; then
        warn "bcd shell integration already exists at $func_file"
        echo "  Skipping function installation."
    else
        info "Adding Fish integration to $func_file"
        if [[ ! -f "$integration_file" ]]; then
            error "Integration file not found: $integration_file"
        fi
        cat "$integration_file" > "$func_file"
        func_added=true
    fi

    # Check if PATH export is needed
    if grep -q 'fish_add_path.*\.local/bin' "$config_file" 2>/dev/null; then
        info "~/.local/bin is already in PATH in Fish config"
    else
        info "Adding ~/.local/bin to PATH in Fish config"
        echo 'fish_add_path $HOME/.local/bin' >> "$config_file"
        path_added=true
    fi

    if [ "$func_added" = true ] || [ "$path_added" = true ]; then
        echo ""
        info "Fish integration added!"
        echo "  Changes will take effect in new fish sessions."
    fi
}

# Main installation flow
main() {
    echo -e "${BOLD}bcd Installer${NC}"
    echo ""

    check_go
    build

    if [[ "$1" == "--system" ]]; then
        install_binary --system
    else
        install_binary
    fi

    setup_shell_integration

    echo ""
    info "Installation complete!"
    echo ""
    echo "Usage:"
    echo "  bcd          # Search from current directory"
    echo "  bcd /path    # Search from specific directory"
    echo ""
    echo "Keyboard shortcuts:"
    echo "  ↑/↓          Navigate results"
    echo "  Enter        Select and cd"
    echo "  Esc          Cancel"
}

# Handle arguments
case "$1" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --system     Install to /usr/local/bin (requires sudo)"
        echo "  --help       Show this help message"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
