source:
  kind: helmChart
  spec:
    url: https://grafana.github.io/loki/charts
    name: loki

conditions:
  exist:
    name: "Loki helm chart available on Registry"
    kind: helmChart
    spec:
      url: https://grafana.github.io/loki/charts
      name: loki
  isNameGrafana:
    kind: yaml
    spec:
      file: "helmfile.d/loki.yaml"
      key: "releases[0].name"
      value: "grafana"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "jenkins-infra"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"

targets:
  chartVersion:
    name: "grafana/loki Helm Chart"
    kind: yaml
    spec:
      file: "helmfile.d/loki.yaml"
      key: "releases[0].version"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "jenkins-infra"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
