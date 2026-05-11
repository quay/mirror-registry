#!/usr/bin/env bash
# P2-6: Deep validation of the /health/instance JSON response.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Health Endpoint Deep Validation ==="

# Install
log_info "Installing..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Fetch health endpoint
response=$(curl -sk "https://${QUAY_ENDPOINT}/health/instance" 2>/dev/null)

if [[ -z "${response}" ]]; then
    log_error "FAIL: Health endpoint returned empty response"
    ((FAIL_COUNT++))
else
    log_info "PASS: Health endpoint returned response"
    ((PASS_COUNT++))
fi

# Validate JSON structure
python3 -c "
import sys, json

data = json.loads('''${response}''')

# Check top-level structure
assert 'data' in data, 'Missing data field'
assert 'services' in data['data'], 'Missing services field'

services = data['data']['services']
print(f'Found {len(services)} services: {list(services.keys())}')

# Check all services are healthy
all_healthy = True
for svc, status in services.items():
    if status:
        print(f'  {svc}: healthy')
    else:
        print(f'  {svc}: UNHEALTHY')
        all_healthy = False

if all_healthy:
    print('ALL SERVICES HEALTHY')
    sys.exit(0)
else:
    print('SOME SERVICES UNHEALTHY')
    sys.exit(1)
" && {
    log_info "PASS: All health services report healthy"
    ((PASS_COUNT++))
} || {
    log_error "FAIL: Health endpoint validation failed"
    ((FAIL_COUNT++))
}

# Verify response has expected content type behavior
http_code=$(curl -sk -o /dev/null -w "%{http_code}" "https://${QUAY_ENDPOINT}/health/instance" 2>/dev/null)
if [[ "${http_code}" == "200" ]]; then
    log_info "PASS: Health endpoint returns HTTP 200"
    ((PASS_COUNT++))
else
    log_error "FAIL: Health endpoint returned HTTP ${http_code}"
    ((FAIL_COUNT++))
fi

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

print_summary
