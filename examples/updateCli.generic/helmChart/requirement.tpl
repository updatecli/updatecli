title: Bump Jenkins Upstream Chart Version
source:
  kind: helmChart
  name: Get official Jenkins Chart Version
  spec:
    url: https://charts.jenkins.io
    name: jenkins
targets:
  chartjenkins:
    name: Bump Jenkins Upstream Chart Version
    kind: helmChart
    spec:
      Name: "charts/jenkins"
      file: "requirements.yaml"
      key: "dependencies[0].version"
      versionIncrement: "patch"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv .github.token }}
        username: "olblak"
        branch: "master"
