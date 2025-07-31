#!/usr/bin/env fish
# GZH Manager v2.0 Backward Compatibility Aliases
# These aliases will be removed in v3.0 (estimated: 2025-01-01)

function _gz_deprecation_warning
    set -l old_cmd $argv[1]
    set -l new_cmd $argv[2]
    echo (set_color yellow)"Warning:"(set_color normal) \
         (set_color red)"gz $old_cmd"(set_color normal) "is deprecated." >&2
    echo "Please use" (set_color yellow)"gz $new_cmd"(set_color normal) "instead." >&2
    echo "This alias will be removed in v3.0." >&2
    echo "" >&2
end

function gz
    switch $argv[1]
        case "shell"
            echo (set_color red)"Error:"(set_color normal) "'gz shell' has been removed." >&2
            echo "For debugging, use:" (set_color yellow)"gz --debug-shell"(set_color normal) >&2
            return 1
        case "config"
            echo (set_color red)"Error:"(set_color normal) "'gz config' has been removed." >&2
            echo "Use:" (set_color yellow)"gz [command] config"(set_color normal) "instead." >&2
            return 1
        case "docker"
            echo (set_color red)"Error:"(set_color normal) "'gz docker' has been removed." >&2
            echo "Please use Docker CLI directly." >&2
            return 1
        case "always-latest"
            _gz_deprecation_warning "always-latest" "pm"
            command gz pm $argv[2..-1]
        case "*"
            command gz $argv
    end
end