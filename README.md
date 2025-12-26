# bcd - Better CD

A fuzzy directory navigator for your terminal. Search and jump to any directory with an interactive TUI.

## Features

- **BFS Directory Discovery**: Finds directories closest to your current location first
- **FZF v2 Fuzzy Matching**: Intelligent scoring algorithm for accurate search results
- **Real-time Updates**: Results appear as directories are discovered
- **Smooth Performance**: Async architecture with batching and heap-based ranking
- **Interactive TUI**: Full-screen terminal interface with real-time fuzzy search
- **Shell Integration**: Seamlessly cd into selected directories

## Built With

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI framework for building interactive applications
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components for Bubble Tea (text input, viewports, etc.)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for terminal rendering
- **FZF v2 Algorithm** - Fuzzy matching scoring algorithm

## Installation

### Requirements

- **Go 1.21 or later** (for building from source)
- Unix-like system (Linux, macOS)

Dependencies like Bubble Tea are automatically downloaded during build - you don't need to install them manually.

### Quick Install

```bash
./install.sh
```

This will:
1. Check for Go installation
2. Build the `bcd` binary (automatically downloads dependencies)
3. Install it to `~/.local/bin/` (or use `--system` flag for `/usr/local/bin`)
4. Add `~/.local/bin` to your PATH if needed
5. Add shell integration function to your shell config
6. Prompt you to reload your shell

After installation, run:
```bash
source ~/.bashrc  # or source ~/.zshrc for Zsh
```

### System-wide Installation

```bash
./install.sh --system
```

Installs to `/usr/local/bin` (requires sudo).

### Uninstallation

```bash
./uninstall.sh
```

This will:
1. Remove the `bcd` binary from `~/.local/bin` or `/usr/local/bin`
2. Remove shell integration from your shell config files
3. Create backup files of all modified configs (`.bcd_backup`)

Use `--force` to skip the confirmation prompt:
```bash
./uninstall.sh --force
```

### Manual Installation

1. Build the binary:
```bash
go build -o bcd ./cmd/bcd
```

2. Move it somewhere in your PATH:
```bash
mkdir -p ~/.local/bin
mv bcd ~/.local/bin/
```

3. Add `~/.local/bin` to your PATH (if not already):
```bash
export PATH="$HOME/.local/bin:$PATH"
```

4. Add the shell function to your shell config:

**Bash** (`~/.bashrc`):
```bash
# bcd shell integration
bcd() {
  local selected_path

  selected_path="$(
    command bcd "$@" 1>/dev/tty 2>&1 \
      | tr -d '\r' \
      | sed -E 's/\x1b\[[0-9;?]*[ -/]*[@-~]//g' \
      | grep -oE 'BCD_SELECTED_PATH:[^[:cntrl:]]+' \
      | tail -n 1 \
      | sed 's/^BCD_SELECTED_PATH://'
  )"

  if [[ -n "$selected_path" ]]; then
    if [[ -d "$selected_path" ]]; then
      builtin cd -- "$selected_path" || return 1
    elif [[ -f "$selected_path" ]]; then
      builtin cd -- "$(dirname -- "$selected_path")" || return 1
    fi
  fi
}
```

**Zsh** (`~/.zshrc`):
```zsh
# bcd shell integration
bcd() {
  local selected_path

  selected_path="$(
    command bcd "$@" 1>/dev/tty 2>&1 \
      | tr -d '\r' \
      | sed -E 's/\x1b\[[0-9;?]*[ -/]*[@-~]//g' \
      | grep -oE 'BCD_SELECTED_PATH:[^[:cntrl:]]+' \
      | tail -n 1 \
      | sed 's/^BCD_SELECTED_PATH://'
  )"

  if [[ -n "$selected_path" ]]; then
    if [[ -d "$selected_path" ]]; then
      builtin cd -- "$selected_path" || return 1
    elif [[ -f "$selected_path" ]]; then
      builtin cd -- "$(dirname -- "$selected_path")" || return 1
    fi
  fi
}
```

