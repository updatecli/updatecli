source:
  kind: githubRelease
  spec:
    owner: olblak
    repository: updatecli
    token: {{ requiredEnv "GITHUB_TOKEN" }}
    username: olblak
    version: latest
conditions:
  dockerFile:
    name: isDockerfileCorrect
    kind: dockerfile
    spec:
      file: Dockerfile
      #instruction: FROM
      ## value: "golang:1.14"
      #value: "golang:1.15"
      instruction: USER
      # value: "golang:1.14"
      value: updatecli
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "updatecli"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: master
targets:
  dockerFile:
    name: "isDockerfileCorrect"
    kind: dockerfile
    prefix: "olblak/updatecli:"
    spec:
      file: "Dockerfile"
      instruction: "FROM"
    scm:
      github:
        user: "updatecli"
        owner: "olblak"
        repository: "updatecli"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: master
        email: "update-bot@olblak.com"
