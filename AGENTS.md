# Mirror Registry - Developer Agent Context

## Project Overview

Mirror Registry is a CLI tool for installing Quay container registry on RHEL/Fedora hosts. It provides a simple way to set up a local registry for mirroring OpenShift container images, particularly for disconnected/air-gapped environments.

**Repository**: https://github.com/quay/mirror-registry

## Tech Stack

- **Language**: Go 1.21+ (Cobra CLI)
- **Deployment**: Ansible (via ansible-runner)
- **Containers**: Podman, systemd user services
- **Target OS**: RHEL, Fedora

## Core Commands

```bash
# Build installers
make build-online-zip   # Online installer
make build-offline-zip  # Offline installer (air-gapped)

# CLI usage
./mirror-registry install [flags]
./mirror-registry upgrade [flags]
./mirror-registry uninstall [flags]
```

## Documentation Map

Read the specific documentation below when your task involves these keywords:

- **Architecture, CLI, Go, Cobra, Ansible, Roles, Code, Structure** → `read_file agent_docs/architecture.md`
  - Project structure, Go CLI commands, Ansible playbook flow

- **Release, Build, Tags, Versioning, Images** → `read_file agent_docs/ops.md`
  - Build process, release workflow, image configuration

- **Testing, CI/CD, GitHub Actions, Integration Tests, Vagrant** → `read_file agent_docs/testing.md`
  - CI pipeline, integration test scenarios, test infrastructure

## Key Files

| Purpose | Path |
|---------|------|
| CLI entry point | `main.go` |
| CLI commands | `cmd/*.go` |
| Image versions | `.env` |
| Build targets | `Makefile` |
| CI workflow | `.github/workflows/jobs.yml` |
| Ansible playbooks | `ansible-runner/context/app/project/*.yml` |
| Ansible role | `ansible-runner/context/app/project/roles/mirror_appliance/` |

## Universal Conventions

- **Go style**: Standard library preferred, idiomatic Go patterns
- **Ansible**: Follow existing role structure in `mirror_appliance`
- **Testing**: PRs require `ok-to-test` label for CI integration tests
- **Safety**: Never commit credentials or secrets
