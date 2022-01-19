# Installation Guide

Below you can find installation guides for the supported versions.

## Ubuntu Launchpad

The **samba exporter** package for Ubuntu is published on [launchpad](https://launchpad.net/~imker/+archive/ubuntu/samba-exporter-ppa). To install from there do the following commands on any supported Ubuntu version:

```sh
sudo add-apt-repository ppa:imker/samba-exporter-ppa
sudo apt-get update
sudo apt-get install samba-exporter
```

## Debian

The **samba exporter** package for Debian is published on the projects GitHub Page. To install execute the commands shown below as root:

```bash
wget -qO - https://imker25.github.io/samba_exporter/repos/debian/archive.key | sudo apt-key add -
echo "deb https://imker25.github.io/samba_exporter/repos/debian bullseye main" > /etc/apt/sources.list.d/samba-exporter.list
apt-get update
apt-get install samba-exporter
```

**Hint:** Change `bullseye` to `buster` in case you use Debian 10.

## GitHub Releases - For all supported distributions

Install the [latest Release](https://github.com/imker25/samba_exporter/releases/latest) by downloading the debian package according to your distribution and version and installing it. For example:

```sh
wget https://github.com/imker25/samba_exporter/releases/download/1.3.5-pre/samba-exporter_1.3.5-pre.ppa1.debian10_amd64.deb
sudo dpkg --install ./samba-exporter_1.3.5-pre.ppa1.debian10_amd64.deb
```

**Hint:** Link and file name needs to be adapted to the latest release.

It's also possible to download and install pre-releases from the GitHub this way.

For manual installation see the [Developer Guide](../DeveloperDocs/Compile.md).