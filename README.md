<div align="center">

# miactl

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</div>

**miactl is the cli of the mia-platform DevOps Console**

## Install

### Using brew

```sh
brew install mia-platform/tap/miactl
```

### Using Go

This library require golang at version >= 1.13

```sh
go get -u github.com/mia-platform/miactl
```

### Enabling shell autocompletion

miactl provides autocompletion support for Bash, Zsh and Fish, which can save you a lot of typing.

#### Bash

Completion could be generate running the `miactl completion bash` command.
In order to make this completion work, you should have [bash completion](https://github.com/scop/bash-completion)
correctly installed.
You could make the completion works running:
```sh
miactl completion bash >/etc/bash_completion.d/miactl
```

#### Fish

Completion could be generate running the `miactl completion fish` command.

In order to make this completion work, you should run:
```sh
miactl completion fish >~/.config/fish/completions/miactl.fish
```

#### Zsh

Completion could be generate running the `miactl completion zsh` command

In order to make this completion work, add the following to your ~/.zshrc file:

```sh
source <(miactl completion zsh)
```

## Example usage

### Get projects

```sh
miactl get projects --apiKey "your-api-key" --apiCookie "sid=your-sid" --apiBaseUrl "https://console.url/"
```

### Projects help

```sh
miactl help
```

[github-actions]: https://github.com/mia-platform/miactl/actions
[github-actions-svg]: https://github.com/mia-platform/miactl/workflows/Test%20and%20build/badge.svg
[godoc-svg]: https://godoc.org/github.com/mia-platform/miactl?status.svg
[godoc-link]: https://godoc.org/github.com/mia-platform/miactl
[go-report-card]: https://goreportcard.com/badge/github.com/mia-platform/miactl
[go-report-card-link]: https://goreportcard.com/report/github.com/mia-platform/miactl
[semver]: https://semver.org/
