#!/usr/bin/env fish
# GZH Manager v2.0 Backward Compatibility Aliases
# These aliases will be removed in v3.0 (estimated: 2025-01-01)

function _gz_deprecation_warning
    set -l old_cmd $argv[1]
    set -l new_cmd $argv[2]
    echo (set_color yellow)"Warning:"(set_color normal) \
         (set_color red)"gz $old_cmd"(set_color normal) "is deprecated." >&2
    echo "Please use" (set_color yellow)"gz $new_cmd"(set_color normal) "instead." >&2
    echo "This alias will be removed in v3.0. Run 'gz migrate help' for details." >&2
    echo "" >&2
end

function gz
    switch $argv[1]
        case "gen-config"
            _gz_deprecation_warning "gen-config" "synclone config generate"
            command gz synclone config generate $argv[2..-1]
        
        case "repo-config"
            # Still available, no deprecation needed
            command gz repo-config $argv[2..-1]
        
        case "event"
            # Still available, no deprecation needed
            command gz event $argv[2..-1]
        
        case "webhook"
            # Still available, no deprecation needed
            command gz webhook $argv[2..-1]
        
        case "ssh-config"
            _gz_deprecation_warning "ssh-config" "dev-env ssh"
            command gz dev-env ssh $argv[2..-1]
        
        case "doctor"
            # Still available, no deprecation needed
            command gz doctor $argv[2..-1]
        
        case "shell"
            echo (set_color red)"Error:"(set_color normal) "'gz shell' has been removed." >&2
            echo "For debugging, use:" (set_color yellow)"gz --debug-shell"(set_color normal) >&2
            echo "or set:" (set_color yellow)"export GZH_DEBUG_SHELL=1"(set_color normal) >&2
            return 1
        
        case "config"
            echo (set_color red)"Error:"(set_color normal) "'gz config' has been removed." >&2
            echo "Use:" (set_color yellow)"gz [command] config"(set_color normal) "instead. For example:" >&2
            echo "  - gz synclone config" >&2
            echo "  - gz dev-env config" >&2
            echo "  - gz net-env config" >&2
            return 1
        
        case "docker"
            echo (set_color red)"Error:"(set_color normal) "'gz docker' has been removed." >&2
            echo "Please use Docker CLI directly for container management." >&2
            return 1
        
        case "always-latest"
            _gz_deprecation_warning "always-latest" "pm"
            command gz pm $argv[2..-1]
        
        case "migrate"
            echo (set_color red)"Error:"(set_color normal) "'gz migrate' has been removed." >&2
            echo "The migration tool was for v1.x to v2.0 transition only." >&2
            echo "For migration help, see: docs/migration/command-migration-guide.md" >&2
            return 1
        
        case "*"
            command gz $argv
    end
end