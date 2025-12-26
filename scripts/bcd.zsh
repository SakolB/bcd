# bcd shell integration for Zsh
# Source this file in your ~/.zshrc or add via install script

bcd() {
  local selected_path

  selected_path="$(
    CLICOLOR_FORCE=1 command bcd-bin "$@" 2>&1 1>/dev/tty \
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