**Fish** (`~/.config/fish/functions/bcd.fish`):
```fish
# bcd shell integration
function bcd
    set selected_path (
        command bcd $argv 1>/dev/tty 2>&1 \
            | tr -d '\r' \
            | sed -E 's/\x1b\[[0-9;?]*[ -/]*[@-~]//g' \
            | grep -oE 'BCD_SELECTED_PATH:[^[:cntrl:]]+' \
            | tail -n 1 \
            | sed 's/^BCD_SELECTED_PATH://'
    )

    if test -n "$selected_path"
        if test -d "$selected_path"
            builtin cd -- $selected_path
        else if test -f "$selected_path"
            builtin cd -- (dirname -- $selected_path)
        end
    end
end
```

5. Reload your shell:
```bash
source ~/.bashrc  # or source ~/.zshrc for Zsh
```

## Usage

```bash
# Search from current directory
bcd

# Search from a specific directory
bcd /path/to/start
```

### Keyboard Shortcuts

- `↑/↓` or `Ctrl+p/n`: Navigate results
- `Enter`: Select directory and cd into it
- `Esc` or `Ctrl+c`: Cancel
- Type to search: Fuzzy match against directory names

## How it Works

### Directory Discovery and Ranking

1. **BFS Traversal**: Discovers directories using breadth-first search, prioritizing closer paths
2. **Distance Calculation**: Ranks results by path distance from starting location
3. **FZF v2 Scoring**: Uses dynamic programming for optimal fuzzy matching
4. **Async Processing**: Background worker processes entries without blocking the UI
5. **Batching**: Groups directory discoveries (100 entries or 50ms intervals) for efficient processing
6. **Heap-Based Ranking**: Maintains top results using a max-heap for O(log k) insertion

### Shell Integration

The shell function wrapper enables `cd` functionality through a clever I/O redirection:

1. **TUI renders to terminal**: `1>/dev/tty` redirects stdout to the terminal so you can see and interact with the TUI
2. **Selection captured via stderr**: The binary outputs `BCD_SELECTED_PATH:/path/to/dir` to stderr
3. **Shell extracts path**: `2>&1` redirects stderr to the pipe, where `sed` extracts the path
4. **cd into directory**: The shell function uses `builtin cd` to change directories

This approach allows the interactive TUI to display normally while the selected path is captured for shell use.

## Project Structure

```
bcd/
├── cmd/bcd/           # Main application entry point
├── internal/          # Internal packages
│   ├── crawler/       # BFS directory traversal
│   ├── entry/         # Path entry data structures
│   ├── ranker/        # FZF v2 scoring and ranking
│   └── tui/           # Bubble Tea TUI interface
├── scripts/           # Shell integration scripts
│   ├── bcd.bash       # Bash integration
│   ├── bcd.zsh        # Zsh integration
│   └── bcd.fish       # Fish integration
├── install.sh         # Installation script
├── uninstall.sh       # Uninstallation script
├── LICENSE            # MIT License with FZF attribution
└── README.md
```

## Troubleshooting

### `bcd: command not found`

Make sure `~/.local/bin` is in your PATH and you've reloaded your shell:

```bash
# Add to your shell config if not present
export PATH="$HOME/.local/bin:$PATH"

# Reload your shell
source ~/.bashrc  # or source ~/.zshrc
```

### Shell function not working

If the `bcd` command doesn't cd you to the selected directory:

1. Verify the shell function is defined:
   ```bash
   type bcd  # Should show it's a shell function
   ```

2. Make sure you've reloaded your shell config after installation

3. For reinstallation, the install script will warn if integration already exists. Use the uninstall script first:
   ```bash
   ./uninstall.sh
   ./install.sh
   ```

### TUI not displaying correctly

The TUI requires a terminal that supports ANSI escape codes. Most modern terminals (Terminal.app, iTerm2, Alacritty, etc.) work fine.

## Development

### Building from Source

```bash
go build -o bcd ./cmd/bcd
```

### Running Tests

```bash
go test ./internal/...
```

### Architecture

- **cmd/bcd**: Entry point, handles TUI initialization and output
- **internal/crawler**: BFS directory discovery with concurrent traversal
- **internal/entry**: Path entry data structures with distance calculation
- **internal/ranker**: FZF v2 fuzzy matching with heap-based ranking
- **internal/tui**: Bubble Tea TUI with async message handling

## License

MIT License - See LICENSE file for details

This project uses the FZF v2 algorithm for fuzzy matching. See LICENSE for attribution.
