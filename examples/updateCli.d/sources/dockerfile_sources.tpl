---
title: Bump jenkinsciinfra/helmfile image
sources:
  helm:
    name: Get Latest helm release version
    kind: githubRelease
    spec:
      owner: "helm"
      repository: "helm"
      token: {{ requiredEnv .github.token }}
      username: olblak
      version: latest
  checksum:
    name: Get Latest kubectl release version
    kind: file
    transformers:
      - find: '^\S*'
    spec:
      # For some reason helm must start with lower case
      file: https://get.helm.sh/helm-{{ pipeline "Sources.helm.Output" }}-linux-amd64.tar.gz.sha256sum
conditions:
  isENVSet:
    sourceID: helm
    name: Is the 2nd ENV instruction having a "keyword" set to "HELM_VERSION"
    kind: dockerfile
    spec:
      file: docker/Dockerfile
      Instruction: "ENV[1][0]"
      Value: "HELM_VERSION"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
targets:
  updateHelm:
    name: Update the 2nd element of the 2nd ENV instruction to the source value
    sourceID: helm
    kind: dockerfile
    spec:
      file: docker/Dockerfile
      Instruction: ENV[1][1]
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
  updateChecksum:
    name: Update Helm checksum
    sourceID: checksum
    kind: dockerfile
    spec:
      file: docker/Dockerfile
      Instruction: ENV[5][1]
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
