# OpenShift Mirror Registry

This application will allow user to easily install Quay and its required components using a simple CLI tool. The purpose is to provide a registry to hold a mirror of OpenShift images.
## Pre-Requisites

- RHEL 8 or Fedora machine with `podman` installed
- fully qualified domain name for the Quay service (must resolve via DNS, or at least [/etc/hosts](#local-dns-resolution))
- passwordless `sudo` access on the target host (rootless install tbd)
- key-based SSH connectivity on the target host (will be set up automatically for local installs, in case of remote hosts see [here](#generate-ssh-keys))
- `make` (only if [compiling](#compile-your-own-installer) your own installer)
## Installation

Download one of the installer package from our [releases](https://github.com/quay/openshift-mirror-registry/releases) page:

- online version (additional container images to run Quay and Postgres will be downloaded by the installer)
- offline version (contains all required images to run Quay)
### Running the installer

To install Quay on your local host with your current user account, run the following command:

```console
$ ./openshift-mirror-registry install
```
The following flags are also available:

```
--ssh-key           -k  The path of your ssh identity key. This defaults to ~/.ssh/quay_installer
--targetHostname    -H  The hostname of the target you wish to install Quay to. This defaults to $HOST".
--targetUsername    -u  The user on the target host which will be used for SSH. This defaults to $USER
--quayHostname          The value to set SERVER_HOSTNAME in the Quay config.yaml. This defaults to <targetHostname>:8443
--initPassword          The password of the init user created during Quay installation.
--quayRoot          -r  The folder where quay persistent data are saved. This defaults to /etc/quay-install
--verbose           -v  Show debug logs and ansible playbook outputs
```

In particular, modify the value for `--quayHostname` in case the public DNS name of your system is different from its local hostname.

### Installing on a Remote Host

You can provide your ssh private key to the installer CLI with the --ssh-key flag.

To install Quay on a remote host, run the following command:

```console
$ ./openshift-mirror-registry install -v --targetHostname some.remote.host.com --targetUsername someuser -k ~/.ssh/my_ssh_key --quayHostname some.remote.host.com
```

Behind the scenes, Ansible is using `ssh -i ~/.ssh/my_ssh_key someuser@some.remote.host.com` as the target to run its playbooks.

### What does the installer do?

This command will make the following changes to your machine

- generate trusted SSH keys in case the deployment target is the local host (required since the installer is ansible-based)
- Pulls Quay, Redis, and Postgres containers from quay.io (if using online installer)
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates the folder defined by `--quayRoot` (default: `/etc/quay-install`) contains install files, local storage, and config bundle.
- Installs Quay and creates an initial user called `init` with an auto-generated password
- Access credentials are printed at the end of the install routine

## Access Quay

Once installed, the Quay console will be accessible at `https://<quayhostname>:8443`. Refer to the output of the install process to retrieve user name and password.

You can then log into the registry using the provided credentials, for example:

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

If Quay has been installed with `--quayRoot` the same option needs to be used during the uninstall.

This command will delete the permanent data directory and disable all systemd services set up by Quay.


## Local DNS resolution

In case the target host does not have a resolvable DNS record, you can rely on the default host name called `quay` and ddd the following line to your host machine's `/etc/hosts` file:

```
<targetHostname ip>   quay
```

## Generate SSH Keys

In order to run the installation playbooks, you must have password-less SSH access in place. Local installations will automatically generate the SSH keys for you.

To generate your own SSH keys to install on a remote host, run the following commands.

```console
$ ssh-keygen
$ ssh-add
$ ssh-copy-id <targetHostname>
```
## Compile your own installer

To compile the openshift-mirror-registry.tar.gz for distribution you need `ansible-runner` installed.

You can build the installer running the following command:

```console
$ git clone https://github.com/quay/openshift-mirror-registry.git
$ cd openshift-mirror-registry
$ make build-online-zip # OR make build-offline-zip
```

**NOTE:** the build process pulls an image from registry.redhat.io, you may need to run `sudo podman login registry.redhat.io` before starting the build. 

This will generate a `openshift-mirror-registry.tar.gz` which contains this README.md, the `openshift-mirror-registry` binary, and the `image-archive.tar` (if using offline installer) which contains the images required to set up Quay.

Once generated, you may untar this file on your desired host machine for installation. You may use the following command:

```console
mkdir openshift-mirror-registry
tar -xzvf openshift-mirror-registry.tar.gz -C openshift-mirror-registry
```

NOTE - With the offline version, this may take some time.
