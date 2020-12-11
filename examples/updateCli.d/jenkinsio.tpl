source:
  kind: dockerDigest
  name: "Get latest nginx:1.17 from dockerhub"
  spec:
    image: "library/nginx"
    tag: "1.17"
targets:
  jenkinsio:
    name: "Update to latest nginx:1.17"
    kind: yaml
    spec:
      file: "charts/jenkinsio/values.yaml"
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
