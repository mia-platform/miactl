# miactl

<center>

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</center>

`miactl` is the CLI for Mia-Platform Console. It will eventually implement most of the actions you can do
via the UI.

## To Start Using `miactl`

Read the documentation [here](./docs/10_overview.md).

## To Start Developing `miactl`

To start developing the CLI you must have this requirements:

- golang 1.19+
- make

Once you have pulled the code locally, you can build the code with make:

```sh
make build
```

`make` will download all the dependencies needed and will build the binary for your current system that you can find
in the `bin` folder.

To build the docker image locally run:

```sh
make docker-build
```

## Testing `miactl`

To run the tests use the command:

```sh
make test
```

Or add the `DEBUG_TEST` flag to run the test with debug mode enabled:

```sh
make test DEBUG_TEST=1
```

Before sending a PR be sure that all the linter pass with success:

```sh
make lint
```

[github-actions]: https://github.com/mia-platform/miactl/actions
[github-actions-svg]: https://github.com/mia-platform/miactl/workflows/Continuous%20Integration%20Pipeline/badge.svg
[godoc-svg]: https://godoc.org/github.com/mia-platform/miactl?status.svg
[godoc-link]: https://godoc.org/github.com/mia-platform/miactl
[go-report-card]: https://goreportcard.com/badge/github.com/mia-platform/miactl
[go-report-card-link]: https://goreportcard.com/report/github.com/mia-platform/miactl
