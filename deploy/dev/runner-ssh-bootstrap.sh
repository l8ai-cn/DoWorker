#!/usr/bin/env bash

bootstrap_runner_ssh() {
    local source_dir="${RUNNER_SSH_SOURCE_DIR:-/run/runner-ssh-source}"
    local target_dir="${HOME}/.ssh"

    [[ -d "$source_dir" ]] || return 0

    umask 077
    rm -rf "$target_dir"
    install -d -m 700 "$target_dir"

    if [[ -f "$source_dir/config" ]]; then
        install -m 600 "$source_dir/config" "$target_dir/config"
    fi
    if [[ -f "$source_dir/id_ed25519" ]]; then
        install -m 600 "$source_dir/id_ed25519" "$target_dir/id_ed25519"
    fi
    if [[ -f "$source_dir/known_hosts" ]]; then
        install -m 644 "$source_dir/known_hosts" "$target_dir/known_hosts"
    fi
}
