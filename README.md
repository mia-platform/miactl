# miactl

<center>

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

</center>

`miactl` is the CLI for Mia-Platform Console. It will eventually implement most of the actions you can do
via the UI.

## To Start Using `miactl`

Read more in the [official Mia-Platform documentation](https://docs.mia-platform.eu/docs/products/console/cli/miactl/overview).

Please make sure to install a `miactl` version that supports the Mia-Platform Console you want to use.

You can help yourself with this table:

| Mia-Platform Console version | `miactl` version |
| --- | --- |
| v15.0.0 and above | v0.25.1 and above |
| v14.1.0 and above | v0.24.0 |
| v14.0.0 and before | v0.23.0 |

## To Start Developing `miactl`

To start developing the CLI you must have this requirements:

- golang 1.25+
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
