# Setup

## Installation

`miactl` can be installed in different ways, you can choose the one that better fits your needs and the operating system
you are using:

- [Linux and MacOs](#linux-and-macos)
  - [Homebrew](#homebrew)
  - [Go](#go)
  - [Binary Download](#binary-download)
  - [Docker](#docker)
- [Windows (with WSL)](#windows)

### Linux and MacOs

#### Homebrew

If you have [Homebrew] installed on your system `miactl` is only a command away:

```sh
brew install mia-platform/tap/miactl
```

#### Go

If you have [Golang] installed with a version >= 1.13 in your system and you have the `$GOPATH`env set, you can
install `miactl` like this:

```sh
go install github.com/mia-platform/miactl/cmd/miactl@v0.20.0
```

Or like this if the `install` command is not available

```sh
go get -u github.com/mia-platform/miactl/cmd/miactl@0.17.0
```

#### Binary Download

You can install `miactl` with the use of `curl` or `wget` and downloading the latest packages available on GitHub
choosing the correct platform and operating system:

```sh
curl -fsSL --proto '=https' --tlsv1.2 https://github.com/mia-platform/miactl/releases/download/v0.20.0/miactl-linux-amd64 -o /tmp/miactl
```

```sh
wget -q --https-only --secure-protocol=TLSv1_2 https://github.com/mia-platform/miactl/releases/download/v0.20.0/miactl-linux-amd64 -O /tmp/miactl
```

After you have downloaded the file you can validate it against the checksum you can find at this [url] running the
command:

```sh
sha256sum /tmp/miactl
```

After you have validated that the downloaded file is correct, move the binary in your `/usr/local/bin` folder

```sh
chmod +x /tmp/miactl
mv /tmp/miactl /usr/local/bin
```

If the `mv` command replies with a `Permission denied`, retry it as root:

```sh
sudo mv /tmp/miactl /usr/local/bin
```

#### Docker

If you want to run the cli in its environment or you want to test the cli you can use the Docker image:

```sh
docker run ghcr.io/mia-platform/miactl:v0.20.0 miactl
```

### Windows

`miactl` is not directly compatible with Windows, even if you have Go installed:
compilation on this OS is not possible due to current technical restrictions.

However, it is still possible to use `miactl` with Windows Subsystem for Linux (WSL), as explained here below.

#### Installation of WSL

If you don't have WSL on your system, follow the [official guide] to get it.

Once WSL is installed, to open a Linux bash terminal, press Start+R, enter `bash` in the text box and press OK.

#### Install `miactl`

You can now install miactl with any of the methods explained above for Linux,
we suggest the [binary installation](#binary-download) since it's the most straightforward.

#### Setup a service account

Due to some technical restriction, it is not possible to login with a browser when using WSL.
For this reason, we need to [setup a service account](/development_suite/identity-and-access-management/manage-service-accounts.md#service-account-authentication).

Once you have created it, you need to use the [`miactl context auth` command](./30_commands.md#auth) to setup
authentication.

You are now ready to use `miactl`.

## Shell Autocompletion

Once you have installed the cli in your system you can setup the commands completion for one of this shells:

- [`bash`](#bash)
- [`zsh`](#zsh)
- [`fish`](#fish)

When you update the command remember to relaunch the command for your shell to update the completion definition
and get the latest command and/or flags that has been added.

### `bash`

The `bash` autocompletion needs the [`bash-completion`] package installed on your system.

**Warning:** for working correctly you need the `bash-completion` V2 that is compatible only with Bash 4.1+,
please be sure to have the correct versions installed on your system before running the command.

```sh
miactl completion bash >/usr/local/etc/bash_completion.d/miactl
```

After done this you must restart your shell environment or launch `exec bash` for reloading the configurations
and enable the autocompletion.

### `zsh`

For setting up the `zsh` completion, you must enable it. You can use the following command:

```sh
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

Or use something like [`oh-my-zsh`] that will enable it by default. Once you have done it you can launch the
following command to create the file needed by `zsh`:

```sh
miactl completion zsh > /usr/local/share/zsh/site-functions/_miactl
```

After done this you must restart your shell environment or launch `exec zsh` for reloading the configurations and
enable the autocompletion.

### `fish`

To enable the autocompletion in `fish` you have to run this command:

```sh
miactl completion fish > ~/.config/fish/completions/miactl.fish
```

After done this you must restart your shell environment or launch `exec fish` for reloading the configurations and
enable the autocompletion.

## Setup Authorization Provider

If you are not using the cloud offering of Mia-Platform Console, the on premise installation must be properly configured
to enable the correct functionality of the login via user account.

You must add the following URL `http://localhost:53535/oauth/callback` as a valid callback to your provider for enabling
the cli to correctly receive the login data. If you are not able to do it, please ask one of the administrator of your
chosen provider to do the changes.

If you plan to use the cli only with service accounts you don’t have to do anything because the login flow will happen
only via APIs.

[Homebrew]: https://brew.sh "The Missing Package Manager for macOS (or Linux)"
[Golang]: https://go.dev "Build simple, secure, scalable systems with Go"
[url]: https://github.com/mia-platform/miactl/releases/download/v0.20.0/checksums.txt "miactl checksums"
[`bash-completion`]: https://github.com/scop/bash-completion "Programmable completion functions for bash"
[`oh-my-zsh`]: https://ohmyz.sh "Oh My Zsh is a delightful, open source, community-driven
	framework for managing your Zsh configuration"
[official guide]: https://learn.microsoft.com/en-us/windows/wsl/install "How to install Linux on Windows with WSL"
