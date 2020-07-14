source:
  kind: helmChart
  spec:
    url: https://kubernetes-charts.storage.googleapis.com
    name: jenkins

conditions:
  exist:
    name: "Docker Image Published on Registry"
    kind: helmChart
    spec:
      url: https://kubernetes-charts.storage.googleapis.com
      name: jenkins
  chartDependencyIsJenkins:
    name: "Helm Chart"
    kind: yaml
    spec:
      file: "charts/jenkins/requirements.yaml"
      key: "dependencies[0].name"
      value: "jenkins"
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
  imageTag:
    name: "Helm Chart"
    kind: yaml
    spec:
      file: "charts/jenkins/requirements.yaml"
      key: "dependencies[0].version"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
