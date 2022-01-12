# Mirror Registry


This application will allow user to easily install Quay and its required components using a simple CLI tool. The purpose is to provide a registry to hold a mirror of OpenShift images.

## Pre-Requisites

- RHEL 8 or Fedora machine with `podman v3.3`  installed
- Fully qualified domain name for the Quay service (must resolve via DNS, or at least [/etc/hosts](#local-dns-resolution))
- Passwordless `sudo` access on the target host (rootless install tbd)
- Key-based SSH connectivity on the target host (will be set up automatically for local installs, in case of remote hosts see [here](#generate-ssh-keys))
- `make` (only if [compiling](#compile-your-own-installer) your own installer)

## Installation

Download one of the installer package from our [releases](https://github.com/quay/mirror-registry/releases) page:

- offline version (contains all required images to run Quay)
- online version (additional container images to run Quay and Postgres will be downloaded by the installer)

### Running the installer

To install Quay on your local host with your current user account, run the following command:

```console
$ ./mirror-registry install
```
The following flags are also available:

```
--autoApprove           A boolean value that disables interactive prompts. Will automatically delete quayRoot directory on uninstall. This defaults to false.
--initPassword          The password of the init user created during Quay installation.
--quayHostname          The value to set SERVER_HOSTNAME in the Quay config.yaml. This defaults to <targetHostname>:8443.
--quayRoot          -r  The folder where quay persistent data are saved. This defaults to /etc/quay-install.
--ssh-key           -k  The path of your ssh identity key. This defaults to ~/.ssh/quay_installer.
--sslCert               The path to the SSL certificate Quay should use.
--sslCheckSkip          Whether or not to check the certificate hostname against the SERVER_HOSTNAME in config.yaml.
--sslKey                The path to the SSL key.
--targetHostname    -H  The hostname of the target you wish to install Quay to. This defaults to $HOST.
--targetUsername    -u  The user on the target host which will be used for SSH. This defaults to $USER
--verbose           -v  Show debug logs and ansible playbook outputs
```

**Note**: You may need to modify the value for `--quayHostname` in case the public DNS name of your system is different from its local hostname.

**Note** If you do not supply `--sslCert` and `--sslKey`, these will be autogenerated and made available on that target host under the `{quayRoot}/quay-rootCA` directory.

### Installing on a Remote Host

You can provide your ssh private key to the installer CLI with the `--ssh-key` flag.

To install Quay on a remote host, run the following command:

```console
$ ./mirror-registry install -v --targetHostname some.remote.host.com --targetUsername someuser -k ~/.ssh/my_ssh_key --quayHostname some.remote.host.com
```

Behind the scenes, Ansible is using `ssh -i ~/.ssh/my_ssh_key someuser@some.remote.host.com` as the target to run its playbooks.

### What does the installer do?

This command will make the following changes to your machine

- Generate trusted SSH keys, if not supplied, in case the deployment target is the local host (required since the installer is ansible-based)
- Pulls Quay, Redis, and Postgres images from `registry.redhat.io` (if using online installer)
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates the folder defined by `--quayRoot` (default: `/etc/quay-install`) contains install files, local storage, and config bundle.
- Installs Quay and creates an initial user called `init` with an auto-generated password
- Access credentials are printed at the end of the install routine

## Access Quay

Once installed, the Quay console will be accessible at `https://<quayhostname>:8443`. **Refer to the output of the install process to retrieve user name and password.**

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
$ sudo ./mirror-registry uninstall -v
```

To uninstall Quay from a remote host, run the following command:

```console
$ ./mirror-registry uninstall -v --targetHostname some.remote.host.com --targetUsername someuser -k ~/.ssh/my_ssh_key
```

**Note**: If Quay has been installed with `--quayRoot` the same option needs to be specified at uninstall.

## Local DNS resolution

In case the target host does not have a resolvable DNS record, you can rely on the default host name called `quay` and add the following line to your host machine's `/etc/hosts` file:

```
<targetHostname ip>   quay
```

## Generate SSH Keys

In order to run the installation playbooks, you must have password-less SSH access in place. Local installations will automatically generate the SSH keys for you.

**Note** Passwordless ssh from `root` account is blocked by default on OpenSSH. 

To generate your own SSH keys to install on a remote host, run the following commands.

```console
$ ssh-keygen
$ ssh-add
$ ssh-copy-id <targetHostname>
```
## Compile your own installer

To compile the `mirror-registry.tar.gz` for distribution you need only `podman` and `make` installed.

**NOTE:** The build process pulls images from registry.redhat.io, you may need to run `sudo podman login registry.redhat.io` before starting the build.

You can build the installer running the following command:

```console
$ git clone https://github.com/quay/mirror-registry.git
$ cd mirror-registry
$ make build-online-zip # OR make build-offline-zip
```

This will generate a `mirror-registry.tar.gz` which contains the `mirror-registry` binary, the `image-archive.tar` and the `execution-environment.tar` (if using offline installer). These archives contain all images required to set up Quay.

Once generated, you may untar this file on your desired host machine for installation. You may use the following command:

```console
mkdir mirror-registry
tar -xzvf mirror-registry.tar.gz -C mirror-registry
```

**NOTE** This command may take some time to complete depending on host resources. 
