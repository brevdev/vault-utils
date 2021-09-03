# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/brevdev/vault-utils/compare/v0.2.0...HEAD)

## [0.2.0](https://github.com/brevdev/vault-utils/releases/tag/v0.2.0)

### Added 

- Added proper exit codes if error instead of panics.
- Added more robust validation messages

### Changed

- Swapping fsnotify with custom hash based poller since mounted volume updates do not trigger fsnotify
- Replacing golint with revive
- Replaced template readme with vault-utils readme

### Removed

- CI builds with mac and windows

## [0.1.0](https://github.com/brevdev/vault-utils/releases/tag/v0.1.0)

### Added 

- Added systemd restarter on config file change
- Added throttler to not restart more than once a second