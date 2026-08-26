# bash completion for gz                                   -*- shell-script -*-

__gz_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__gz_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__gz_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__gz_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__gz_handle_go_custom_completion()
{
    __gz_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly gz allows handling aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="GZ_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __gz_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __gz_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __gz_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __gz_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __gz_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __gz_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __gz_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __gz_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __gz_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __gz_debug "Listing directories in $subdir"
            __gz_handle_subdirs_in_dir_flag "$subdir"
        else
            __gz_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__gz_handle_reply()
{
    __gz_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __gz_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __gz_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __gz_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __gz_custom_func >/dev/null; then
            # try command name qualified custom func
            __gz_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__gz_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__gz_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__gz_handle_flag()
{
    __gz_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __gz_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __gz_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __gz_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __gz_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __gz_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__gz_handle_noun()
{
    __gz_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __gz_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __gz_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__gz_handle_command()
{
    __gz_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_gz_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __gz_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__gz_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __gz_handle_reply
        return
    fi
    __gz_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __gz_handle_flag
    elif __gz_contains_word "${words[c]}" "${commands[@]}"; then
        __gz_handle_command
    elif [[ $c -eq 0 ]]; then
        __gz_handle_command
    elif __gz_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __gz_handle_command
        else
            __gz_handle_noun
        fi
    else
        __gz_handle_noun
    fi
    __gz_handle_word
}

_gz_dev-env_aws_list()
{
    last_command="gz_dev-env_aws_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws_load()
{
    last_command="gz_dev-env_aws_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws_save()
{
    last_command="gz_dev-env_aws_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws()
{
    last_command="gz_dev-env_aws"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-credentials_list()
{
    last_command="gz_dev-env_aws-credentials_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-credentials_load()
{
    last_command="gz_dev-env_aws-credentials_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-credentials_save()
{
    last_command="gz_dev-env_aws-credentials_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-credentials()
{
    last_command="gz_dev-env_aws-credentials"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile_list()
{
    last_command="gz_dev-env_aws-profile_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile_login()
{
    last_command="gz_dev-env_aws-profile_login"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile_show()
{
    last_command="gz_dev-env_aws-profile_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile_switch()
{
    last_command="gz_dev-env_aws-profile_switch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--interactive")
    flags+=("-i")
    local_nonpersistent_flags+=("--interactive")
    local_nonpersistent_flags+=("-i")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile_validate()
{
    last_command="gz_dev-env_aws-profile_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_aws-profile()
{
    last_command="gz_dev-env_aws-profile"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("login")
    commands+=("show")
    commands+=("switch")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_list()
{
    last_command="gz_dev-env_azure-subscription_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--tenant=")
    two_word_flags+=("--tenant")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--tenant")
    local_nonpersistent_flags+=("--tenant=")
    local_nonpersistent_flags+=("-t")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_login()
{
    last_command="gz_dev-env_azure-subscription_login"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--device-code")
    local_nonpersistent_flags+=("--device-code")
    flags+=("--service-principal")
    local_nonpersistent_flags+=("--service-principal")
    flags+=("--tenant=")
    two_word_flags+=("--tenant")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--tenant")
    local_nonpersistent_flags+=("--tenant=")
    local_nonpersistent_flags+=("-t")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_show()
{
    last_command="gz_dev-env_azure-subscription_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_switch()
{
    last_command="gz_dev-env_azure-subscription_switch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--interactive")
    flags+=("-i")
    local_nonpersistent_flags+=("--interactive")
    local_nonpersistent_flags+=("-i")
    flags+=("--tenant=")
    two_word_flags+=("--tenant")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--tenant")
    local_nonpersistent_flags+=("--tenant=")
    local_nonpersistent_flags+=("-t")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_tenant_list()
{
    last_command="gz_dev-env_azure-subscription_tenant_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_tenant_switch()
{
    last_command="gz_dev-env_azure-subscription_tenant_switch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_tenant()
{
    last_command="gz_dev-env_azure-subscription_tenant"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("switch")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription_validate()
{
    last_command="gz_dev-env_azure-subscription_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--check-permissions")
    local_nonpersistent_flags+=("--check-permissions")
    flags+=("--check-quotas")
    local_nonpersistent_flags+=("--check-quotas")
    flags+=("--check-resource-groups")
    local_nonpersistent_flags+=("--check-resource-groups")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_azure-subscription()
{
    last_command="gz_dev-env_azure-subscription"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("login")
    commands+=("show")
    commands+=("switch")
    commands+=("tenant")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_docker_list()
{
    last_command="gz_dev-env_docker_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_docker_load()
{
    last_command="gz_dev-env_docker_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_docker_save()
{
    last_command="gz_dev-env_docker_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_docker()
{
    last_command="gz_dev-env_docker"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud_list()
{
    last_command="gz_dev-env_gcloud_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud_load()
{
    last_command="gz_dev-env_gcloud_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud_save()
{
    last_command="gz_dev-env_gcloud_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud()
{
    last_command="gz_dev-env_gcloud"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud-credentials_list()
{
    last_command="gz_dev-env_gcloud-credentials_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud-credentials_load()
{
    last_command="gz_dev-env_gcloud-credentials_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud-credentials_save()
{
    last_command="gz_dev-env_gcloud-credentials_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcloud-credentials()
{
    last_command="gz_dev-env_gcloud-credentials"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_config_activate()
{
    last_command="gz_dev-env_gcp-project_config_activate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_config_create()
{
    last_command="gz_dev-env_gcp-project_config_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--account=")
    two_word_flags+=("--account")
    two_word_flags+=("-a")
    local_nonpersistent_flags+=("--account")
    local_nonpersistent_flags+=("--account=")
    local_nonpersistent_flags+=("-a")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--region=")
    two_word_flags+=("--region")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--region")
    local_nonpersistent_flags+=("--region=")
    local_nonpersistent_flags+=("-r")
    flags+=("--zone=")
    two_word_flags+=("--zone")
    two_word_flags+=("-z")
    local_nonpersistent_flags+=("--zone")
    local_nonpersistent_flags+=("--zone=")
    local_nonpersistent_flags+=("-z")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_config_delete()
{
    last_command="gz_dev-env_gcp-project_config_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_config_list()
{
    last_command="gz_dev-env_gcp-project_config_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_config()
{
    last_command="gz_dev-env_gcp-project_config"

    command_aliases=()

    commands=()
    commands+=("activate")
    commands+=("create")
    commands+=("delete")
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_list()
{
    last_command="gz_dev-env_gcp-project_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_activate()
{
    last_command="gz_dev-env_gcp-project_service-account_activate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--key-file=")
    two_word_flags+=("--key-file")
    two_word_flags+=("-k")
    local_nonpersistent_flags+=("--key-file")
    local_nonpersistent_flags+=("--key-file=")
    local_nonpersistent_flags+=("-k")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_create()
{
    last_command="gz_dev-env_gcp-project_service-account_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--description=")
    two_word_flags+=("--description")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    flags+=("--display-name=")
    two_word_flags+=("--display-name")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--display-name")
    local_nonpersistent_flags+=("--display-name=")
    local_nonpersistent_flags+=("-d")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_create-key()
{
    last_command="gz_dev-env_gcp-project_service-account_create-key"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--key-type=")
    two_word_flags+=("--key-type")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--key-type")
    local_nonpersistent_flags+=("--key-type=")
    local_nonpersistent_flags+=("-t")
    flags+=("--output-file=")
    two_word_flags+=("--output-file")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output-file")
    local_nonpersistent_flags+=("--output-file=")
    local_nonpersistent_flags+=("-o")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_delete()
{
    last_command="gz_dev-env_gcp-project_service-account_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_delete-key()
{
    last_command="gz_dev-env_gcp-project_service-account_delete-key"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_list()
{
    last_command="gz_dev-env_gcp-project_service-account_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account_show()
{
    last_command="gz_dev-env_gcp-project_service-account_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--project=")
    two_word_flags+=("--project")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--project")
    local_nonpersistent_flags+=("--project=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_service-account()
{
    last_command="gz_dev-env_gcp-project_service-account"

    command_aliases=()

    commands=()
    commands+=("activate")
    commands+=("create")
    commands+=("create-key")
    commands+=("delete")
    commands+=("delete-key")
    commands+=("list")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_show()
{
    last_command="gz_dev-env_gcp-project_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_switch()
{
    last_command="gz_dev-env_gcp-project_switch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--interactive")
    flags+=("-i")
    local_nonpersistent_flags+=("--interactive")
    local_nonpersistent_flags+=("-i")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project_validate()
{
    last_command="gz_dev-env_gcp-project_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--check-apis")
    local_nonpersistent_flags+=("--check-apis")
    flags+=("--check-billing")
    local_nonpersistent_flags+=("--check-billing")
    flags+=("--check-permissions")
    local_nonpersistent_flags+=("--check-permissions")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_gcp-project()
{
    last_command="gz_dev-env_gcp-project"

    command_aliases=()

    commands=()
    commands+=("config")
    commands+=("list")
    commands+=("service-account")
    commands+=("show")
    commands+=("switch")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_kubeconfig_list()
{
    last_command="gz_dev-env_kubeconfig_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_kubeconfig_load()
{
    last_command="gz_dev-env_kubeconfig_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_kubeconfig_save()
{
    last_command="gz_dev-env_kubeconfig_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_kubeconfig()
{
    last_command="gz_dev-env_kubeconfig"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_install-key()
{
    last_command="gz_dev-env_ssh_install-key"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--host=")
    two_word_flags+=("--host")
    local_nonpersistent_flags+=("--host")
    local_nonpersistent_flags+=("--host=")
    flags+=("--password=")
    two_word_flags+=("--password")
    local_nonpersistent_flags+=("--password")
    local_nonpersistent_flags+=("--password=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--private-key=")
    two_word_flags+=("--private-key")
    local_nonpersistent_flags+=("--private-key")
    local_nonpersistent_flags+=("--private-key=")
    flags+=("--public-key=")
    two_word_flags+=("--public-key")
    local_nonpersistent_flags+=("--public-key")
    local_nonpersistent_flags+=("--public-key=")
    flags+=("--user=")
    two_word_flags+=("--user")
    local_nonpersistent_flags+=("--user")
    local_nonpersistent_flags+=("--user=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--host=")
    must_have_one_flag+=("--user=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_install-key-simple()
{
    last_command="gz_dev-env_ssh_install-key-simple"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--host=")
    two_word_flags+=("--host")
    local_nonpersistent_flags+=("--host")
    local_nonpersistent_flags+=("--host=")
    flags+=("--public-key=")
    two_word_flags+=("--public-key")
    local_nonpersistent_flags+=("--public-key")
    local_nonpersistent_flags+=("--public-key=")
    flags+=("--user=")
    two_word_flags+=("--user")
    local_nonpersistent_flags+=("--user")
    local_nonpersistent_flags+=("--user=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--host=")
    must_have_one_flag+=("--public-key=")
    must_have_one_flag+=("--user=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_list()
{
    last_command="gz_dev-env_ssh_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_list-keys()
{
    last_command="gz_dev-env_ssh_list-keys"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_load()
{
    last_command="gz_dev-env_ssh_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh_save()
{
    last_command="gz_dev-env_ssh_save"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-path=")
    two_word_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path")
    local_nonpersistent_flags+=("--config-path=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--include-keys")
    local_nonpersistent_flags+=("--include-keys")
    flags+=("--include-public")
    local_nonpersistent_flags+=("--include-public")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--store-path=")
    two_word_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path")
    local_nonpersistent_flags+=("--store-path=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("-n")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_ssh()
{
    last_command="gz_dev-env_ssh"

    command_aliases=()

    commands=()
    commands+=("install-key")
    commands+=("install-key-simple")
    commands+=("list")
    commands+=("list-keys")
    commands+=("load")
    commands+=("save")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_status()
{
    last_command="gz_dev-env_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--check-health")
    local_nonpersistent_flags+=("--check-health")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--no-color")
    local_nonpersistent_flags+=("--no-color")
    flags+=("--service=")
    two_word_flags+=("--service")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--service")
    local_nonpersistent_flags+=("--service=")
    local_nonpersistent_flags+=("-s")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--watch")
    local_nonpersistent_flags+=("--watch")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_switch-all()
{
    last_command="gz_dev-env_switch-all"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--env=")
    two_word_flags+=("--env")
    local_nonpersistent_flags+=("--env")
    local_nonpersistent_flags+=("--env=")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--from-file=")
    two_word_flags+=("--from-file")
    local_nonpersistent_flags+=("--from-file")
    local_nonpersistent_flags+=("--from-file=")
    flags+=("--interactive")
    local_nonpersistent_flags+=("--interactive")
    flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env_tui()
{
    last_command="gz_dev-env_tui"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_dev-env()
{
    last_command="gz_dev-env"

    command_aliases=()

    commands=()
    commands+=("aws")
    commands+=("aws-credentials")
    commands+=("aws-profile")
    commands+=("azure-subscription")
    commands+=("docker")
    commands+=("gcloud")
    commands+=("gcloud-credentials")
    commands+=("gcp-project")
    commands+=("kubeconfig")
    commands+=("ssh")
    commands+=("status")
    commands+=("switch-all")
    commands+=("tui")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_apply()
{
    last_command="gz_git_config_apply"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--interactive")
    local_nonpersistent_flags+=("--interactive")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--template=")
    two_word_flags+=("--template")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("--template=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_audit()
{
    last_command="gz_git_config_audit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_dashboard()
{
    last_command="gz_git_config_dashboard"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-refresh")
    local_nonpersistent_flags+=("--auto-refresh")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--refresh-rate=")
    two_word_flags+=("--refresh-rate")
    local_nonpersistent_flags+=("--refresh-rate")
    local_nonpersistent_flags+=("--refresh-rate=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_diff()
{
    last_command="gz_git_config_diff"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--detailed")
    local_nonpersistent_flags+=("--detailed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--group-by-impact")
    local_nonpersistent_flags+=("--group-by-impact")
    flags+=("--impact=")
    two_word_flags+=("--impact")
    local_nonpersistent_flags+=("--impact")
    local_nonpersistent_flags+=("--impact=")
    flags+=("--non-compliant")
    local_nonpersistent_flags+=("--non-compliant")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--show-values")
    local_nonpersistent_flags+=("--show-values")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_list()
{
    last_command="gz_git_config_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--show-config")
    local_nonpersistent_flags+=("--show-config")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_risk-assessment()
{
    last_command="gz_git_config_risk-assessment"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--risk-threshold=")
    two_word_flags+=("--risk-threshold")
    local_nonpersistent_flags+=("--risk-threshold")
    local_nonpersistent_flags+=("--risk-threshold=")
    flags+=("--severity=")
    two_word_flags+=("--severity")
    local_nonpersistent_flags+=("--severity")
    local_nonpersistent_flags+=("--severity=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_template_list()
{
    last_command="gz_git_config_template_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_template_show()
{
    last_command="gz_git_config_template_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_template_validate()
{
    last_command="gz_git_config_template_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--strict")
    local_nonpersistent_flags+=("--strict")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_template()
{
    last_command="gz_git_config_template"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("show")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_validate()
{
    last_command="gz_git_config_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--strict")
    local_nonpersistent_flags+=("--strict")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_bulk()
{
    last_command="gz_git_config_webhook_bulk"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--dry-run-bulk")
    local_nonpersistent_flags+=("--dry-run-bulk")
    flags+=("--operation=")
    two_word_flags+=("--operation")
    local_nonpersistent_flags+=("--operation")
    local_nonpersistent_flags+=("--operation=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--parallel-jobs=")
    two_word_flags+=("--parallel-jobs")
    local_nonpersistent_flags+=("--parallel-jobs")
    local_nonpersistent_flags+=("--parallel-jobs=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--webhook-config=")
    two_word_flags+=("--webhook-config")
    local_nonpersistent_flags+=("--webhook-config")
    local_nonpersistent_flags+=("--webhook-config=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--operation=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_create()
{
    last_command="gz_git_config_webhook_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_delete()
{
    last_command="gz_git_config_webhook_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_get()
{
    last_command="gz_git_config_webhook_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_list()
{
    last_command="gz_git_config_webhook_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook_update()
{
    last_command="gz_git_config_webhook_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config_webhook()
{
    last_command="gz_git_config_webhook"

    command_aliases=()

    commands=()
    commands+=("bulk")
    commands+=("create")
    commands+=("delete")
    commands+=("get")
    commands+=("list")
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_config()
{
    last_command="gz_git_config"

    command_aliases=()

    commands=()
    commands+=("apply")
    commands+=("audit")
    commands+=("dashboard")
    commands+=("diff")
    commands+=("list")
    commands+=("risk-assessment")
    commands+=("template")
    commands+=("validate")
    commands+=("webhook")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_archive()
{
    last_command="gz_git_repo_archive"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--match=")
    two_word_flags+=("--match")
    local_nonpersistent_flags+=("--match")
    local_nonpersistent_flags+=("--match=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--unarchive")
    local_nonpersistent_flags+=("--unarchive")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_clone()
{
    last_command="gz_git_repo_clone"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--branch=")
    two_word_flags+=("--branch")
    local_nonpersistent_flags+=("--branch")
    local_nonpersistent_flags+=("--branch=")
    flags+=("--cleanup-orphans")
    local_nonpersistent_flags+=("--cleanup-orphans")
    flags+=("--config=")
    two_word_flags+=("--config")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    flags+=("--create-gzh-file")
    local_nonpersistent_flags+=("--create-gzh-file")
    flags+=("--depth=")
    two_word_flags+=("--depth")
    local_nonpersistent_flags+=("--depth")
    local_nonpersistent_flags+=("--depth=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--include-forks")
    local_nonpersistent_flags+=("--include-forks")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--match=")
    two_word_flags+=("--match")
    local_nonpersistent_flags+=("--match")
    local_nonpersistent_flags+=("--match=")
    flags+=("--max-retries=")
    two_word_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries=")
    flags+=("--max-stars=")
    two_word_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars=")
    flags+=("--min-stars=")
    two_word_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--password=")
    two_word_flags+=("--password")
    local_nonpersistent_flags+=("--password")
    local_nonpersistent_flags+=("--password=")
    flags+=("--protocol=")
    two_word_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--resume=")
    two_word_flags+=("--resume")
    local_nonpersistent_flags+=("--resume")
    local_nonpersistent_flags+=("--resume=")
    flags+=("--retry-delay=")
    two_word_flags+=("--retry-delay")
    local_nonpersistent_flags+=("--retry-delay")
    local_nonpersistent_flags+=("--retry-delay=")
    flags+=("--single-branch")
    local_nonpersistent_flags+=("--single-branch")
    flags+=("--strategy=")
    two_word_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy=")
    flags+=("--target=")
    two_word_flags+=("--target")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--topics=")
    two_word_flags+=("--topics")
    local_nonpersistent_flags+=("--topics")
    local_nonpersistent_flags+=("--topics=")
    flags+=("--updated-since=")
    two_word_flags+=("--updated-since")
    local_nonpersistent_flags+=("--updated-since")
    local_nonpersistent_flags+=("--updated-since=")
    flags+=("--username=")
    two_word_flags+=("--username")
    local_nonpersistent_flags+=("--username")
    local_nonpersistent_flags+=("--username=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--visibility=")
    two_word_flags+=("--visibility")
    local_nonpersistent_flags+=("--visibility")
    local_nonpersistent_flags+=("--visibility=")
    flags+=("--debug")
    flags+=("--experimental")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_clone-or-update()
{
    last_command="gz_git_repo_clone-or-update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--branch=")
    two_word_flags+=("--branch")
    two_word_flags+=("-b")
    local_nonpersistent_flags+=("--branch")
    local_nonpersistent_flags+=("--branch=")
    local_nonpersistent_flags+=("-b")
    flags+=("--depth=")
    two_word_flags+=("--depth")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--depth")
    local_nonpersistent_flags+=("--depth=")
    local_nonpersistent_flags+=("-d")
    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")
    local_nonpersistent_flags+=("-f")
    flags+=("--strategy=")
    two_word_flags+=("--strategy")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy=")
    local_nonpersistent_flags+=("-s")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_create()
{
    last_command="gz_git_repo_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--allow-merge-commit")
    local_nonpersistent_flags+=("--allow-merge-commit")
    flags+=("--allow-rebase-merge")
    local_nonpersistent_flags+=("--allow-rebase-merge")
    flags+=("--allow-squash-merge")
    local_nonpersistent_flags+=("--allow-squash-merge")
    flags+=("--auto-init")
    local_nonpersistent_flags+=("--auto-init")
    flags+=("--default-branch=")
    two_word_flags+=("--default-branch")
    local_nonpersistent_flags+=("--default-branch")
    local_nonpersistent_flags+=("--default-branch=")
    flags+=("--description=")
    two_word_flags+=("--description")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--gitignore=")
    two_word_flags+=("--gitignore")
    local_nonpersistent_flags+=("--gitignore")
    local_nonpersistent_flags+=("--gitignore=")
    flags+=("--homepage=")
    two_word_flags+=("--homepage")
    local_nonpersistent_flags+=("--homepage")
    local_nonpersistent_flags+=("--homepage=")
    flags+=("--issues")
    local_nonpersistent_flags+=("--issues")
    flags+=("--license=")
    two_word_flags+=("--license")
    local_nonpersistent_flags+=("--license")
    local_nonpersistent_flags+=("--license=")
    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--private")
    local_nonpersistent_flags+=("--private")
    flags+=("--projects")
    local_nonpersistent_flags+=("--projects")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--template=")
    two_word_flags+=("--template")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("--template=")
    flags+=("--topics=")
    two_word_flags+=("--topics")
    local_nonpersistent_flags+=("--topics")
    local_nonpersistent_flags+=("--topics=")
    flags+=("--wiki")
    local_nonpersistent_flags+=("--wiki")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_delete()
{
    last_command="gz_git_repo_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup")
    local_nonpersistent_flags+=("--backup")
    flags+=("--backup-format=")
    two_word_flags+=("--backup-format")
    local_nonpersistent_flags+=("--backup-format")
    local_nonpersistent_flags+=("--backup-format=")
    flags+=("--backup-path=")
    two_word_flags+=("--backup-path")
    local_nonpersistent_flags+=("--backup-path")
    local_nonpersistent_flags+=("--backup-path=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--match=")
    two_word_flags+=("--match")
    local_nonpersistent_flags+=("--match")
    local_nonpersistent_flags+=("--match=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_list()
{
    last_command="gz_git_repo_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all-providers")
    local_nonpersistent_flags+=("--all-providers")
    flags+=("--archived-only")
    local_nonpersistent_flags+=("--archived-only")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--match=")
    two_word_flags+=("--match")
    local_nonpersistent_flags+=("--match")
    local_nonpersistent_flags+=("--match=")
    flags+=("--max-stars=")
    two_word_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars=")
    flags+=("--min-stars=")
    two_word_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars=")
    flags+=("--no-archived")
    local_nonpersistent_flags+=("--no-archived")
    flags+=("--order=")
    two_word_flags+=("--order")
    local_nonpersistent_flags+=("--order")
    local_nonpersistent_flags+=("--order=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    flags+=("--updated-since=")
    two_word_flags+=("--updated-since")
    local_nonpersistent_flags+=("--updated-since")
    local_nonpersistent_flags+=("--updated-since=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--visibility=")
    two_word_flags+=("--visibility")
    local_nonpersistent_flags+=("--visibility")
    local_nonpersistent_flags+=("--visibility=")
    flags+=("--debug")
    flags+=("--experimental")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_migrate()
{
    last_command="gz_git_repo_migrate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_pull-all()
{
    last_command="gz_git_repo_pull-all"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--exclude-pattern=")
    two_word_flags+=("--exclude-pattern")
    local_nonpersistent_flags+=("--exclude-pattern")
    local_nonpersistent_flags+=("--exclude-pattern=")
    flags+=("--include-pattern=")
    two_word_flags+=("--include-pattern")
    local_nonpersistent_flags+=("--include-pattern")
    local_nonpersistent_flags+=("--include-pattern=")
    flags+=("--json")
    local_nonpersistent_flags+=("--json")
    flags+=("--max-depth=")
    two_word_flags+=("--max-depth")
    local_nonpersistent_flags+=("--max-depth")
    local_nonpersistent_flags+=("--max-depth=")
    flags+=("--no-fetch")
    local_nonpersistent_flags+=("--no-fetch")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    local_nonpersistent_flags+=("-p")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_search()
{
    last_command="gz_git_repo_search"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--order=")
    two_word_flags+=("--order")
    local_nonpersistent_flags+=("--order")
    local_nonpersistent_flags+=("--order=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--page=")
    two_word_flags+=("--page")
    local_nonpersistent_flags+=("--page")
    local_nonpersistent_flags+=("--page=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--query=")
    two_word_flags+=("--query")
    local_nonpersistent_flags+=("--query")
    local_nonpersistent_flags+=("--query=")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    flags+=("--stars=")
    two_word_flags+=("--stars")
    local_nonpersistent_flags+=("--stars")
    local_nonpersistent_flags+=("--stars=")
    flags+=("--topic=")
    two_word_flags+=("--topic")
    local_nonpersistent_flags+=("--topic")
    local_nonpersistent_flags+=("--topic=")
    flags+=("--user=")
    two_word_flags+=("--user")
    local_nonpersistent_flags+=("--user")
    local_nonpersistent_flags+=("--user=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--provider=")
    must_have_one_flag+=("--query=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo_sync()
{
    last_command="gz_git_repo_sync"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--create-missing")
    local_nonpersistent_flags+=("--create-missing")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--from=")
    two_word_flags+=("--from")
    local_nonpersistent_flags+=("--from")
    local_nonpersistent_flags+=("--from=")
    flags+=("--include-code")
    local_nonpersistent_flags+=("--include-code")
    flags+=("--include-issues")
    local_nonpersistent_flags+=("--include-issues")
    flags+=("--include-prs")
    local_nonpersistent_flags+=("--include-prs")
    flags+=("--include-releases")
    local_nonpersistent_flags+=("--include-releases")
    flags+=("--include-settings")
    local_nonpersistent_flags+=("--include-settings")
    flags+=("--include-wiki")
    local_nonpersistent_flags+=("--include-wiki")
    flags+=("--match=")
    two_word_flags+=("--match")
    local_nonpersistent_flags+=("--match")
    local_nonpersistent_flags+=("--match=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--to=")
    two_word_flags+=("--to")
    local_nonpersistent_flags+=("--to")
    local_nonpersistent_flags+=("--to=")
    flags+=("--update-existing")
    local_nonpersistent_flags+=("--update-existing")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--from=")
    must_have_one_flag+=("--to=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_repo()
{
    last_command="gz_git_repo"

    command_aliases=()

    commands=()
    commands+=("archive")
    commands+=("clone")
    commands+=("clone-or-update")
    commands+=("create")
    commands+=("delete")
    commands+=("list")
    commands+=("migrate")
    commands+=("pull-all")
    commands+=("search")
    commands+=("sync")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_bulk_create()
{
    last_command="gz_git_webhook_bulk_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--repos=")
    two_word_flags+=("--repos")
    local_nonpersistent_flags+=("--repos")
    local_nonpersistent_flags+=("--repos=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_bulk()
{
    last_command="gz_git_webhook_bulk"

    command_aliases=()

    commands=()
    commands+=("create")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_org_get()
{
    last_command="gz_git_webhook_config_org_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_org_update()
{
    last_command="gz_git_webhook_config_org_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_org_validate()
{
    last_command="gz_git_webhook_config_org_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_org()
{
    last_command="gz_git_webhook_config_org"

    command_aliases=()

    commands=()
    commands+=("get")
    commands+=("update")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_policy_apply()
{
    last_command="gz_git_webhook_config_policy_apply"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--policies=")
    two_word_flags+=("--policies")
    local_nonpersistent_flags+=("--policies")
    local_nonpersistent_flags+=("--policies=")
    flags+=("--repos=")
    two_word_flags+=("--repos")
    local_nonpersistent_flags+=("--repos")
    local_nonpersistent_flags+=("--repos=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_policy_create()
{
    last_command="gz_git_webhook_config_policy_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_policy_list()
{
    last_command="gz_git_webhook_config_policy_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_policy_preview()
{
    last_command="gz_git_webhook_config_policy_preview"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--policies=")
    two_word_flags+=("--policies")
    local_nonpersistent_flags+=("--policies")
    local_nonpersistent_flags+=("--policies=")
    flags+=("--repos=")
    two_word_flags+=("--repos")
    local_nonpersistent_flags+=("--repos")
    local_nonpersistent_flags+=("--repos=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_policy()
{
    last_command="gz_git_webhook_config_policy"

    command_aliases=()

    commands=()
    commands+=("apply")
    commands+=("create")
    commands+=("list")
    commands+=("preview")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_report_compliance()
{
    last_command="gz_git_webhook_config_report_compliance"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_report_inventory()
{
    last_command="gz_git_webhook_config_report_inventory"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_report_sync()
{
    last_command="gz_git_webhook_config_report_sync"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config_report()
{
    last_command="gz_git_webhook_config_report"

    command_aliases=()

    commands=()
    commands+=("compliance")
    commands+=("inventory")
    commands+=("sync")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_config()
{
    last_command="gz_git_webhook_config"

    command_aliases=()

    commands=()
    commands+=("org")
    commands+=("policy")
    commands+=("report")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_monitor_deliveries()
{
    last_command="gz_git_webhook_monitor_deliveries"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_monitor_test()
{
    last_command="gz_git_webhook_monitor_test"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_monitor()
{
    last_command="gz_git_webhook_monitor"

    command_aliases=()

    commands=()
    commands+=("deliveries")
    commands+=("test")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_org_create()
{
    last_command="gz_git_webhook_org_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_org_list()
{
    last_command="gz_git_webhook_org_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_org()
{
    last_command="gz_git_webhook_org"

    command_aliases=()

    commands=()
    commands+=("create")
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo_create()
{
    last_command="gz_git_webhook_repo_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--name=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo_delete()
{
    last_command="gz_git_webhook_repo_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo_get()
{
    last_command="gz_git_webhook_repo_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo_list()
{
    last_command="gz_git_webhook_repo_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo_update()
{
    last_command="gz_git_webhook_repo_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook_repo()
{
    last_command="gz_git_webhook_repo"

    command_aliases=()

    commands=()
    commands+=("create")
    commands+=("delete")
    commands+=("get")
    commands+=("list")
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git_webhook()
{
    last_command="gz_git_webhook"

    command_aliases=()

    commands=()
    commands+=("bulk")
    commands+=("config")
    commands+=("monitor")
    commands+=("org")
    commands+=("repo")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git()
{
    last_command="gz_git"

    command_aliases=()

    commands=()
    commands+=("config")
    commands+=("repo")
    commands+=("webhook")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync_config_generate()
{
    last_command="gz_git-sync_config_generate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--base-url=")
    two_word_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url=")
    flags+=("--clone-proto=")
    two_word_flags+=("--clone-proto")
    local_nonpersistent_flags+=("--clone-proto")
    local_nonpersistent_flags+=("--clone-proto=")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    flags+=("--full")
    local_nonpersistent_flags+=("--full")
    flags+=("--include=")
    two_word_flags+=("--include")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--include-forks")
    local_nonpersistent_flags+=("--include-forks")
    flags+=("--include-private")
    local_nonpersistent_flags+=("--include-private")
    flags+=("--include-subgroups")
    local_nonpersistent_flags+=("--include-subgroups")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--last-push-within=")
    two_word_flags+=("--last-push-within")
    local_nonpersistent_flags+=("--last-push-within")
    local_nonpersistent_flags+=("--last-push-within=")
    flags+=("--max-retries=")
    two_word_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries=")
    flags+=("--max-stars=")
    two_word_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars=")
    flags+=("--min-stars=")
    two_word_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--ssh-port=")
    two_word_flags+=("--ssh-port")
    local_nonpersistent_flags+=("--ssh-port")
    local_nonpersistent_flags+=("--ssh-port=")
    flags+=("--subgroup-mode=")
    two_word_flags+=("--subgroup-mode")
    local_nonpersistent_flags+=("--subgroup-mode")
    local_nonpersistent_flags+=("--subgroup-mode=")
    flags+=("--sync-strategy=")
    two_word_flags+=("--sync-strategy")
    local_nonpersistent_flags+=("--sync-strategy")
    local_nonpersistent_flags+=("--sync-strategy=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--user")
    local_nonpersistent_flags+=("--user")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--path=")
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync_config()
{
    last_command="gz_git-sync_config"

    command_aliases=()

    commands=()
    commands+=("generate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync_from()
{
    last_command="gz_git-sync_from"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--base-url=")
    two_word_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url=")
    flags+=("--cleanup-orphans")
    local_nonpersistent_flags+=("--cleanup-orphans")
    flags+=("--clone-proto=")
    two_word_flags+=("--clone-proto")
    local_nonpersistent_flags+=("--clone-proto")
    local_nonpersistent_flags+=("--clone-proto=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    flags+=("--include=")
    two_word_flags+=("--include")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--include-forks")
    local_nonpersistent_flags+=("--include-forks")
    flags+=("--include-private")
    local_nonpersistent_flags+=("--include-private")
    flags+=("--include-subgroups")
    local_nonpersistent_flags+=("--include-subgroups")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--last-push-within=")
    two_word_flags+=("--last-push-within")
    local_nonpersistent_flags+=("--last-push-within")
    local_nonpersistent_flags+=("--last-push-within=")
    flags+=("--max-retries=")
    two_word_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries=")
    flags+=("--max-stars=")
    two_word_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars=")
    flags+=("--min-stars=")
    two_word_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--resume")
    local_nonpersistent_flags+=("--resume")
    flags+=("--ssh-key=")
    two_word_flags+=("--ssh-key")
    local_nonpersistent_flags+=("--ssh-key")
    local_nonpersistent_flags+=("--ssh-key=")
    flags+=("--ssh-key-content=")
    two_word_flags+=("--ssh-key-content")
    local_nonpersistent_flags+=("--ssh-key-content")
    local_nonpersistent_flags+=("--ssh-key-content=")
    flags+=("--ssh-port=")
    two_word_flags+=("--ssh-port")
    local_nonpersistent_flags+=("--ssh-port")
    local_nonpersistent_flags+=("--ssh-port=")
    flags+=("--state-file=")
    two_word_flags+=("--state-file")
    local_nonpersistent_flags+=("--state-file")
    local_nonpersistent_flags+=("--state-file=")
    flags+=("--subgroup-mode=")
    two_word_flags+=("--subgroup-mode")
    local_nonpersistent_flags+=("--subgroup-mode")
    local_nonpersistent_flags+=("--subgroup-mode=")
    flags+=("--sync-strategy=")
    two_word_flags+=("--sync-strategy")
    local_nonpersistent_flags+=("--sync-strategy")
    local_nonpersistent_flags+=("--sync-strategy=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--user")
    local_nonpersistent_flags+=("--user")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--path=")
    must_have_one_flag+=("--provider=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync_setup()
{
    last_command="gz_git-sync_setup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync_status()
{
    last_command="gz_git-sync_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--history-dir=")
    two_word_flags+=("--history-dir")
    local_nonpersistent_flags+=("--history-dir")
    local_nonpersistent_flags+=("--history-dir=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    two_word_flags+=("-j")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    local_nonpersistent_flags+=("-j")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--save-history")
    local_nonpersistent_flags+=("--save-history")
    flags+=("--scan-depth=")
    two_word_flags+=("--scan-depth")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--scan-depth")
    local_nonpersistent_flags+=("--scan-depth=")
    local_nonpersistent_flags+=("-d")
    flags+=("--skip-fetch")
    local_nonpersistent_flags+=("--skip-fetch")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--tui")
    local_nonpersistent_flags+=("--tui")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_git-sync()
{
    last_command="gz_git-sync"

    command_aliases=()

    commands=()
    commands+=("config")
    commands+=("from")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("from-forge")
        aliashash["from-forge"]="from"
    fi
    commands+=("setup")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_fix-sync()
{
    last_command="gz_ide_fix-sync"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--product=")
    two_word_flags+=("--product")
    local_nonpersistent_flags+=("--product")
    local_nonpersistent_flags+=("--product=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_list()
{
    last_command="gz_ide_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_monitor()
{
    last_command="gz_ide_monitor"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--daemon")
    local_nonpersistent_flags+=("--daemon")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    flags+=("--log=")
    two_word_flags+=("--log")
    local_nonpersistent_flags+=("--log")
    local_nonpersistent_flags+=("--log=")
    flags+=("--product=")
    two_word_flags+=("--product")
    local_nonpersistent_flags+=("--product")
    local_nonpersistent_flags+=("--product=")
    flags+=("--recursive")
    local_nonpersistent_flags+=("--recursive")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--watch-dir=")
    two_word_flags+=("--watch-dir")
    local_nonpersistent_flags+=("--watch-dir")
    local_nonpersistent_flags+=("--watch-dir=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_open()
{
    last_command="gz_ide_open"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--background")
    local_nonpersistent_flags+=("--background")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--wait")
    local_nonpersistent_flags+=("--wait")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_scan()
{
    last_command="gz_ide_scan"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--refresh")
    local_nonpersistent_flags+=("--refresh")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide_status()
{
    last_command="gz_ide_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_ide()
{
    last_command="gz_ide"

    command_aliases=()

    commands=()
    commands+=("fix-sync")
    commands+=("list")
    commands+=("monitor")
    commands+=("open")
    commands+=("scan")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_profile_delete()
{
    last_command="gz_net-env_profile_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_profile_init()
{
    last_command="gz_net-env_profile_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_profile_list()
{
    last_command="gz_net-env_profile_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_profile_show()
{
    last_command="gz_net-env_profile_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_profile()
{
    last_command="gz_net-env_profile"

    command_aliases=()

    commands=()
    commands+=("delete")
    commands+=("init")
    commands+=("list")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_status()
{
    last_command="gz_net-env_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--health")
    local_nonpersistent_flags+=("--health")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env_watch()
{
    last_command="gz_net-env_watch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--interval=")
    two_word_flags+=("--interval")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--interval")
    local_nonpersistent_flags+=("--interval=")
    local_nonpersistent_flags+=("-i")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_net-env()
{
    last_command="gz_net-env"

    command_aliases=()

    commands=()
    commands+=("profile")
    commands+=("status")
    commands+=("watch")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_analyze()
{
    last_command="gz_profile_analyze"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-suggest")
    local_nonpersistent_flags+=("--auto-suggest")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--threshold=")
    two_word_flags+=("--threshold")
    local_nonpersistent_flags+=("--threshold")
    local_nonpersistent_flags+=("--threshold=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_compare()
{
    last_command="gz_profile_compare"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--threshold=")
    two_word_flags+=("--threshold")
    local_nonpersistent_flags+=("--threshold")
    local_nonpersistent_flags+=("--threshold=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_continuous()
{
    last_command="gz_profile_continuous"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-analyze")
    local_nonpersistent_flags+=("--auto-analyze")
    flags+=("--duration=")
    two_word_flags+=("--duration")
    local_nonpersistent_flags+=("--duration")
    local_nonpersistent_flags+=("--duration=")
    flags+=("--interval=")
    two_word_flags+=("--interval")
    local_nonpersistent_flags+=("--interval")
    local_nonpersistent_flags+=("--interval=")
    flags+=("--type=")
    two_word_flags+=("--type")
    local_nonpersistent_flags+=("--type")
    local_nonpersistent_flags+=("--type=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_cpu()
{
    last_command="gz_profile_cpu"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--duration=")
    two_word_flags+=("--duration")
    local_nonpersistent_flags+=("--duration")
    local_nonpersistent_flags+=("--duration=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_memory()
{
    last_command="gz_profile_memory"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_server()
{
    last_command="gz_profile_server"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile_stats()
{
    last_command="gz_profile_stats"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_profile()
{
    last_command="gz_profile"

    command_aliases=()

    commands=()
    commands+=("analyze")
    commands+=("compare")
    commands+=("continuous")
    commands+=("cpu")
    commands+=("memory")
    commands+=("server")
    commands+=("stats")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_analyze()
{
    last_command="gz_quality_analyze"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_check()
{
    last_command="gz_quality_check"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--report=")
    two_word_flags+=("--report")
    local_nonpersistent_flags+=("--report")
    local_nonpersistent_flags+=("--report=")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_init()
{
    last_command="gz_quality_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_install()
{
    last_command="gz_quality_install"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_list()
{
    last_command="gz_quality_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_run()
{
    last_command="gz_quality_run"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--format-only")
    local_nonpersistent_flags+=("--format-only")
    flags+=("--lint-only")
    local_nonpersistent_flags+=("--lint-only")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--report=")
    two_word_flags+=("--report")
    local_nonpersistent_flags+=("--report")
    local_nonpersistent_flags+=("--report=")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_black()
{
    last_command="gz_quality_tool_black"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_cargo-fmt()
{
    last_command="gz_quality_tool_cargo-fmt"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_clippy()
{
    last_command="gz_quality_tool_clippy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_eslint()
{
    last_command="gz_quality_tool_eslint"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_gofumpt()
{
    last_command="gz_quality_tool_gofumpt"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_goimports()
{
    last_command="gz_quality_tool_goimports"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_golangci-lint()
{
    last_command="gz_quality_tool_golangci-lint"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_prettier()
{
    last_command="gz_quality_tool_prettier"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_pylint()
{
    last_command="gz_quality_tool_pylint"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_ruff()
{
    last_command="gz_quality_tool_ruff"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_rustfmt()
{
    last_command="gz_quality_tool_rustfmt"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool_tsc()
{
    last_command="gz_quality_tool_tsc"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_tool()
{
    last_command="gz_quality_tool"

    command_aliases=()

    commands=()
    commands+=("black")
    commands+=("cargo-fmt")
    commands+=("clippy")
    commands+=("eslint")
    commands+=("gofumpt")
    commands+=("goimports")
    commands+=("golangci-lint")
    commands+=("prettier")
    commands+=("pylint")
    commands+=("ruff")
    commands+=("rustfmt")
    commands+=("tsc")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--changed")
    local_nonpersistent_flags+=("--changed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--extra-args=")
    two_word_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args")
    local_nonpersistent_flags+=("--extra-args=")
    flags+=("--files=")
    two_word_flags+=("--files")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--files")
    local_nonpersistent_flags+=("--files=")
    local_nonpersistent_flags+=("-f")
    flags+=("--fix")
    flags+=("-x")
    local_nonpersistent_flags+=("--fix")
    local_nonpersistent_flags+=("-x")
    flags+=("--since=")
    two_word_flags+=("--since")
    local_nonpersistent_flags+=("--since")
    local_nonpersistent_flags+=("--since=")
    flags+=("--staged")
    local_nonpersistent_flags+=("--staged")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--workers=")
    two_word_flags+=("--workers")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workers")
    local_nonpersistent_flags+=("--workers=")
    local_nonpersistent_flags+=("-w")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_upgrade()
{
    last_command="gz_quality_upgrade"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality_version()
{
    last_command="gz_quality_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_quality()
{
    last_command="gz_quality"

    command_aliases=()

    commands=()
    commands+=("analyze")
    commands+=("check")
    commands+=("init")
    commands+=("install")
    commands+=("list")
    commands+=("run")
    commands+=("tool")
    commands+=("upgrade")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_apply()
{
    last_command="gz_repo-config_apply"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--interactive")
    local_nonpersistent_flags+=("--interactive")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--template=")
    two_word_flags+=("--template")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("--template=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_audit()
{
    last_command="gz_repo-config_audit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_dashboard()
{
    last_command="gz_repo-config_dashboard"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-refresh")
    local_nonpersistent_flags+=("--auto-refresh")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--refresh-rate=")
    two_word_flags+=("--refresh-rate")
    local_nonpersistent_flags+=("--refresh-rate")
    local_nonpersistent_flags+=("--refresh-rate=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_diff()
{
    last_command="gz_repo-config_diff"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--detailed")
    local_nonpersistent_flags+=("--detailed")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--group-by-impact")
    local_nonpersistent_flags+=("--group-by-impact")
    flags+=("--impact=")
    two_word_flags+=("--impact")
    local_nonpersistent_flags+=("--impact")
    local_nonpersistent_flags+=("--impact=")
    flags+=("--non-compliant")
    local_nonpersistent_flags+=("--non-compliant")
    flags+=("--org=")
    two_word_flags+=("--org")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    local_nonpersistent_flags+=("-o")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--show-values")
    local_nonpersistent_flags+=("--show-values")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_list()
{
    last_command="gz_repo-config_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--show-config")
    local_nonpersistent_flags+=("--show-config")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_risk-assessment()
{
    last_command="gz_repo-config_risk-assessment"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--risk-threshold=")
    two_word_flags+=("--risk-threshold")
    local_nonpersistent_flags+=("--risk-threshold")
    local_nonpersistent_flags+=("--risk-threshold=")
    flags+=("--severity=")
    two_word_flags+=("--severity")
    local_nonpersistent_flags+=("--severity")
    local_nonpersistent_flags+=("--severity=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_template_list()
{
    last_command="gz_repo-config_template_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_template_show()
{
    last_command="gz_repo-config_template_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_template_validate()
{
    last_command="gz_repo-config_template_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--strict")
    local_nonpersistent_flags+=("--strict")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_template()
{
    last_command="gz_repo-config_template"

    command_aliases=()

    commands=()
    commands+=("list")
    commands+=("show")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_validate()
{
    last_command="gz_repo-config_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--strict")
    local_nonpersistent_flags+=("--strict")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_bulk()
{
    last_command="gz_repo-config_webhook_bulk"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--dry-run-bulk")
    local_nonpersistent_flags+=("--dry-run-bulk")
    flags+=("--operation=")
    two_word_flags+=("--operation")
    local_nonpersistent_flags+=("--operation")
    local_nonpersistent_flags+=("--operation=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--parallel-jobs=")
    two_word_flags+=("--parallel-jobs")
    local_nonpersistent_flags+=("--parallel-jobs")
    local_nonpersistent_flags+=("--parallel-jobs=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--webhook-config=")
    two_word_flags+=("--webhook-config")
    local_nonpersistent_flags+=("--webhook-config")
    local_nonpersistent_flags+=("--webhook-config=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--operation=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_create()
{
    last_command="gz_repo-config_webhook_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_delete()
{
    last_command="gz_repo-config_webhook_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_get()
{
    last_command="gz_repo-config_webhook_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_list()
{
    last_command="gz_repo-config_webhook_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--output=")
    two_word_flags+=("--output")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook_update()
{
    last_command="gz_repo-config_webhook_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--active")
    local_nonpersistent_flags+=("--active")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file")
    local_nonpersistent_flags+=("--config-file=")
    flags+=("--content-type=")
    two_word_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type")
    local_nonpersistent_flags+=("--content-type=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--events=")
    two_word_flags+=("--events")
    local_nonpersistent_flags+=("--events")
    local_nonpersistent_flags+=("--events=")
    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--secret=")
    two_word_flags+=("--secret")
    local_nonpersistent_flags+=("--secret")
    local_nonpersistent_flags+=("--secret=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--id=")
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--repo=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config_webhook()
{
    last_command="gz_repo-config_webhook"

    command_aliases=()

    commands=()
    commands+=("bulk")
    commands+=("create")
    commands+=("delete")
    commands+=("get")
    commands+=("list")
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_repo-config()
{
    last_command="gz_repo-config"

    command_aliases=()

    commands=()
    commands+=("apply")
    commands+=("audit")
    commands+=("dashboard")
    commands+=("diff")
    commands+=("list")
    commands+=("risk-assessment")
    commands+=("template")
    commands+=("validate")
    commands+=("webhook")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_selfupdate()
{
    last_command="gz_selfupdate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_backup()
{
    last_command="gz_shellforge_backup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup-dir=")
    two_word_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir=")
    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--message=")
    two_word_flags+=("--message")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--message")
    local_nonpersistent_flags+=("--message=")
    local_nonpersistent_flags+=("-m")
    flags+=("--no-git")
    local_nonpersistent_flags+=("--no-git")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--file=")
    must_have_one_flag+=("-f")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_build()
{
    last_command="gz_shellforge_build"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-dir=")
    two_word_flags+=("--config-dir")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config-dir")
    local_nonpersistent_flags+=("--config-dir=")
    local_nonpersistent_flags+=("-c")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--manifest=")
    two_word_flags+=("--manifest")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--manifest")
    local_nonpersistent_flags+=("--manifest=")
    local_nonpersistent_flags+=("-m")
    flags+=("--os=")
    two_word_flags+=("--os")
    local_nonpersistent_flags+=("--os")
    local_nonpersistent_flags+=("--os=")
    flags+=("--output-dir=")
    two_word_flags+=("--output-dir")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--output-dir")
    local_nonpersistent_flags+=("--output-dir=")
    local_nonpersistent_flags+=("-d")
    flags+=("--shell=")
    two_word_flags+=("--shell")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--shell")
    local_nonpersistent_flags+=("--shell=")
    local_nonpersistent_flags+=("-s")
    flags+=("--target=")
    two_word_flags+=("--target")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    local_nonpersistent_flags+=("-t")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_cleanup()
{
    last_command="gz_shellforge_cleanup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup-dir=")
    two_word_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--keep-count=")
    two_word_flags+=("--keep-count")
    local_nonpersistent_flags+=("--keep-count")
    local_nonpersistent_flags+=("--keep-count=")
    flags+=("--keep-days=")
    two_word_flags+=("--keep-days")
    local_nonpersistent_flags+=("--keep-days")
    local_nonpersistent_flags+=("--keep-days=")
    flags+=("--no-git")
    local_nonpersistent_flags+=("--no-git")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--file=")
    must_have_one_flag+=("-f")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_deploy()
{
    last_command="gz_shellforge_deploy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup")
    local_nonpersistent_flags+=("--backup")
    flags+=("--build-dir=")
    two_word_flags+=("--build-dir")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--build-dir")
    local_nonpersistent_flags+=("--build-dir=")
    local_nonpersistent_flags+=("-d")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_diff()
{
    last_command="gz_shellforge_diff"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-F")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-F")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_doctor()
{
    last_command="gz_shellforge_doctor"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--manifest=")
    two_word_flags+=("--manifest")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--manifest")
    local_nonpersistent_flags+=("--manifest=")
    local_nonpersistent_flags+=("-m")
    flags+=("--os=")
    two_word_flags+=("--os")
    local_nonpersistent_flags+=("--os")
    local_nonpersistent_flags+=("--os=")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_list()
{
    last_command="gz_shellforge_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-dir=")
    two_word_flags+=("--config-dir")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config-dir")
    local_nonpersistent_flags+=("--config-dir=")
    local_nonpersistent_flags+=("-c")
    flags+=("--filter=")
    two_word_flags+=("--filter")
    two_word_flags+=("-F")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("--filter=")
    local_nonpersistent_flags+=("-F")
    flags+=("--manifest=")
    two_word_flags+=("--manifest")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--manifest")
    local_nonpersistent_flags+=("--manifest=")
    local_nonpersistent_flags+=("-m")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_migrate()
{
    last_command="gz_shellforge_migrate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--manifest=")
    two_word_flags+=("--manifest")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--manifest")
    local_nonpersistent_flags+=("--manifest=")
    local_nonpersistent_flags+=("-m")
    flags+=("--output-dir=")
    two_word_flags+=("--output-dir")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output-dir")
    local_nonpersistent_flags+=("--output-dir=")
    local_nonpersistent_flags+=("-o")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_profiles_check()
{
    last_command="gz_shellforge_profiles_check"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--data-dir=")
    two_word_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_profiles_list()
{
    last_command="gz_shellforge_profiles_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--data-dir=")
    two_word_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir=")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_profiles_show()
{
    last_command="gz_shellforge_profiles_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--data-dir=")
    two_word_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir")
    local_nonpersistent_flags+=("--data-dir=")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_profiles()
{
    last_command="gz_shellforge_profiles"

    command_aliases=()

    commands=()
    commands+=("check")
    commands+=("list")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_restore()
{
    last_command="gz_shellforge_restore"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup-dir=")
    two_word_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir")
    local_nonpersistent_flags+=("--backup-dir=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--no-git")
    local_nonpersistent_flags+=("--no-git")
    flags+=("--snapshot=")
    two_word_flags+=("--snapshot")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--snapshot")
    local_nonpersistent_flags+=("--snapshot=")
    local_nonpersistent_flags+=("-s")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_flag+=("--file=")
    must_have_one_flag+=("-f")
    must_have_one_flag+=("--snapshot=")
    must_have_one_flag+=("-s")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_template_generate()
{
    last_command="gz_shellforge_template_generate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-dir=")
    two_word_flags+=("--config-dir")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config-dir")
    local_nonpersistent_flags+=("--config-dir=")
    local_nonpersistent_flags+=("-c")
    flags+=("--field=")
    two_word_flags+=("--field")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--field")
    local_nonpersistent_flags+=("--field=")
    local_nonpersistent_flags+=("-f")
    flags+=("--requires=")
    two_word_flags+=("--requires")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--requires")
    local_nonpersistent_flags+=("--requires=")
    local_nonpersistent_flags+=("-r")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_template_list()
{
    last_command="gz_shellforge_template_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_template()
{
    last_command="gz_shellforge_template"

    command_aliases=()

    commands=()
    commands+=("generate")
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge_validate()
{
    last_command="gz_shellforge_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--check-prereqs")
    local_nonpersistent_flags+=("--check-prereqs")
    flags+=("--config-dir=")
    two_word_flags+=("--config-dir")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config-dir")
    local_nonpersistent_flags+=("--config-dir=")
    local_nonpersistent_flags+=("-c")
    flags+=("--manifest=")
    two_word_flags+=("--manifest")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--manifest")
    local_nonpersistent_flags+=("--manifest=")
    local_nonpersistent_flags+=("-m")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_shellforge()
{
    last_command="gz_shellforge"

    command_aliases=()

    commands=()
    commands+=("backup")
    commands+=("build")
    commands+=("cleanup")
    commands+=("deploy")
    commands+=("diff")
    commands+=("doctor")
    commands+=("list")
    commands+=("migrate")
    commands+=("profiles")
    commands+=("restore")
    commands+=("template")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_convert()
{
    last_command="gz_synclone_config_convert"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backup")
    local_nonpersistent_flags+=("--backup")
    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--from=")
    two_word_flags+=("--from")
    local_nonpersistent_flags+=("--from")
    local_nonpersistent_flags+=("--from=")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--to=")
    two_word_flags+=("--to")
    local_nonpersistent_flags+=("--to")
    local_nonpersistent_flags+=("--to=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--file=")
    must_have_one_flag+=("-f")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_generate_discover()
{
    last_command="gz_synclone_config_generate_discover"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--depth=")
    two_word_flags+=("--depth")
    local_nonpersistent_flags+=("--depth")
    local_nonpersistent_flags+=("--depth=")
    flags+=("--follow-symlinks")
    local_nonpersistent_flags+=("--follow-symlinks")
    flags+=("--ignore=")
    two_word_flags+=("--ignore")
    local_nonpersistent_flags+=("--ignore")
    local_nonpersistent_flags+=("--ignore=")
    flags+=("--merge-existing")
    local_nonpersistent_flags+=("--merge-existing")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--recursive")
    local_nonpersistent_flags+=("--recursive")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_generate_init()
{
    last_command="gz_synclone_config_generate_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_generate_template()
{
    last_command="gz_synclone_config_generate_template"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--interactive")
    flags+=("-i")
    local_nonpersistent_flags+=("--interactive")
    local_nonpersistent_flags+=("-i")
    flags+=("--list-templates")
    local_nonpersistent_flags+=("--list-templates")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--template=")
    two_word_flags+=("--template")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("--template=")
    local_nonpersistent_flags+=("-t")
    flags+=("--template-dir=")
    two_word_flags+=("--template-dir")
    local_nonpersistent_flags+=("--template-dir")
    local_nonpersistent_flags+=("--template-dir=")
    flags+=("--var=")
    two_word_flags+=("--var")
    local_nonpersistent_flags+=("--var")
    local_nonpersistent_flags+=("--var=")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_generate()
{
    last_command="gz_synclone_config_generate"

    command_aliases=()

    commands=()
    commands+=("discover")
    commands+=("init")
    commands+=("template")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config_validate()
{
    last_command="gz_synclone_config_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--strict")
    local_nonpersistent_flags+=("--strict")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_config()
{
    last_command="gz_synclone_config"

    command_aliases=()

    commands=()
    commands+=("convert")
    commands+=("generate")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_forge()
{
    last_command="gz_synclone_forge"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--base-url=")
    two_word_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url")
    local_nonpersistent_flags+=("--base-url=")
    flags+=("--cleanup-orphans")
    local_nonpersistent_flags+=("--cleanup-orphans")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--exclude-topics=")
    two_word_flags+=("--exclude-topics")
    local_nonpersistent_flags+=("--exclude-topics")
    local_nonpersistent_flags+=("--exclude-topics=")
    flags+=("--include-archived")
    local_nonpersistent_flags+=("--include-archived")
    flags+=("--include-forks")
    local_nonpersistent_flags+=("--include-forks")
    flags+=("--include-private")
    local_nonpersistent_flags+=("--include-private")
    flags+=("--language=")
    two_word_flags+=("--language")
    local_nonpersistent_flags+=("--language")
    local_nonpersistent_flags+=("--language=")
    flags+=("--max-retries=")
    two_word_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries=")
    flags+=("--max-stars=")
    two_word_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars")
    local_nonpersistent_flags+=("--max-stars=")
    flags+=("--min-stars=")
    two_word_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars")
    local_nonpersistent_flags+=("--min-stars=")
    flags+=("--org=")
    two_word_flags+=("--org")
    local_nonpersistent_flags+=("--org")
    local_nonpersistent_flags+=("--org=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--resume")
    local_nonpersistent_flags+=("--resume")
    flags+=("--ssh")
    local_nonpersistent_flags+=("--ssh")
    flags+=("--state-file=")
    two_word_flags+=("--state-file")
    local_nonpersistent_flags+=("--state-file")
    local_nonpersistent_flags+=("--state-file=")
    flags+=("--strategy=")
    two_word_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy=")
    flags+=("--target=")
    two_word_flags+=("--target")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--topics=")
    two_word_flags+=("--topics")
    local_nonpersistent_flags+=("--topics")
    local_nonpersistent_flags+=("--topics=")
    flags+=("--user")
    local_nonpersistent_flags+=("--user")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--org=")
    must_have_one_flag+=("--provider=")
    must_have_one_flag+=("--target=")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_state_clean()
{
    last_command="gz_synclone_state_clean"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    local_nonpersistent_flags+=("--all")
    flags+=("--organization=")
    two_word_flags+=("--organization")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--organization")
    local_nonpersistent_flags+=("--organization=")
    local_nonpersistent_flags+=("-o")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_state_list()
{
    last_command="gz_synclone_state_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_state_show()
{
    last_command="gz_synclone_state_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--organization=")
    two_word_flags+=("--organization")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--organization")
    local_nonpersistent_flags+=("--organization=")
    local_nonpersistent_flags+=("-o")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    local_nonpersistent_flags+=("-p")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--organization=")
    must_have_one_flag+=("-o")
    must_have_one_flag+=("--provider=")
    must_have_one_flag+=("-p")
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_state()
{
    last_command="gz_synclone_state"

    command_aliases=()

    commands=()
    commands+=("clean")
    commands+=("list")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone_validate()
{
    last_command="gz_synclone_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--use-config")
    local_nonpersistent_flags+=("--use-config")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_synclone()
{
    last_command="gz_synclone"

    command_aliases=()

    commands=()
    commands+=("config")
    commands+=("forge")
    commands+=("state")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cleanup-orphans")
    local_nonpersistent_flags+=("--cleanup-orphans")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config")
    local_nonpersistent_flags+=("--config=")
    local_nonpersistent_flags+=("-c")
    flags+=("--max-retries=")
    two_word_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries")
    local_nonpersistent_flags+=("--max-retries=")
    flags+=("--parallel=")
    two_word_flags+=("--parallel")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--parallel")
    local_nonpersistent_flags+=("--parallel=")
    local_nonpersistent_flags+=("-p")
    flags+=("--progress-mode=")
    two_word_flags+=("--progress-mode")
    local_nonpersistent_flags+=("--progress-mode")
    local_nonpersistent_flags+=("--progress-mode=")
    flags+=("--provider=")
    two_word_flags+=("--provider")
    local_nonpersistent_flags+=("--provider")
    local_nonpersistent_flags+=("--provider=")
    flags+=("--resume")
    local_nonpersistent_flags+=("--resume")
    flags+=("--strategy=")
    two_word_flags+=("--strategy")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--strategy")
    local_nonpersistent_flags+=("--strategy=")
    local_nonpersistent_flags+=("-s")
    flags+=("--use-config")
    local_nonpersistent_flags+=("--use-config")
    flags+=("--use-gzh-config")
    local_nonpersistent_flags+=("--use-gzh-config")
    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gz_root_command()
{
    last_command="gz"

    command_aliases=()

    commands=()
    commands+=("dev-env")
    commands+=("git")
    commands+=("git-sync")
    commands+=("ide")
    commands+=("net-env")
    commands+=("profile")
    commands+=("quality")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("q")
        aliashash["q"]="quality"
        command_aliases+=("qual")
        aliashash["qual"]="quality"
    fi
    commands+=("repo-config")
    commands+=("selfupdate")
    commands+=("shellforge")
    commands+=("synclone")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--experimental")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_gz()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __gz_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("gz")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __gz_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_gz gz
else
    complete -o default -o nospace -F __start_gz gz
fi

# ex: ts=4 sw=4 et filetype=sh
