title: Bump Jenkins controller docker image tag
source:
  kind: jenkins
  name: Get Latest Jenkins Stable version
  spec:
    release: stable
    github:
      token: {{ requiredEnv .github.token }}
      username: {{ .github.username }}
targets:
  chartjenkins:
    name: Bump Jenkins controller docker image tag
    kind: helmChart
    spec:
      appVersion: true
      Name: "charts/jenkins"
      file: "values.yaml"
      key: "jenkins.controller.tag"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv .github.token }}
        username: "olblak"
        branch: "master"
