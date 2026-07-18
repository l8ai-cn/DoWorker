#!/usr/bin/env bash

bootstrap_runner_ssh() {
    local source_dir="${RUNNER_SSH_SOURCE_DIR:-/run/runner-ssh-source}"
    local target_dir="${HOME}/.ssh"
    local runner_owner="${RUNNER_SSH_TARGET_OWNER:-${RUNNER_USER:-runner}:${RUNNER_GROUP:-runner}}"

    [[ -d "$source_dir" ]] || return 0

    umask 077
    rm -rf "$target_dir"
    install -d -m 700 "$target_dir"

    if [[ -f "$source_dir/config" ]]; then
        [[ -r "$source_dir/config" ]] || {
            echo "runner SSH config exists but is not readable: $source_dir/config" >&2
            exit 1
        }
        install -m 600 "$source_dir/config" "$target_dir/config"
    fi
    if [[ -f "$source_dir/id_ed25519" ]]; then
        [[ -r "$source_dir/id_ed25519" ]] || {
            echo "runner SSH private key exists but is not readable: $source_dir/id_ed25519" >&2
            exit 1
        }
        install -m 600 "$source_dir/id_ed25519" "$target_dir/id_ed25519"
    fi
    if [[ -f "$source_dir/id_ed25519.pub" ]]; then
        [[ -r "$source_dir/id_ed25519.pub" ]] || {
            echo "runner SSH public key exists but is not readable: $source_dir/id_ed25519.pub" >&2
            exit 1
        }
        install -m 644 "$source_dir/id_ed25519.pub" "$target_dir/id_ed25519.pub"
    fi
    if [[ -f "$source_dir/known_hosts" ]]; then
        [[ -r "$source_dir/known_hosts" ]] || {
            echo "runner SSH known_hosts exists but is not readable: $source_dir/known_hosts" >&2
            exit 1
        }
        install -m 644 "$source_dir/known_hosts" "$target_dir/known_hosts"
    fi
    if [[ "$(id -u)" = "0" ]] && id "${runner_owner%%:*}" >/dev/null 2>&1; then
        chown -R "$runner_owner" "$target_dir"
    fi
}

run_privileged_bootstrap_then_drop() {
    [[ "$(id -u)" = "0" ]] || return 0
    [[ "$RUNNER_PRIVILEGED_BOOTSTRAP_DONE" != "1" ]] || return 0

    bootstrap_runner_ssh
    chown -R "${RUNNER_USER}:${RUNNER_GROUP}" "${HOME}" /workspace
    export RUNNER_PRIVILEGED_BOOTSTRAP_DONE=1
    exec sudo -E -u "$RUNNER_USER" env HOME="$HOME" RUNNER_PRIVILEGED_BOOTSTRAP_DONE=1 "$0" "$@"
}

