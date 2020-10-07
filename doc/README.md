## Documentation

Updatecli is a tool to define and apply file update strategy.

It reads its configuration from a yaml or go template configuration file, then works into three stages

1. Source: Based on a rule fetch a value that will be injected in later stages
2. Conditions: Ensure that conditions are met based on the value retrieved during the source rule
3. Target: Update and publish the target files based on a value retrieved from the source stage.

## Installation
Updatecli is available [here](https://github.com/olblak/updatecli/releases/latest).

A docker image is also available on [DockerHub](https://hub.docker.com/r/olblak/updatecli)

## Usage

'updateCli' is a tool that updates files according to a custom update strategy definition. Once your strategy has been defined, just call one of the following:

- `updatecli diff --config strategy.yaml`
- `updatecli apply --config strategy.yaml` 
- `updatecli help`

or using the docker image

- `docker run -i -t -v "$PWD/updateCli.yaml":/home/updatecli/updateCli.yaml:ro olblak/updatecli:v0.0.20 diff --config /home/updatecli/updateCli.yaml`
- `docker run -i -t -v "$PWD/updateCli.yaml":/home/updatecli/updateCli.yaml:ro olblak/updatecli:v0.0.20 apply --config /home/updatecli/updateCli.yaml`
- `docker run -i -t olblak/updatecli:v0.0.20 help`

## Strategy

.strategy.yaml
```
source:
  kind: <sourceType>
  spec:
    <sourceTypeSpec>>
conditions:
  conditionID:
    kind: <conditionType>
    spec: 
      <conditionTypeSpec>
targets:
  target1:
    kind: <targetType>
    spec:
      <targetTypeSpec>
```

**YAML** 

Accepted extensions: ".yaml",".yml"

A YAML configuration can be specified using `--config <yaml_file>`, it accepts either a single file or a directory, if a directory is specified, then it runs recursively on every file inside the directory.

**Go Templates**

Accepted extensions: ".tpl",".tmpl"

Another way to use this tool is by using go template files in place of YAML. 
Using go templates allow us to specify generic values in a different yaml file then reference those values from each go templates.
We also provide a custom function called requireEnd to inject any environment variable in the template example, `{{ requiredEnv "PATH" }}`.

The strategy file can either be using a yaml format or a golang template.

### Source

#### Github Release

This source will get a release version from Github Release api. If `latest` is specified, it retrieves the version referenced by 'latest'.

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

**A configuration using go template can be used to retrieve the environment variable instead of writing secrets in files, cfr later.**

#### DockerRegistry

This source will get a docker image tag from a docker registry and return its digest, so we can always reference a specific image tag like `latest`.

```
source:
  kind: dockerDigest
  spec:
    image: "Docker Image"
    url: "Docker registry url" # Not Mandatory
    tag: "Docker Image Tag to fetch the checksum"
```

#### HelmChart
This source will get the latest helm chart version available.

```
source
  kind: helmChart
  spec:
    url: https://kubernetes-charts.storage.googleapis.com
    name: jenkins
```

#### Maven

This source will get the latest maven artifact version.

```
source:
  kind: maven
  spec:
    url:  "repo.jenkins-ci.org",
	repository: "releases",
	groupID:    "org.jenkins-ci.main",
	artifactID: "jenkins-war",
```

#### Replacer
A List of replacer rules can be provided to modify the value retrieved from source.

```
source:
  kind: githubRelease
  replaces: 
    - from: "string"
      to: ""
    - from: "substring1"
      to: "substring2"
  spec:
    owner: "Github Owner"
    repository: "Github Repository"
    token: "Don't commit your secrets!"
    url: "Github Url"
    version: "Version to fetch"
```


#### Prefix/Postfix
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


### Condition
During this stage, we check if conditions are met based on the value retrieved from the source stage otherwise we can skip the "target" stage.

#### dockerImage

This condition checks if a docker image tag is available from a Docker Registry.

```
conditions:
  id:
    kind: dockerImage
    spec:
      image: _Docker Image_
      url: _Docker Registry url_ #Not mandatory
```

#### Maven
This condition checks if the source value is available on a maven repository

```
condition:
  kind: maven
  spec:
    url:  "repo.jenkins-ci.org",
	repository: "releases",
	groupID:    "org.jenkins-ci.main",
	artifactID: "jenkins-war",
```

#### HelmChart
This source checks if a helm chart exist, a version can also be specified

```
source
  kind: helmChart
  spec:
    url: https://kubernetes-charts.storage.googleapis.com
    name: jenkins
    version: 'x.y.x' (Optional)
```

### Targets

"Targets" stage will update the definition for every target based on the value returned during the source stage if all conditions are met.

#### yaml

This target will update a yaml file base a value retrieve during the source stage.

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

### scm
Depending on the situation a specific scm block can be provided to the target and condition stage. At the moment it supports github and git.

#### git
Git push every change on the remote git repository

```
git:
  url: "git repository url"
  branch: "git branch to push changes"
  user: "git user to push from changes"
  email: "git user email to push from change"
  directory: "directory where to clone the git repository"
```

#### github
Github  push every change on a temporary branch then open a pull request

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
A prefix and/or postfix can be added based on the value retrieved from the source.
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

## Examples

This project is currently used to automate Jenkins OSS kubernetes cluster
* [UpdateCli configuration](https://github.com/jenkins-infra/charts/tree/master/updateCli/updateCli.d)
* [Jenkinsfile](https://github.com/jenkins-infra/charts/blob/master/Jenkinsfile_k8s#L35L48)
* [Results]()
  * [Docker Digest](https://github.com/jenkins-infra/charts/pull/188)
  * [Maven Repository](https://github.com/jenkins-infra/charts/pull/179)
  * [Github Release](https://github.com/jenkins-infra/charts/pull/145)
