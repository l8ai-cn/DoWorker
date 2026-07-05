# shellcheck shell=bash
# lifecycle_launch.sh — detach long-running dev servers from dev.sh exit.

_launch_setsid() {
    local name="$1" log_file="$2"
    shift 2
    local rt_dir="$SCRIPT_DIR/runtime/frontends"
    mkdir -p "$rt_dir"
    local pgid_file="$rt_dir/$name.pgid"

    if [[ -f "$pgid_file" ]]; then
        local old_pgid
        old_pgid=$(cat "$pgid_file" 2>/dev/null || true)
        if [[ -n "$old_pgid" ]]; then
            kill -TERM -- "-$old_pgid" 2>/dev/null || true
            sleep 1
            kill -KILL -- "-$old_pgid" 2>/dev/null || true
        fi
        rm -f "$pgid_file"
    fi

    python3 -c "import os, sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])" \
        "$@" >"$log_file" 2>&1 &
    echo $! >"$pgid_file"
}

_stop_setsid() {
    local name="$1"
    local pgid_file="$SCRIPT_DIR/runtime/frontends/$name.pgid"
    [[ -f "$pgid_file" ]] || return 0
    local pgid
    pgid=$(cat "$pgid_file" 2>/dev/null || true)
    if [[ -n "$pgid" ]]; then
        kill -TERM -- "-$pgid" 2>/dev/null || true
        sleep 1
        kill -KILL -- "-$pgid" 2>/dev/null || true
    fi
    rm -f "$pgid_file"
}
