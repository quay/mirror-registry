# Contributing to mirror-registry

## Setup

```bash
# Requires Go 1.21+ and Ansible
make build

# Test install locally
./mirror-registry install --quayHostname localhost
```

## Development

CLI for air-gapped Quay deployments on RHEL/Fedora.

Wraps:
- Podman/systemd for container runtime
- Ansible playbooks for config management
- Postgres/Redis via containers

## Testing

```bash
# Unit tests
make test

# Integration test (requires VM)
./test/integration.sh

# Manual test on fresh RHEL
vagrant up
vagrant ssh
./mirror-registry install
```

## Pull Requests

- Test on RHEL 8, RHEL 9, Fedora
- Update Ansible playbooks in `ansible-runner/`
- Document new flags in README
- Verify uninstall cleans up fully

## Code Structure

- `cmd/mirror-registry/` - CLI entrypoint
- `ansible-runner/` - embedded Ansible playbooks
- `test/` - integration tests
- `pkg/` - install/upgrade logic
