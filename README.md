# Quay Installer

This application will allow user to install Quay and its required components using a simple CLI tool.

## Pre-Requisites

- RHEL 8 or Fedora machine with Podman installed
- `sudo` access on desired host (rootless install tbd)
- make (only if compiling using Makefile)
- sshd must be enabled

### Compile

To compile the openshift-mirror-registry.tar.gz for distribution, run the following command:

```console
$ git clone https://github.com/quay/openshift-mirror-registry.git
$ cd openshift-mirror-registry
$ make build-online-zip # OR make build-offline-zip
```

This will generate a `openshift-mirror-registry.tar.gz` which contains this README.md, the `openshift-mirror-registry` binary, and the `image-archive.tar` (if using offline installer) which contains the images required to set up Quay.

Once generated, you may untar this file on your desired host machine for installation. You may use the following command:

```console
mkdir openshift-mirror-registry
tar -xzvf openshift-mirror-registry.tar.gz -C openshift-mirror-registry
```

NOTE - With the offline version, this may take some time.

## Installation

Add the following line to host machine `/etc/hosts` file:

```
<targetHostname ip>   quay
```

### SSH Keys

In order to run the installation playbooks, you must also set up SSH keys. Local installations will automatically generate the SSH keys for you.

To generate your own SSH keys to install on a remote host, run the following commands.

```console
$ ssh-keygen
$ ssh-add
$ ssh-copy-id <targetHostname>
```

You can provide your ssh private key to the installer CLI with the --ssh-key flag.

### Running the installer

To install Quay on localhost, run the following command:

```console
$ ./openshift-mirror-registry install -v
```

The following flags are also available:

```
--ssh-key   -k  The path of your ssh identity key. This defaults to ~/.ssh/quay_installer
--targetHostname  -H  The hostname of the target you wish to install Quay to. This defaults to localhost.
--targetUsername   -u  The user you wish to ssh into your target with. This defaults to $USER
--quayHostname          The hostname to set as SERVER_HOSTNAME in the Quay config.yaml. This defaults to quay:8443
--initPassword          The password of the init user created during Quay installation.
```

#### Installing on a Remote Host

To install Quay on a remote host, run the following command:

```console
$ ./openshift-mirror-registry install -v --targetHostname some.remote.host.com --targetUsername someuser -k ~/.ssh/my_ssh_key
```

Behind the scenes, Ansible is using `ssh -i ~/.ssh/my_ssh_key someuser@some.remote.host.com` as the target to run its playbooks.

#### What does the installer do?

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io (if using online installer)
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `/etc/quay-install` contains install files, local storage, and config bundle.
- Installs Quay and creates an initial user

## Access Quay

Once installed, the Quay console will be accessible at `https://quay:8443`.

You can then log into the registry using the provided credentials.

```console
$ podman login -u init -p <password> --tls-verify=false quay:8443
```

After logging in, you can run commands such as:

```console
$ podman tag docker.io/library/busybox:latest quay:8080/init/busybox:latest
$ podman push quay:8443/init/busybox:latest --tls-verify=false
$ podman pull quay:8443/init/busybox:latest --tls-verify=false
```

## Uninstall

To uninstall Quay from localhost, run the following command:

```console
$ sudo ./openshift-mirror-registry uninstall -v
```

To uninstall Quay from a remote host, run the following command:

```console
$ ./openshift-mirror-registry uninstall -v --targetHostname some.remote.host.com --targetUsername someuser -k ~/.ssh/my_ssh_key
```

This command will delete the `/etc/quay-install` directory and disable all systemd services set up by Quay.
