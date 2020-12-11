source:
  kind: githubRelease
  spec:
    owner: "jenkins-infra"
    repository: "jenkins-wiki-exporter"
    token: {{ requiredEnv "GITHUB_TOKEN" }}
    username: "olblak"
    version: "latest"
conditions:
  docker:
    name: "Docker Image Published on Registry"
    kind: dockerImage
    spec:
      image: "halkeye/jenkins-wiki-exporter"
targets:
  imageTag:
    name: "Docker Image"
    kind: yaml
    spec:
      file: "charts/jenkins-wiki-exporter/values.yaml"
      key: image.tag
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
  #appVersion:
  #  name: "Chart appVersion"
  #  kind: yaml
  #  spec:
  #    file: "charts/jenkins-wiki-exporter/Chart.yaml"
  #    key: appVersion
  #  scm:
  #    github:
  #      user: "updatecli"
  #      email: "updatecli@olblak.com"
  #      owner: "olblak"
  #      repository: "charts"
  #      token: {{ requiredEnv "GITHUB_TOKEN" }}
  #      username: "olblak"
  #      branch: "master"
  appVersion:
    name: "Chart appVersion"
    kind: yaml
    spec:
      file: "charts/jenkins-wiki-exporter/Chart.yaml"
      key: appVersion
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
