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
      key: "jenkins.controller.tag"
      incminor: true
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv .github.token }}
        username: "olblak"
        branch: "master"
