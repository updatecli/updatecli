---
source:
  name: Get Latest helm release version
  kind: githubRelease
  spec:
    owner: "helm"
    repository: "helm"
    token: {{ requiredEnv .github.token }}
    username: olblak
    version: latest
conditions:
  isENVSet:
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
        branch: "main"
  isARGSet:
    name: Is there any ARG instruction starting with "HELM_VERSION"
    kind: dockerfile
    spec:
      file: Dockerfile
      instruction:
        keyword: "ARG"
        matcher: "HELM_VERSION"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "main"
targets:
  updateENVHELMVERSION:
    name: Update the 2nd element of the 2nd ENV instruction to the source value
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
        branch: "main"
  updateARGTERRAFORMVERSION:
    name: Update all ARG instructions starting with HELM_VERSION
    kind: dockerfile
    spec:
      file: Dockerfile
      instruction:
        keyword: "ARG"
        matcher: "TERRAFORM_VERSION"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "main"
