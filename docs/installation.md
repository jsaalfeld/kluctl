<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: "Installation"
linkTitle: "Installation"
weight: 20
description: "Installing kluctl."
---
-->

# Installation

## Binaries

The kluctl CLI is available as a binary executable for all major platforms,
the binaries can be downloaded form GitHub
[releases page](https://github.com/kluctl/kluctl/releases).

## Installation with Homebrew

With [Homebrew](https://brew.sh) for macOS and Linux:

```sh
brew install kluctl/tap/kluctl
```

## Installation with Bash

With [Bash](https://www.gnu.org/software/bash/) for macOS and Linux:

```sh
curl -s https://kluctl.io/install.sh | bash
```

The install script does the following:
* attempts to detect your OS
* downloads and unpacks the release tar file in a temporary directory
* copies the kluctl binary to `/usr/local/bin`
* removes the temporary directory

## Build from source

Clone the repository:

```bash
git clone https://github.com/kluctl/kluctl
cd kluctl
```

Build the `kluctl` binary (requires go >= 1.19):

```bash
make build
```

Run the binary:

```bash
./bin/kluctl -h
```


<!-- TODO uncomment when chocolatey support is implemented
## Chocolatey

With [Chocolatey](https://chocolatey.org/) for Windows:

```powershell
choco install kluctl
```

-->

<!-- TODO uncomment this when completion is implemented
To configure your shell to load `kluctl` [bash completions](./cmd/kluctl_completion_bash.md) add to your profile:

```sh
. <(kluctl completion bash)
```

[`zsh`](./cmd/kluctl_completion_zsh.md), [`fish`](./cmd/kluctl_completion_fish.md),
and [`powershell`](./cmd/kluctl_completion_powershell.md)
are also supported with their own sub-commands.

-->

## Container images

A container image with `kluctl` is available on GitHub:

* `ghcr.io/kluctl/kluctl:<version>`
