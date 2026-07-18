#!/usr/bin/env bash

bootstrap_runner_ssh() {
    local source_dir="${RUNNER_SSH_SOURCE:-/run/runner-ssh-source}"
    local target_dir="${HOME}/.ssh"

    [[ -d "$source_dir" ]] || return 0

    umask 077
    mkdir -p "$target_dir"
    chmod 700 "$target_dir"

    if [[ -f "$source_dir/config" ]]; then
        cp "$source_dir/config" "$target_dir/config"
        chmod 600 "$target_dir/config"
    fi
    if [[ -f "$source_dir/id_ed25519" ]]; then
        cp "$source_dir/id_ed25519" "$target_dir/id_ed25519"
        chmod 600 "$target_dir/id_ed25519"
    fi
    if [[ -f "$source_dir/known_hosts" ]]; then
        cp "$source_dir/known_hosts" "$target_dir/known_hosts"
        chmod 644 "$target_dir/known_hosts"
    fi
}
