source:
  kind: githubRelease
  spec:
    owner: "jenkins-infra"
    repository: "plugin-site-api"
    token: {{ requiredEnv "GITHUB_TOKEN" }}
    username: "olblak"
    version: "latest"
targets:
  imageTag:
    name: "Docker Image"
    kind: yaml
    spec:
      file: "charts/plugin-site/values.yaml"
      key: "backend.image.tag"
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
  appVersion:
    name: "Chart appVersion"
    kind: yaml
    spec:
      file: "charts/plugin-site/Chart.yaml"
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
#      git:
#        url: "git@github.com:olblak/charts.git"
#        branch: "updatecli/Helm_Chart/2.3.3"
#        user: "update-bot"
#        email: "update-bot@olblak.com"
