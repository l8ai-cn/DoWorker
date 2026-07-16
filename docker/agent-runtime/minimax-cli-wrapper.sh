#!/bin/sh
set -eu

if [ -n "${MINIMAX_API_KEY:-}" ]; then
  config_dir="${MMX_CONFIG_DIR:-${HOME}/.mmx}"
  mkdir -p "$config_dir"
  umask 077
  export MMX_CONFIG_DIR="$config_dir"
  node -e '
    const fs = require("fs");
    const path = require("path");
    const configPath = path.join(process.env.MMX_CONFIG_DIR, "config.json");
    let config = {};
    if (fs.existsSync(configPath)) {
      config = JSON.parse(fs.readFileSync(configPath, "utf8"));
    }
    config.api_key = process.env.MINIMAX_API_KEY;
    fs.writeFileSync(configPath, JSON.stringify(config) + "\n", { mode: 0o600 });
  '
fi

exec node /usr/local/lib/node_modules/mmx-cli/dist/mmx.mjs "$@"
