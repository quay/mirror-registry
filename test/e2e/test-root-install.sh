#!/usr/bin/env bash
# P0-8: Test root (sudo) installation.
#
# Installs mirror-registry as root and verifies systemd services are
# created in /etc/systemd/system/ (system scope, not user scope).

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Root (sudo) Install ==="
log_info "Hostname: ${HOSTNAME}"

# Install as root
log_info "Installing as root..."
sudo ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

# Verify systemd services are in system scope (/etc/systemd/system/)
for svc in quay-pod quay-app quay-redis; do
    if [[ -f "/etc/systemd/system/${svc}.service" ]]; then
        log_info "PASS: ${svc}.service in /etc/systemd/system/ (system scope)"
        ((PASS_COUNT++))
    else
        log_error "FAIL: ${svc}.service not found in /etc/systemd/system/"
        ((FAIL_COUNT++))
    fi
done

# Verify services are running under system scope
assert_services_running "true"

# Verify push/pull works
verify_push_pull "${QUAY_ENDPOINT}" "init" "password" "--tls-verify=false"

# Uninstall as root
log_info "Uninstalling as root..."
sudo ${MIRROR_REGISTRY} uninstall --autoApprove -v

assert_clean_uninstall "~/quay-install" "true"

# Verify system-scope unit files are removed
for svc in quay-pod quay-app quay-redis; do
    assert_file_not_exists "${svc}.service removed from /etc/systemd/system/" \
        "/etc/systemd/system/${svc}.service"
done

print_summary
