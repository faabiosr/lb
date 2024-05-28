# lb

[![Build Status](https://img.shields.io/github/actions/workflow/status/faabiosr/lb/test.yaml?logo=github&style=flat-square)](https://github.com/faabiosr/lb/actions?query=workflow:test)
[![Codecov branch](https://img.shields.io/codecov/c/github/faabiosr/lb/master.svg?style=flat-square)](https://codecov.io/gh/faabiosr/lb)
[![Go Report Card](https://goreportcard.com/badge/github.com/faabiosr/lb?style=flat-square)](https://goreportcard.com/report/github.com/faabiosr/lb)
[![Release](https://img.shields.io/github/v/release/faabiosr/lb?display_name=tag&style=flat-square)](https://github.com/faabiosr/lb/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/faabiosr/lb/blob/master/LICENSE)

## :tada: Overview
`lb` lets you balance your AWS lambda layers across regions.

## :relaxed: Motivation
Managing AWS lambda layer across regions is difficult, because each new layer deployment will increment the version automatically, and if you need to introduce a new region, the result will be different versions.

## :dart: Installation

### Unix-like

#### Manual installation
```sh
# by default will install into ~/.local/bin folder.
curl -sSL https://raw.githubusercontent.com/faabiosr/lb/main/install.sh | bash 

# install into /usr/local/bin
curl -sSL https://raw.githubusercontent.com/faabiosr/lb/main/install.sh | sudo INSTALL_PATH=/usr/local/bin bash
```

### go
```sh
go install github.com/faabiosr/lb@latest
```

## :gem: Usage

### Verify the versions deployed across regions
```sh
lb verify --regions 'us-east-1,eu-central-1,sa-east-1' my-layer
```

### Bump all regions with the latest version
```sh
lb bump --regions 'us-east-1,eu-central-1,sa-east-1' my-layer
```

## :toolbox: Development

### Requirements

The entire environment is based on Golang, and you need to install the tools below:
- Install [Go](https://golang.org)
- Install [GolangCI-Lint](https://github.com/golangci/golangci-lint#install) - Linter

### Makefile

Please run the make target below to see the provided targets.

```sh
$ make help
```

## :page_with_curl: License

This project is released under the MIT licence. See [LICENSE](https://github.com/faabiosr/lb/blob/master/LICENSE) for more details.
