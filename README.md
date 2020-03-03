# Updatecli

**Prototype**

Updatecli is a small cli tool used for automating yaml values updates.
It fetches its configuration from one yaml file and then works in three stages

1. Source: Based on a rule fetch a value that will be used during later stages
2. Conditions: Ensure that conditions are met based on the value retrieved during the source rule
3. Target: Update and publish the target files based on a value retrieved from the source stage.

**Remark: Environment variable with a name matching the key uppercased and prefixed with UPDATECLI, can be used in place of key/value in configuration file.**

## Source

Currently "source" only supports two kinds of sources

### Github Release

This source will check Github Release for a specific version if latest is specified, it retrieves the version referenced by 'latest'

```
source:
  kind: githubRelease
  spec:
    owner: "Github Owner"
    repository: "Github Repository"
    token: "You should use environment variable!"
    url: "Github Url"
    version: "Version to fetch"
```

Environment variable `UPDATECLI_SOURCE_SPEC_TOKEN` can be used instead of writing secrets in files

### DockerRegistry

**Not Ready Yet Implemented**

This source will check a docker image tag from a docker registry and return its digest, so we always reference a specific image, even when the tag is updating regularly.

```
source:
  kind: dockerTag
  spec:
    image: "Docker Image"
    url: "Docker registry url"
    tag: "Docker Image Tag to fetch the checksum"
```

## Condition
 It will check for a environment variable with a name matching the key uppercased and prefixed with the EnvPrefix
During this stage, we check if conditions are met based on the value retrieved in the source stage

### dockerImage

This condition checks if a docker image with a specific tag is published on Docker Registry.
If the condition is not met, it skips the target stage.

```
conditions:
  id:
    kind: dockerImage
    spec:
      image: _Docker Image_
      url: _Docker Registry url_
      tag: _Docker Image Tag_
```

## Targets

"Targets" stage will update the definition for every targets based on the value return during the source stage if all conditions are met.

### yaml

This target will update an yaml file base a value retrieve during the source stage.

```
targets:
  id:
    kind: yaml
    spec:
      file: "Yaml file path from the root repository"
      key: "yaml key to update"
      message: "Git message to identify this commit change"
      scm: "scm repository type"
      repository:
        url: "git repository url"
        branch: "git branch to push changes"
        user: "git user to push from changes"
        email: "git user email to push from change"
        directory: "directory where to clone the git repository"
```
#### scm
Yaml accept two kind of scm, github and git.

##### git
Git push every changes on the remote git repository

repository:
  url: "git repository url"
  branch: "git branch to push changes"
  user: "git user to push from changes"
  email: "git user email to push from change"
  directory: "directory where to clone the git repository"

##### github
Github  push every changes on a temporary branch then open a pull request

repository:
  user: "git user to push from changes"
  email: "git user email to push from change"
  directory: "directory where to clone the git repository"
  owner: "github owner"
  repository: "github repository"
  token: "github token with enough permission on repository"
  username: "github username used for push git changes"
  branch: "git branch where to push changes"

## Usage

A configuration can be specified by using --config, it accepts either a single file or a directory, if a directory is specified, then it checks for every files inside.

### Docker
A docker image is available.

`docker run -i -t -v "$PWD/updateCli.yaml":/home/updatecli/updateCli.yaml:ro olblak/updatecli:latest --config /home/updatecli/updateCli.yaml`
