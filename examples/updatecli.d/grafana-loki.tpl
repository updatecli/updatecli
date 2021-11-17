source:
  kind: helmChart
  name: "Get latest loki helm chart version"
  spec:
    url: https://grafana.github.io/loki/charts
    name: loki

conditions:
  exist:
    name: "Is Loki helm chart available on Registry?"
    kind: helmChart
    spec:
      url: https://grafana.github.io/loki/charts
      name: loki
  isNameGrafana:
    kind: yaml
    name: "Is loki release name is correctly set?"
    spec:
      file: "helmfile.d/loki.yaml"
      key: "releases[0].name"
      value: "loki"
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
    name: "Update grafana/loki Helm Chart to latest version"
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
