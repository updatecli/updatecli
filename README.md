# Updatecli

[![Go Report Card](https://goreportcard.com/badge/github.com/olblak/updatecli)](https://goreportcard.com/report/github.com/olblak/updatecli)

[![Docker Pulls](https://img.shields.io/docker/pulls/olblak/updatecli?label=olblak%2Fupdatecli&logo=docker&logoColor=white)](https://hub.docker.com/r/olblak/updatecli)

[![Go](https://github.com/olblak/updatecli/workflows/Go/badge.svg)](https://github.com/olblak/updatecli/actions?query=workflow%3AGo)
[![Release Drafter](https://github.com/olblak/updatecli/workflows/Release%20Drafter/badge.svg)](https://github.com/olblak/updatecli/actions?query=workflow%3A%22Release+Drafter%22)

**Prototype**

Updatecli is a tool used for automating values updates.
It fetches its configuration from one yaml configuration file, then works into three stages

1. Source: Based on a rule fetch a value that will be injected in later stages
2. Conditions: Ensure that conditions are met based on the value retrieved during the source rule
3. Target: Update and publish the target files based on a value retrieved from the source stage.

**Remark: Environment variable with a name matching the key uppercased and prefixed with UPDATECLI, can be used in place of key/value in configuration file.**

## Source

### Github Release

This source will check Github Release api for a specific version. If `latest` is specified, it retrieves the version referenced by 'latest'.

.Example
```
source:
  kind: githubRelease
  spec:
    owner: "Github Owner"
    repository: "Github Repository"
    token: "Don't commit your secrets!"
    url: "Github Url"
    version: "Version to fetch"
```

Environment variable `UPDATECLI_SOURCE_SPEC_TOKEN` can be used instead of writing secrets in files or go templates could be used instead of plain YAML, cfr later.

### DockerRegistry

This source will check a docker image tag from a docker registry and return its digest, so we can always reference a specific image tag like `latest`, even when the tag is updating regularly.

```
source:
  kind: dockerDigest
  spec:
    image: "Docker Image"
    url: "Docker registry url" # Not Mandatory
    tag: "Docker Image Tag to fetch the checksum"
```

### HelmChart
This source check if a helm chart version can be updated based on a repository and a chart name

```
source
  kind: helmChart
  spec:
    url: https://kubernetes-charts.storage.googleapis.com
    name: jenkins
```

### Maven

This source will look for the latest version returned from a maven repository

```
source:
  kind: maven
  spec:
    url:  "repo.jenkins-ci.org",
	repository: "releases",
	groupID:    "org.jenkins-ci.main",
	artifactID: "jenkins-war",
```

### Prefix/Postfix
A prefix and/or postfix can be added to any value retrieved from the source.
This prefix/postfix will be used by 'condition' checks, then by every target unless one is explicitly defined in a target.

.Example
```
source:
  kind: githubRelease
  prefix: "v"
  postfix: "-beta"
  spec:
    owner: "Github Owner"
    repository: "Github Repository"
    token: "Don't commit your secrets!"
    url: "Github Url"
    version: "Version to fetch"
```


## Condition
During this stage, we check if conditions are met based on the value retrieved in the source stage otherwise we can skip the "target" stage.

### dockerImage

This condition checks if a docker image with a specific tag is published on Docker Registry.
If the condition is not met, it skips the target stage.

```
conditions:
  id:
    kind: dockerImage
    spec:
      image: _Docker Image_
      url: _Docker Registry url_ #Not mandatory
```

### Maven
This condition checks if a specific version, returned by the source, is published on a maven repository

```
condition:
  kind: maven
  spec:
    url:  "repo.jenkins-ci.org",
	repository: "releases",
	groupID:    "org.jenkins-ci.main",
	artifactID: "jenkins-war",
```

### HelmChart
This source check if a helm chart exist, a version can also be specified

```
source
  kind: helmChart
  spec:
    url: https://kubernetes-charts.storage.googleapis.com
    name: jenkins
    version: 'x.y.x' (Optional)
```

## Targets

"Targets" stage will update the definition for every targets based on the value returned during the source stage if all conditions are met.

### yaml

This target will update an yaml file base a value retrieve during the source stage.

```
targets:
  id:
    kind: yaml
    spec:
      file: "Yaml file path from the root repository"
      key: "yaml key to update"
    scm: #scm repository type"
      #github:
      # or
      #git:
```

NOTE: A key can either be string like 'key' or a position in an array like `array[0]` where 0 means the first element of `array`.
Keys and arrays can also be grouped with dot like `key.array[3].key`.

#### scm
Yaml accept two kind of scm, github and git.

##### git
Git push every changes on the remote git repository

```
git:
  url: "git repository url"
  branch: "git branch to push changes"
  user: "git user to push from changes"
  email: "git user email to push from change"
  directory: "directory where to clone the git repository"
```

##### github
Github  push every changes on a temporary branch then open a pull request

```
github:
  user: "git user to push from changes"
  email: "git user email to push from change"
  directory: "directory where to clone the git repository"
  owner: "github owner"
  repository: "github repository"
  token: "github token with enough permission on repository"
  username: "github username used for push git changes"
  branch: "git branch where to push changes"
```

### Prefix/Postfix
A prefix and/or postfix can be added based value retrieved from the source.
This prefix/postfix won't be used by 'condition' checks. Any value specified at the target level override values defined in the source.

.Example
```
targets:
  imageTag:
    name: "Docker Image"
    kind: yaml
    prefix: "beta-"
    postfix: "-jdk11"
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.master.imageTag"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@example.com"
        owner: "jenkins-infra"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "updatecli"
        branch: "master"
```

## Usage
The best way know how to use this tool is : `updateCli --help`

### YAML {.yaml,.yml}
A YAML configuration can be specified using `--config <yaml_file>`, it accepts either a single file or a directory, if a directory is specified, then it runs recursively on every file inside the directory.

### Go Templates {.tpl, tmpl}
Another way to use this tool is by using go template files in place of YAML, in that case, updateCli can also use the parameter --values <yaml file> to specify YAML key value and then they can be referenced from the go template using {{ key.key2 }}.
We also provide a custom function called requireEnd to inject any environment variable in the template example, `{{ requiredEnv "PATH" }}`.


## Examples

This project is currently used to automate Jenkins OSS kubernetes cluster
* [UpdateCli configuration](https://github.com/jenkins-infra/charts/tree/master/updateCli/updateCli.d)
* [Jenkinsfile](https://github.com/jenkins-infra/charts/blob/master/Jenkinsfile_k8s#L35L48)
* [Results]()
  * [Docker Digest](https://github.com/jenkins-infra/charts/pull/188)
  * [Maven Repository](https://github.com/jenkins-infra/charts/pull/179)
  * [Github Release](https://github.com/jenkins-infra/charts/pull/145)

### Docker
A docker image is available.

`docker run -i -t -v "$PWD/updateCli.yaml":/home/updatecli/updateCli.yaml:ro olblak/updatecli:latest --config /home/updatecli/updateCli.yaml`
