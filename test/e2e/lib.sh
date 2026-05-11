#!/usr/bin/env bash
# Shared test helper library for mirror-registry e2e tests.
# Source this file from individual test scripts: source "$(dirname "$0")/lib.sh"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PASS_COUNT=0
FAIL_COUNT=0

# Colors (disabled if NO_COLOR is set)
if [[ -z "${NO_COLOR:-}" ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    GREEN='' RED='' YELLOW='' NC=''
fi

log_info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
log_error() { echo -e "${RED}[FAIL]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }

assert_success() {
    local description="$1"
    shift
    if "$@"; then
        log_info "PASS: ${description}"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${description}"
        log_error "  Command: $*"
        ((FAIL_COUNT++))
    fi
}

assert_failure() {
    local description="$1"
    shift
    if "$@" 2>/dev/null; then
        log_error "FAIL: ${description} (expected failure but succeeded)"
        ((FAIL_COUNT++))
    else
        log_info "PASS: ${description}"
        ((PASS_COUNT++))
    fi
}

assert_contains() {
    local description="$1"
    local haystack="$2"
    local needle="$3"
    if echo "${haystack}" | grep -q "${needle}"; then
        log_info "PASS: ${description}"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${description}"
        log_error "  Expected to find: ${needle}"
        ((FAIL_COUNT++))
    fi
}

assert_file_exists() {
    local description="$1"
    local filepath="$2"
    if [[ -e "${filepath}" ]]; then
        log_info "PASS: ${description}"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${description}"
        log_error "  File not found: ${filepath}"
        ((FAIL_COUNT++))
    fi
}

assert_file_not_exists() {
    local description="$1"
    local filepath="$2"
    if [[ ! -e "${filepath}" ]]; then
        log_info "PASS: ${description}"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${description}"
        log_error "  File should not exist: ${filepath}"
        ((FAIL_COUNT++))
    fi
}

# Wait for Quay health endpoint to become healthy (up to timeout seconds)
wait_for_quay() {
    local hostname="${1:-localhost:8443}"
    local timeout="${2:-120}"
    local elapsed=0
    log_info "Waiting for Quay at ${hostname} to become healthy (timeout: ${timeout}s)..."
    while [[ ${elapsed} -lt ${timeout} ]]; do
        if curl -sk "https://${hostname}/health/instance" 2>/dev/null | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    services = data.get('data', {}).get('services', {})
    if services and all(services.values()):
        sys.exit(0)
except:
    pass
sys.exit(1)
" 2>/dev/null; then
            log_info "Quay is healthy after ${elapsed}s"
            return 0
        fi
        sleep 5
        ((elapsed+=5))
    done
    log_error "Quay did not become healthy within ${timeout}s"
    return 1
}

assert_quay_healthy() {
    local hostname="${1:-localhost:8443}"
    local response
    response=$(curl -sk "https://${hostname}/health/instance" 2>/dev/null)
    if [[ -z "${response}" ]]; then
        log_error "FAIL: Health endpoint returned empty response"
        ((FAIL_COUNT++))
        return 1
    fi

    if echo "${response}" | python3 -c "
import sys, json
data = json.load(sys.stdin)
services = data.get('data', {}).get('services', {})
assert services, 'No services in health response'
for svc, status in services.items():
    assert status, f'Service {svc} is not healthy'
print('All services healthy')
" 2>/dev/null; then
        log_info "PASS: Quay health check"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Quay health check"
        log_error "  Response: ${response}"
        ((FAIL_COUNT++))
    fi
}

assert_services_running() {
    local use_sudo="${1:-false}"
    local systemctl_cmd="systemctl --user"
    if [[ "${use_sudo}" == "true" ]]; then
        systemctl_cmd="sudo systemctl"
    fi

    for svc in quay-pod quay-app quay-redis; do
        if ${systemctl_cmd} is-active "${svc}.service" &>/dev/null; then
            log_info "PASS: ${svc}.service is active"
            ((PASS_COUNT++))
        else
            log_error "FAIL: ${svc}.service is not active"
            ((FAIL_COUNT++))
        fi
    done
}

assert_clean_uninstall() {
    local quay_root="${1:-~/quay-install}"
    local use_sudo="${2:-false}"
    local systemctl_cmd="systemctl --user"
    if [[ "${use_sudo}" == "true" ]]; then
        systemctl_cmd="sudo systemctl"
    fi

    # No quay containers running
    local containers
    containers=$(podman ps -q -f name=quay 2>/dev/null || true)
    if [[ -z "${containers}" ]]; then
        log_info "PASS: No quay containers running"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Quay containers still running: ${containers}"
        ((FAIL_COUNT++))
    fi

    # No systemd services enabled
    for svc in quay-app quay-pod quay-redis; do
        if ! ${systemctl_cmd} is-enabled "${svc}.service" &>/dev/null; then
            log_info "PASS: ${svc}.service is not enabled"
            ((PASS_COUNT++))
        else
            log_error "FAIL: ${svc}.service is still enabled"
            ((FAIL_COUNT++))
        fi
    done

    # quay-install directory removed
    eval local expanded_root="${quay_root}"
    if [[ ! -d "${expanded_root}" ]]; then
        log_info "PASS: ${quay_root} directory removed"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${quay_root} directory still exists"
        ((FAIL_COUNT++))
    fi
}

# Generate a self-signed CA and server certificate for testing
generate_test_certs() {
    local cert_dir="$1"
    local hostname="$2"

    mkdir -p "${cert_dir}"

    # Generate CA key and cert
    openssl genrsa -out "${cert_dir}/ca.key" 2048 2>/dev/null
    openssl req -x509 -new -nodes \
        -key "${cert_dir}/ca.key" \
        -sha256 -days 1 \
        -out "${cert_dir}/ca.pem" \
        -subj "/CN=Test CA" 2>/dev/null

    # Generate server key
    openssl genrsa -out "${cert_dir}/server.key" 2048 2>/dev/null

    # Generate server CSR
    openssl req -new \
        -key "${cert_dir}/server.key" \
        -out "${cert_dir}/server.csr" \
        -subj "/CN=${hostname}" 2>/dev/null

    # Create extensions file for SAN
    cat > "${cert_dir}/ext.cnf" <<EXTEOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = DNS:${hostname}
EXTEOF

    # Sign server cert with CA
    openssl x509 -req \
        -in "${cert_dir}/server.csr" \
        -CA "${cert_dir}/ca.pem" \
        -CAkey "${cert_dir}/ca.key" \
        -CAcreateserial \
        -out "${cert_dir}/server.cert" \
        -days 1 \
        -sha256 \
        -extfile "${cert_dir}/ext.cnf" 2>/dev/null

    log_info "Generated test certificates for ${hostname} in ${cert_dir}"
}

# Push and pull a busybox image to verify registry functionality
verify_push_pull() {
    local hostname="${1:-localhost:8443}"
    local username="${2:-init}"
    local password="${3:-password}"
    local tls_verify="${4:---tls-verify=false}"

    log_info "Testing push/pull to ${hostname} as ${username}..."

    podman login -u "${username}" -p "${password}" "${hostname}" ${tls_verify}
    podman pull docker.io/library/busybox 2>/dev/null || true
    podman tag docker.io/library/busybox "${hostname}/${username}/busybox:test"
    podman push "${hostname}/${username}/busybox:test" ${tls_verify}

    # Remove local copy and pull back
    podman rmi "${hostname}/${username}/busybox:test" 2>/dev/null || true
    podman pull "${hostname}/${username}/busybox:test" ${tls_verify}

    if podman images | grep -q "${hostname}/${username}/busybox"; then
        log_info "PASS: Push/pull verification"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Push/pull verification"
        ((FAIL_COUNT++))
    fi

    # Cleanup
    podman rmi "${hostname}/${username}/busybox:test" 2>/dev/null || true
}

print_summary() {
    echo ""
    echo "=============================="
    echo "Test Summary: ${PASS_COUNT} passed, ${FAIL_COUNT} failed"
    echo "=============================="
    if [[ ${FAIL_COUNT} -gt 0 ]]; then
        return 1
    fi
    return 0
}
