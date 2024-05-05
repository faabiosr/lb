# LB - (Lambda) Layer Balancer

[![Build Status](https://img.shields.io/github/actions/workflow/status/faabiosr/lb/test.yaml?logo=github&style=flat-square)](https://github.com/faabiosr/lb/actions?query=workflow:test)
[![Codecov branch](https://img.shields.io/codecov/c/github/faabiosr/lb/master.svg?style=flat-square)](https://codecov.io/gh/faabiosr/lb)
[![Go Report Card](https://goreportcard.com/badge/github.com/faabiosr/lb?style=flat-square)](https://goreportcard.com/report/github.com/faabiosr/lb)
[![Release](https://img.shields.io/github/v/release/faabiosr/lb?display_name=tag&style=flat-square)](https://github.com/faabiosr/lb/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/faabiosr/lb/blob/master/LICENSE)

Balance your AWS lambda layers across regions.

## Installation

### Linux (apt)
```sh
curl -LO https://github.com/faabiosr/lb/releases/download/v1.0.0/lb_1.0.0_linux_x86_64.deb
sudo apt install -f ./lb_1.0.0_linux_x86_64.deb
```

## Development

### Requirements

The entire environment is based on Golang, and you need to install the tools below:
- Install [Go](https://golang.org)
- Install [GolangCI-Lint](https://github.com/golangci/golangci-lint#install) - Linter

### Makefile

Please run the make target below to see the provided targets.

```sh
$ make help
```

## License

This project is released under the MIT licence. See [LICENSE](https://github.com/faabiosr/lb/blob/master/LICENSE) for more details.
