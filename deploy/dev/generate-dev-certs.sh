#!/bin/bash
# Generate self-signed certificates for development gRPC + mTLS
#
# Uses RSA-2048 (Go x509 reliably parses these; some OpenSSL EC outputs
# trigger "invalid ECDSA parameters" in the backend PKI loader).
#
# Usage:
#   ./generate-dev-certs.sh          # generate if missing
#   ./generate-dev-certs.sh --force  # regenerate even if files exist
#
# Output directory: ./ssl/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSL_DIR="${SCRIPT_DIR}/ssl"

force=false
if [[ "${1:-}" == "--force" ]]; then
    force=true
fi

mkdir -p "${SSL_DIR}"

if [[ "$force" == "true" ]]; then
    rm -f "${SSL_DIR}/ca.crt" "${SSL_DIR}/ca.key" "${SSL_DIR}/server.crt" "${SSL_DIR}/server.key"
fi

if [[ -f "${SSL_DIR}/ca.crt" && -f "${SSL_DIR}/server.crt" ]]; then
    echo "Certificates already exist. To regenerate: ./generate-dev-certs.sh --force"
    exit 0
fi

echo "Generating development certificates in ${SSL_DIR}..."

openssl genrsa -out "${SSL_DIR}/ca.key" 2048
openssl req -new -x509 -days 3650 -key "${SSL_DIR}/ca.key" -out "${SSL_DIR}/ca.crt" \
    -subj "/CN=AgentsMesh Dev CA/O=AgentsMesh/OU=Development"

openssl genrsa -out "${SSL_DIR}/server.key" 2048
openssl req -new -key "${SSL_DIR}/server.key" -out "${SSL_DIR}/server.csr" \
    -subj "/CN=localhost/O=AgentsMesh/OU=Backend"

cat > "${SSL_DIR}/server_ext.cnf" << 'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = host.docker.internal
DNS.3 = host.lan
DNS.4 = backend
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

openssl x509 -req -days 3650 -in "${SSL_DIR}/server.csr" \
    -CA "${SSL_DIR}/ca.crt" -CAkey "${SSL_DIR}/ca.key" -CAcreateserial \
    -out "${SSL_DIR}/server.crt" -extfile "${SSL_DIR}/server_ext.cnf"

rm -f "${SSL_DIR}/server.csr" "${SSL_DIR}/server_ext.cnf" "${SSL_DIR}/ca.srl"

chmod 600 "${SSL_DIR}/ca.key" "${SSL_DIR}/server.key"
chmod 644 "${SSL_DIR}/ca.crt" "${SSL_DIR}/server.crt"

echo ""
echo "Development certificates generated (RSA-2048, SAN includes host.docker.internal)."
