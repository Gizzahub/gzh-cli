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
            _gz_deprecation_warning "repo-config" "repo-sync config"
            command gz repo-sync config $argv[2..-1]
        
        case "event"
            _gz_deprecation_warning "event" "repo-sync event"
            command gz repo-sync event $argv[2..-1]
        
        case "webhook"
            _gz_deprecation_warning "webhook" "repo-sync webhook"
            command gz repo-sync webhook $argv[2..-1]
        
        case "ssh-config"
            _gz_deprecation_warning "ssh-config" "dev-env ssh"
            command gz dev-env ssh $argv[2..-1]
        
        case "doctor"
            _gz_deprecation_warning "doctor" "validate --all"
            command gz validate --all $argv[2..-1]
        
        case "shell"
            echo (set_color red)"Error:"(set_color normal) "'gz shell' has been removed." >&2
            echo "For debugging, use:" (set_color yellow)"gz --debug-shell"(set_color normal) >&2
            echo "or set:" (set_color yellow)"export GZH_DEBUG_SHELL=1"(set_color normal) >&2
            return 1
        
        case "config"
            if test (count $argv) -gt 1
                echo (set_color red)"Error:"(set_color normal) \
                     "'gz config' has been distributed to individual commands." >&2
                echo "Use:" (set_color yellow)"gz [command] config"(set_color normal) "instead. For example:" >&2
                echo "  - gz synclone config" >&2
                echo "  - gz dev-env config" >&2
                echo "  - gz net-env config" >&2
                echo "  - gz repo-sync config" >&2
                return 1
            end
        
        case "migrate"
            if test "$argv[2]" = "help" -o (count $argv) -eq 1
                echo "GZ Migration Guide - Command Changes in v2.0"
                echo "==========================================="
                echo ""
                echo "The following commands have been renamed or restructured:"
                echo ""
                echo "  Old Command          →  New Command"
                echo "  -----------             -----------"
                echo "  gz gen-config        →  gz synclone config generate"
                echo "  gz repo-config       →  gz repo-sync config"
                echo "  gz event             →  gz repo-sync event"
                echo "  gz webhook           →  gz repo-sync webhook"
                echo "  gz ssh-config        →  gz dev-env ssh"
                echo "  gz doctor            →  gz validate --all"
                echo "  gz shell             →  gz --debug-shell (hidden)"
                echo "  gz config            →  gz [command] config"
                echo ""
                echo "To install backward compatibility aliases:"
                echo "  source ~/.config/gzh-manager/aliases.fish"
                echo ""
                echo "For more details, see: docs/migration/command-migration-guide.md"
                return 0
            else
                command gz $argv
            end
        
        case "*"
            command gz $argv
    end
end