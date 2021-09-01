# Brev Vault Utils

[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/brevdev/vault-utils)](https://github.com/brevdev/vault-utils/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/brevdev/vault-utils.svg)](https://pkg.go.dev/github.com/brevdev/vault-utils)
[![go.mod](https://img.shields.io/github/go-mod/go-version/brevdev/vault-utils)](go.mod)
[![LICENSE](https://img.shields.io/github/license/brevdev/vault-utils)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/brevdev/vault-utils/build)](https://github.com/brevdev/vault-utils/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/brevdev/vault-utils)](https://goreportcard.com/report/github.com/brevdev/vault-utils)
[![Codecov](https://codecov.io/gh/brevdev/vault-utils/branch/main/graph/badge.svg)](https://codecov.io/gh/brevdev/vault-utils)

`Star` this repository if you find it valuable and worth maintaining.

`Watch` this repository to get notified about new releases, issues, etc.

## Features
Restart systemd service on file change

## Usage
`Needs root access`

1. make build
2. `sudo ./dist/vault-utils_linux_amd64/vault-utils -service=${SERVICE} -configPath=${CONFIG}`

## Contributing

Create an issue or a pull request. Ensure you update CHANGELOG.md with changes and additions with each PR.

The release workflow is triggered each time a tag with `v` prefix is pushed.
