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
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"

targets:
  chartVersion:
    name: "grafana/loki Helm Chart"
    kind: yaml
    spec:
      file: "helmfile.d/loki.yaml"
      key: "releases[0].version"
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
