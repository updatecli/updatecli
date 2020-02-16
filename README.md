# Updatecli

**Prototype**

Updatecli is a small cli tool used for automating yaml values updates.
It fetches its configuration from one yaml file and then works in three stages

.1 Source: Based on a rule fetch a value that will be used during later stages
.2 Conditions: Ensure that conditions are met based on the value retrieved during the source rule
.3 Target: Update and publish the target files based on a value retrieved from the source stage.

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
    token: "Github Token"
    url: "Github Url"
    version: "Version to fetch"
```

### DockerRegistry

**Not Yet Implemented**

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

During this stage, we check if conditions are met based on the value retrieved in the source stage

### dockerImage

This condition checks if a docker image with a specific tag is published on Docker Registry

```
conditions:
  - kind: dockerImage
    spec:
      image: _Docker Image_
      url: _Docker Registry url_
      tag: _Docker Image Tag_
```

## Targets

"Targets" stage will update the definition of every targets based on the value return during the source stage if all conditions are met.

### yaml

This target will update a yaml value file

```
targets:
  - kind: yaml
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

## Usage
### Docker
A docker image is available.

`docker run -i -t -v "$PWD/updateCli.yaml":/home/updatecli/updateCli.yaml:ro olblak/updatecli:latest --config /home/updatecli/updateCli.yaml`
