# Updatecli

<img src="https://www.updatecli.io/images/updatecli.png" alt="Updatecli" align="right" width="200" height="200"/>

[![](https://img.shields.io/matrix/updatecli:matrix.org)](https://matrix.to/#/#Updatecli_community:gitter.im)
[![GitHub](https://img.shields.io/github/license/updatecli/updatecli)](https://github.com/updatecli/updatecli/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/updatecli/updatecli)](https://goreportcard.com/report/github.com/updatecli/updatecli)
[![Codecov](https://img.shields.io/codecov/c/github/updatecli/updatecli)](https://codecov.io/gh/updatecli/updatecli)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c)](https://pkg.go.dev/github.com/updatecli/updatecli)
[![GitHub Releases](https://img.shields.io/github/downloads/updatecli/updatecli/total)](https://github.com/updatecli/updatecli/releases)
[![GitHub Releases](https://img.shields.io/github/downloads/updatecli/updatecli/latest/total)](https://github.com/updatecli/updatecli/releases)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/updatecli/updatecli/go.yaml?branch=main)](https://img.shields.io/github/actions/workflow/status/updatecli/updatecli/go.yaml?branch=main)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/updatecli/updatecli/badge)](https://api.securityscorecards.dev/projects/github.com/updatecli/updatecli)
[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/6731/badge)](https://bestpractices.coreinfrastructure.org/projects/6731)

_"Automatically open a PR on your Git repository when a file update is needed"_

Updatecli is a universal declarative update policy engine. Designed to be used from everywhere, each Updatecli "run" detects if a file needs to be updated using a tailored update policy then apply changes.

You describe your update strategy in a policy then you enforce it using Updatecli.

Every Updatecli policy is a YAML (or Go template) file that runs through three stages:

1. **Sources** — Fetch the new value to apply (e.g. latest Docker image tag, newest Helm chart version, latest GitHub release).
2. **Conditions** — Verify that all prerequisites are met before making any change (optional but recommended).
3. **Targets** — Apply the change to the right file or service, and open a pull request if an SCM is configured.

## What Can You Update?

Updatecli ships with 30+ built-in integrations. Here are the most common scenarios:

- 📄 **File formats** —
  Update values in YAML, JSON, TOML, XML, HCL, CSV, Dockerfiles, and `.tool-versions` files,
  or any text file with pattern matching.

- 🐳 **Container images** —
  Track Docker image tags and digests from Docker Hub or any OCI-compliant registry.

- 📦 **Package registries** —
  Helm charts (including OCI), npm, PyPI, Maven, Cargo (Rust crates), Go modules,
  and Terraform providers/modules.

- 🏷️ **Git & releases** —
  GitHub/GitLab releases, Git tags and branches.

- ☕ **Languages & runtimes** —
  Jenkins LTS/weekly releases, Eclipse Temurin (JDK), Go language versions,
  Bazel modules and registry.

- ☁️ **Cloud** —
  AWS AMIs.

- 🔧 **Custom logic** —
  Shell scripts and HTTP endpoints, for anything not covered above.

Find the full list and documentation on [www.updatecli.io](https://www.updatecli.io/docs/prologue/introduction/).

## Feature

| SCM Platform | Supported | Capabilities | Plugin |
|---|---|---|---|
| GitHub | ✅ | Clone, branch, commit, push, Pull Requests, releases | `github` |
| GitLab | ✅ | Clone, branch, commit, push, Merge Requests, releases | `gitlab` |
| Gitea | ✅ | Clone, branch, commit, push, Pull Requests, releases | `gitea` |
| Forgejo | ✅ | Compatible through Gitea API support | `gitea` |
| Bitbucket Cloud | ✅ | Clone, branch, commit, push, Pull Requests | `bitbucket` |
| Bitbucket Server / Stash | ✅ | Clone, branch, commit, push, Pull Requests | `stash` |
| Azure DevOps | ✅ | Clone, branch, commit, push, Pull Requests | `azuredevops` |
| Generic Git Repository | ✅ | Clone, branch, commit, push | `git` |

- **Declarative** — Define your update policy once in YAML; Updatecli handles detection and application.
- **30+ integrations** — Docker, Helm, GitHub releases, npm, PyPI, Terraform, AWS AMIs, and more — out of the box.
- **Any CI/CD** — Runs as a single binary. Drop it into GitHub Actions, Jenkins, GitLab CI, or any shell.
- **Safe by default** — Use `--dry-run` to preview every change before it is applied.
- **Extensible** — Add custom logic via shell scripts or HTTP, or contribute a new Go plugin.

## Why

There are already many projects out there to continuously update your files, but they all have an opinionated way of doing it and they often want you to adopt a new platform.
Building and distributing software is a difficult task and good practices constantly evolve.
Updatecli was built to work independently of the underlying dependencies to update, wherever you need it and combining whatever workflow you are using, as you can see in the following section.

## Demo

[![Asciinema](https://asciinema.org/a/CR5DIxyTLnvtt8NllEeYAx83U.svg)](https://asciinema.org/a/CR5DIxyTLnvtt8NllEeYAx83U)

**The Quick-start is available on [www.updatecli.io/docs/prologue/quick-start](https://www.updatecli.io/docs/prologue/quick-start/)**

## Installation

Updatecli is a Go binary available for Linux, MacOS and Windows from the [release page](https://github.com/updatecli/updatecli/releases) or installed via [other methods](https://www.updatecli.io/docs/prologue/installation/).

**Verify File Checksum Signature**

Instead of signing all release assets, Updatecli signs the checksums file containing the different release assets checksum.
You can download/copy the three files 'checksums.txt.sig' and 'checksums.txt' from the latest [release](https://github.com/updatecli/updatecli/releases/latest).
Once you have the three files locally, you can execute the following command

```
cosign verify-blob --certificate-identity-regexp "https://github.com/updatecli/updatecli" --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' --bundle checksums.txt.sig checksums.txt
```

A successful output looks like

```
Verified OK
```

Now you can verify the assets checksum integrity.

**Verify File Checksum Integrity**

Before verifying the file integrity, you should first verify the checksum file signature.
Once you've download both the checksums.txt and your binary, you can verify the integrity of your file by running:

```
sha256sum --ignore-missing -c checksums.txt
```

**Verify Container signature**

```
cosign verify --certificate-identity-regexp "https://github.com/updatecli/updatecli" --certificate-oidc-issuer "https://token.actions.githubusercontent.com" ghcr.io/updatecli/updatecli:v0.117.0
```

## Documentation

The documentation of Updatecli is available at [www.updatecli.io](https://www.updatecli.io/docs/prologue/introduction/), but you can also look at the `examples` section to get an overview.

### Example

This example is copy of the quickstart. You can also find it on [www.updatecli.io/docs/prologue/quick-start](https://www.updatecli.io/docs/prologue/quick-start/)

We define an update strategy in "updatecli.yaml" then we run `updatecli apply --config updatecli.yaml`.
Our objective is to know if the Jenkins project published a new stable version, if they build an appropriated docker image specifically for jdk11 and automatically update our infrastructure accordingly.

<table>
<tr>
<td>

```yaml
## updatecli.yaml
name: Update Jenkins Version

scms:
  default:
    kind: github
    spec:
      user: olblak
      email: me@olblak.com
      owner: olblak
      repository: charts
      token: mySecretTokenWhichShouldNeverUsedThisWay
      username: olblak
      branch: main

sources:
  jenkins:
    name: Get latest Jenkins version
    kind: jenkins
    spec:
      release: weekly

conditions:
  docker:
    name: "Test if Docker Image jenkins/jenkins is Published on DockerHub"
    kind: dockerimage
    spec:
      image: jenkins/jenkins
      architecture: amd64

targets:
  bumpJenkins:
    name: Update values.yaml to the latest Jenkins version
    scmID: default
    kind: yaml
    spec:
      file: charts/jenkins/values.yaml
      key: $.jenkins.controller.imageTag

actions:
  default:
    title: Open a GitHub pull request with new Jenkins version
    kind: github/pullrequest
    scmID: default
    target:
      - bumpJenkins
    spec:
      automerge: true
      mergemethod: squash
      labels:
        - dependencies
```

</td>
<td>

What it says:

1. Sources:
   What's the latest jenkins weekly version?
   => 2.335

2. Conditions:
   Is there a docker image "jenkins/jenkins" from Dockerhub with the tag "2.335"
   => Yes then proceed otherwise abort

3. Targets:
   Do we have to update the key "jenkins.controller.imageTag" from file "./charts/jenkins/values.yaml" located on the GitHub repository olblak/charts to "2.335"?
   => If yes then execute the action `default` opening a GitHub pull request to the "main" branch

</td>
</tr>
</table>

More information [here](https://www.updatecli.io/docs/prologue/introduction/)

---

## Roadmap

We use the GitHub milestone [**Next**](https://github.com/updatecli/updatecli/milestone/73) to prioritize our effort. As our requirements evolve we regularly add plugins or improve existing ones.

If you ever need a specific integration, feel free to either:

1. Contribute it, we are more than happy to help. [Link](https://github.com/updatecli/updatecli/blob/main/CONTRIBUTING.adoc)
2. Comment on existing issues as we may prioritize issues affecting other users. [Link](https://github.com/updatecli/updatecli/issues)
3. Sponsor financially the project [link](https://github.com/sponsors/olblak)
4. Feel free to reach out to [contact@updatecli.io](mailto:contact@updatecli.io) to see how we can help you.

## Contributing

As a community-oriented project, all contributions are greatly appreciated!

Here is a non-exhaustive list of possible contributions:

- ⭐️ this repository.
- Propose a new feature request.
- Highlight an existing feature request with 👍.
- Contribute to any repository in the [updatecli](https://github.com/updatecli/) organization
- Share the love

More information available at [CONTRIBUTING](https://github.com/updatecli/updatecli/blob/main/CONTRIBUTING.adoc)

## Conferences

- 2026
  - Devoxx (FR) - Industrialisez la maintenance de vos stacks Dev et Ops [Video](https://www.youtube.com/watch?v=RC489sMFrF0)
- 2025
  - FOSDEM (BE) - Continuously Update Everything two years later [Event](https://fosdem.org/2025/schedule/event/fosdem-2025-6076-continuously-update-everything-two-years-later)
  - Incontro DevOps (IT) - Automate or stagnate: surviving the era of continuous updates [Video](https://www.youtube.com/watch?v=kVOYFFbbCho)
  - OSX (FR) - Updatecli: how to keep your applications up-to-date without losing your mind [Video](https://www.youtube.com/watch?v=Z-RPVbUPYiI)
- 2024
  - DevOps Days 15 year anniversary celebration (BE) - Continuously Update Everything [Video](https://www.youtube.com/watch?v=PlEm-YinALk)
- 2023
  - FOSDEM (BE) - Cloud Native Dependencies [Video](https://fosdem.org/2023/schedule/event/continuous_update_everything/)
  - CIVO Cloud (UK) - Onward a Continuously Updated Kubernetes Marketplace [Video](https://www.youtube.com/watch?v=B2wmA627E4w)
- 2022
  - CDcon (US) - Dependency Management: Where the Fork are We? [Video](https://youtu.be/157bsLD-0mM)

## Links

- [ADOPTERS](https://github.com/updatecli/updatecli/blob/main/ADOPTERS.md)
- [CONTRIBUTING](https://github.com/updatecli/updatecli/blob/main/CONTRIBUTING.adoc)
- [DOCUMENTATION](https://www.updatecli.io/docs/prologue/introduction/)
- [LICENSE](https://github.com/updatecli/updatecli/blob/main/LICENSE)

## Thanks to you

### Contributors ❤️

[![](https://contrib.rocks/image?repo=updatecli/updatecli)](https://github.com/updatecli/updatecli/graphs/contributors)
