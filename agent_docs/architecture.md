# Architecture

## Project Structure

```
mirror-registry/
├── cmd/                    # Go CLI commands (Cobra)
│   ├── root.go            # Root command, global flags
│   ├── install.go         # Install command implementation
│   ├── upgrade.go         # Upgrade command implementation
│   ├── uninstall.go       # Uninstall command implementation
│   └── utils.go           # Shared utilities
├── main.go                # Entry point
├── ansible-runner/        # Ansible execution environment
│   └── context/
│       └── app/project/
│           ├── install_mirror_appliance.yml
│           ├── upgrade_mirror_appliance.yml
│           ├── uninstall_mirror_appliance.yml
│           └── roles/mirror_appliance/
├── test/                  # Vagrant-based testing
├── .github/workflows/     # CI/CD
├── .env                   # Container image versions
├── Makefile              # Build targets
├── Dockerfile            # Offline installer build
└── Dockerfile.online     # Online installer build
```

## Go CLI Architecture

Built with Cobra CLI framework. Each command is defined in `cmd/`:

- **root.go**: Defines global flags (verbose, no-color, etc.) and initializes the CLI
- **install.go**: Handles Quay installation on local or remote hosts
- **upgrade.go**: Upgrades existing Quay installations
- **uninstall.go**: Removes Quay and cleans up resources
- **utils.go**: SSH key generation, password generation, Ansible runner invocation

### Build-time Configuration

Image versions are injected via ldflags during build (see Makefile):
- `cmd.quayImage` - Quay container image
- `cmd.redisImage` - Redis container image
- `cmd.eeImage` - Ansible execution environment image
- `cmd.pauseImage` - Pause container image
- `cmd.sqliteImage` - SQLite CLI image
- `cmd.releaseVersion` - Version string

## Ansible Structure

The CLI executes Ansible playbooks via ansible-runner. Playbooks are in `ansible-runner/context/app/project/`:

### Playbooks
- `install_mirror_appliance.yml` - Main installation playbook
- `upgrade_mirror_appliance.yml` - Upgrade playbook
- `uninstall_mirror_appliance.yml` - Uninstall playbook

### Role: mirror_appliance

Located in `roles/mirror_appliance/`:

**Key task files:**
- `main.yaml` - Entry point, includes other tasks
- `install-deps.yaml` - Installs dependencies on target host
- `install-pod-service.yaml` - Sets up Podman pod systemd service
- `install-quay-service.yaml` - Configures Quay container service
- `install-redis-service.yaml` - Configures Redis container service
- `create-init-user.yaml` - Creates initial Quay admin user
- `upgrade.yaml` - Handles upgrade logic
- `uninstall.yaml` - Cleanup and removal
- `migrate.yaml` - Database migration (PostgreSQL to SQLite)

**Templates:**
- `config.yaml.j2` - Quay configuration template
- `pod.service.j2` - Systemd pod service unit
- `quay.service.j2` - Systemd Quay service unit
- `redis.service.j2` - Systemd Redis service unit

## CLI to Ansible Flow

1. User runs `./mirror-registry install [flags]`
2. CLI parses flags and builds Ansible extra vars
3. CLI invokes ansible-runner with the appropriate playbook
4. Ansible connects to target host (local or remote via SSH)
5. Ansible executes the role tasks to configure Quay

## Key Patterns

- **Systemd user services**: All containers run as systemd user services for persistence
- **SSH-based remote execution**: Remote installs use SSH key authentication
- **Idempotent operations**: Ansible tasks are designed to be re-runnable
- **Offline support**: Images can be pre-packaged in the installer tarball
