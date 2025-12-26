# bcd shell integration for Fish
# Save this to ~/.config/fish/functions/bcd.fish or add via install script

function bcd
    set selected_path (
        command bcd-bin $argv 1>/dev/tty 2>&1 \
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
