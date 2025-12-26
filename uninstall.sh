#!/usr/bin/env bash
set -e

# bcd uninstaller script
# This script removes bcd binary and shell integration

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

# Remove binary
remove_binary() {
    local binary_removed=false

    # Check ~/.local/bin
    if [[ -f "$HOME/.local/bin/bcd-bin" ]]; then
        info "Removing bcd-bin from ~/.local/bin"
        rm -f "$HOME/.local/bin/bcd-bin"
        binary_removed=true
    fi

    # Check /usr/local/bin
    if [[ -f "/usr/local/bin/bcd-bin" ]]; then
        info "Removing bcd-bin from /usr/local/bin (requires sudo)"
        if [[ -w "/usr/local/bin" ]]; then
            rm -f "/usr/local/bin/bcd-bin"
        else
            sudo rm -f "/usr/local/bin/bcd-bin"
        fi
        binary_removed=true
    fi

    if [[ "$binary_removed" = false ]]; then
        warn "bcd-bin binary not found in ~/.local/bin or /usr/local/bin"
    else
        info "Binary removed successfully"
    fi
}

# Remove bash integration
remove_bash_integration() {
    local rc_file="$HOME/.bashrc"
    local integration_marker="# bcd shell integration"

    if [[ ! -f "$rc_file" ]]; then
        return
    fi

    if grep -q "$integration_marker" "$rc_file" 2>/dev/null; then
        info "Removing Bash integration from $rc_file"

        # Create backup
        cp "$rc_file" "$rc_file.bcd_backup"

        # Remove the integration block (from marker to end of function)
        sed -i.tmp '/# bcd shell integration/,/^}$/d' "$rc_file"
        rm -f "$rc_file.tmp"

        info "Bash integration removed (backup saved as $rc_file.bcd_backup)"
        echo "  Run: source ~/.bashrc"
    fi
}

# Remove zsh integration
remove_zsh_integration() {
    local rc_file="$HOME/.zshrc"
    local integration_marker="# bcd shell integration"

    if [[ ! -f "$rc_file" ]]; then
        return
    fi

    if grep -q "$integration_marker" "$rc_file" 2>/dev/null; then
        info "Removing Zsh integration from $rc_file"

        # Create backup
        cp "$rc_file" "$rc_file.bcd_backup"

        # Remove the integration block (from marker to end of function)
        sed -i.tmp '/# bcd shell integration/,/^}$/d' "$rc_file"
        rm -f "$rc_file.tmp"

        info "Zsh integration removed (backup saved as $rc_file.bcd_backup)"
        echo "  Run: source ~/.zshrc"
    fi
}

# Remove fish integration
remove_fish_integration() {
    local func_file="$HOME/.config/fish/functions/bcd.fish"

    if [[ -f "$func_file" ]]; then
        info "Removing Fish integration from $func_file"

        # Create backup
        cp "$func_file" "$func_file.bcd_backup"
        rm -f "$func_file"

        info "Fish integration removed (backup saved as $func_file.bcd_backup)"
        echo "  Changes will take effect in new fish sessions"
    fi
}

# Remove shell integrations
remove_shell_integration() {
    info "Removing shell integrations..."

    remove_bash_integration
    remove_zsh_integration
    remove_fish_integration
}

# Main uninstallation flow
main() {
    echo -e "${BOLD}bcd Uninstaller${NC}"
    echo ""

    # Confirm uninstallation
    read -p "Are you sure you want to uninstall bcd? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Uninstallation cancelled."
        exit 0
    fi

    remove_binary
    remove_shell_integration

    echo ""
    info "Uninstallation complete!"
    echo ""
    echo "Note: Backup files have been created for all modified configs"
    echo "You may want to manually remove the PATH export for ~/.local/bin if you added it"
}

# Handle arguments
case "$1" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help       Show this help message"
        echo "  --force      Skip confirmation prompt"
        exit 0
        ;;
    --force)
        # Skip confirmation
        ;;
    *)
        if [[ -n "$1" ]]; then
            error "Unknown option: $1"
        fi
        ;;
esac

if [[ "$1" == "--force" ]]; then
    # Run without confirmation
    echo -e "${BOLD}bcd Uninstaller${NC}"
    echo ""
    remove_binary
    remove_shell_integration
    echo ""
    info "Uninstallation complete!"
    echo ""
    echo "Note: Backup files have been created for all modified configs"
    echo "You may want to manually remove the PATH export for ~/.local/bin if you added it"
else
    main
fi
