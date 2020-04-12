<div align="center">

# miactl

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</div>

**miactl is the cli of the mia-platform DevOps Console**

## Install

This library require golang at version >= 1.13

```sh
go get -u github.com/mia-platform/miactl
```

## Example usage

### Get projects

```sh
miactl get projects --secret "your-secret" --apiCookie "sid=your-sid" --apiBaseUrl "https://console.url/"
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