generate_runner_cert() {
    echo "▶ generate_runner_cert: CERTS_DIR=$CERTS_DIR SSL_DIR=$SSL_DIR"
    mkdir -p "$CERTS_DIR"
    ls -la "$CERTS_DIR" >&2 || true
    if [[ ! -f "$SSL_DIR/ca.crt" || ! -f "$SSL_DIR/ca.key" ]]; then
        echo "✗ CA 证书未找到: $SSL_DIR" >&2
        ls -la "$SSL_DIR" >&2 || true
        exit 1
    fi
    if [[ -s "$CERTS_DIR/runner.crt" && -s "$CERTS_DIR/runner.key" ]] \
        && openssl verify -CAfile "$SSL_DIR/ca.crt" "$CERTS_DIR/runner.crt" >/dev/null 2>&1; then
        cp "$SSL_DIR/ca.crt" "$CERTS_DIR/ca.crt"
        echo "✓ Runner 证书已存在"
        return 0
    fi
    rm -f "$CERTS_DIR/runner.crt" "$CERTS_DIR/runner.key" "$CERTS_DIR/ca.crt" \
          "$CERTS_DIR/ca.srl" "$CERTS_DIR/runner.csr" "$CERTS_DIR/runner_ext.cnf"
    echo "生成 Runner 客户端证书..."
    openssl genrsa -out "$CERTS_DIR/runner.key" 2048
    openssl req -new -key "$CERTS_DIR/runner.key" \
        -out "$CERTS_DIR/runner.csr" \
        -subj "/CN=${RUNNER_NODE_ID}/O=${RUNNER_ORG_SLUG}/OU=Runner"
    cat > "$CERTS_DIR/runner_ext.cnf" << 'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF
    openssl x509 -req -days 365 \
        -in "$CERTS_DIR/runner.csr" \
        -CA "$SSL_DIR/ca.crt" -CAkey "$SSL_DIR/ca.key" \
        -CAserial "$CERTS_DIR/ca.srl" -CAcreateserial \
        -out "$CERTS_DIR/runner.crt" \
        -extfile "$CERTS_DIR/runner_ext.cnf"
    cp "$SSL_DIR/ca.crt" "$CERTS_DIR/ca.crt"
    rm -f "$CERTS_DIR/runner.csr" "$CERTS_DIR/runner_ext.cnf"
    chmod 600 "$CERTS_DIR/runner.key"
    chmod 644 "$CERTS_DIR/runner.crt" "$CERTS_DIR/ca.crt"
    echo "✓ Runner 证书生成完成"
}

init_claude_config() {
    local claude_dir="${HOME}/.claude"
    local claude_actual="${claude_dir}/claude.json"
    local claude_link="${HOME}/.claude.json"
    mkdir -p "$claude_dir"
    [[ -f "$claude_actual" ]] || printf '%s\n' \
        '{"hasCompletedOnboarding":true,"theme":"dark","autoUpdaterStatus":"disabled","shiftEnterKeyBindingInstalled":true}' \
        > "$claude_actual"
    if [[ ! -L "$claude_link" ]]; then
        rm -f "$claude_link"
        ln -s "$claude_actual" "$claude_link"
    fi
    [[ -f "$claude_dir/settings.json" ]] || printf '%s\n' \
        '{"permissions":{"allow":["Bash(*)","Read(*)","Write(*)","Edit(*)","Glob(*)","Grep(*)","WebFetch(*)","WebSearch(*)"],"deny":[]},"autoUpdaterStatus":"disabled","spinnerTipsEnabled":false}' \
        > "$claude_dir/settings.json"
}

init_codex_config() {
    mkdir -p "${HOME}/.codex"
    [[ -f "${HOME}/.codex/config.toml" ]] || cat > "${HOME}/.codex/config.toml" << 'EOF'
model = "gpt-4.1"
approval_policy = "never"
sandbox_mode = "danger-full-access"

[shell_environment_policy]
inherit = "all"
EOF
}

init_gemini_config() {
    mkdir -p "${HOME}/.gemini"
    [[ -f "${HOME}/.gemini/settings.json" ]] || printf '%s\n' \
        '{"coreTools":["read_file","edit_file","write_file","run_shell_command","search_files","list_directory","web_search"],"excludeTools":[],"theme":"Default (Dark)","checkForUpdates":false,"sandbox":false,"yolo":false}' \
        > "${HOME}/.gemini/settings.json"
}

init_do_agent_config() {
    local settings="${HOME}/.agent/settings.json"
    mkdir -p "${HOME}/.agent"
    [[ -f "$settings" ]] || printf '%s\n' '{"model":"minimax/MiniMax-M3"}' > "$settings"
}

init_ai_cli_configs() {
    case "${AGENT_RUNTIME}" in
        claude-code) init_claude_config ;;
        codex-cli|video-studio) init_codex_config ;;
        gemini-cli) init_gemini_config ;;
        minimax-cli) mkdir -p "${HOME}/.minimax" ;;
        do-agent) init_do_agent_config ;;
        grok-build) mkdir -p "${HOME}/.grok" ;;
        openclaw) mkdir -p "${HOME}/.openclaw" ;;
        hermes) mkdir -p "${HOME}/.hermes" ;;
        e2e-echo|loopal|aider|opencode) ;;
    esac
}
