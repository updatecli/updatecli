# Updatecli

[![Go Report Card](https://goreportcard.com/badge/github.com/olblak/updatecli)](https://goreportcard.com/report/github.com/olblak/updatecli)

[![Docker Pulls](https://img.shields.io/docker/pulls/olblak/updatecli?label=olblak%2Fupdatecli&logo=docker&logoColor=white)](https://hub.docker.com/r/olblak/updatecli)

[![Go](https://github.com/olblak/updatecli/workflows/Go/badge.svg)](https://github.com/olblak/updatecli/actions?query=workflow%3AGo)
[![Release Drafter](https://github.com/olblak/updatecli/workflows/Release%20Drafter/badge.svg)](https://github.com/olblak/updatecli/actions?query=workflow%3A%22Release+Drafter%22)

Updatecli is a tool uses to apply file update strategies. Designed to be used from any CI environment, it detects if a value needs to be updated using a custom strategy.
It helps to fight outdated configuration files.

Updatecli reads a yaml or a go template configuration file, then works into three stages

1. Source: Based on a rule fetch a value that will be injected in later stages.
2. Conditions: Ensure that conditions are met based on the value retrieved during the source stage.
3. Target: Update and publish the target files based on a value retrieved from the source stage.

**[Documentation](doc/README.adoc)**

**[Contributing](/CONTRIBUTING.md)**

**[Adopters](/ADOPTERS.md)**

**[License](/LICENSE.md)**
