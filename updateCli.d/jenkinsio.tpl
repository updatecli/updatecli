source:
  kind: dockerDigest
  spec:
    image: "library/nginx"
    tag: "1.17"
targets:
  jenkinsio:
    name: "Jenkins.io nginx"
    kind: yaml
    spec:
      file: "charts/jenkinsio/values.yaml"
      key: image.tag
    scm:
      github:
        user: "update-bot"
        email: "update-bot@olblak.com"
        owner: "jenkins-infra"
        repository: "charts"
        token: "{{ requiredEnv "GITHUB_TOKEN" }}"
        username: "olblak"
        branch: "master"
