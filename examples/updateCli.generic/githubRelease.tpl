---
###
# This strategy will:
#
# Source:
# =======

#   Retrieve the latest version from the github Release on the repository "jenkins-infra/jenkins-wiki-exporter"
#     => 1.10.3
#
# Conditions:
# ===========
#
#   Then it will test two conditions.
#   1 - Test a docker image condition, Does it exist a docker image "jenkinsciinfra/jenkins-wiki-exporter" with the tag 1.10.3
#     => Yes, proceed, No then abort
#   2 - Test a yaml condition, "Do we have an yaml file named "charts/jenkins-wiki-exporter/values.yaml" with the key "image.repository" set to "jenkinsciinfra/jenkins-wiki-exporter" ?"
#     => Yes, proceed, No then abort
#
#  Targets:
#  ========
#
#  If conditions are all met, then updatecli will update (if needed) the key
#  "image.tag" to "1.10.3" for the file "charts/jenkins-wiki-exporter/values.yaml"
#  from the github repository olblak/chart then commit the change to a temporary branch then open
#  a pull request targeting main
#
# Remark: The specificity in this example is that we are using a go template
# so we could reuse information accross the yaml file or use environment variable which contains the github token
#
###

sources:
  kubectl:
    kind: githubRelease
    spec:
      owner: "kubernetes"
      repository: "kubectl"
      token: "{{ requiredEnv .github.token }}"
      username: "olblak"
      #version: 'kubernetes-1.(\d*).(\d*)$' # return latest publish version starting with kubernetes-1.20
      #version: latest
      version: v0.20
      versioning:
        kind: text
    transformers:
      - trimPrefix: "kubernetes-"
  jenkins-wiki-exporter:
    kind: githubRelease
    spec:
      owner: "jenkins-infra"
      repository: "jenkins-wiki-exporter"
      token: "{{ requiredEnv .github.token }}"
      username: "olblak"
      version: "~1.10"
      versioning:
        kind: semver
    transformers:
      - addPrefix: "v"
conditions:
  docker:
    sourceID: jenkins-wiki-exporter
    name: "Docker Image Published on Registry"
    kind: dockerImage
    spec:
      image: "jenkinsciinfra/jenkins-wiki-exporter"
  imageName:
    sourceID: jenkins-wiki-exporter
    name: "jenkinsci/jenkins Helm Chart used"
    kind: yaml
    spec:
      file: "charts/jenkins-wiki-exporter/values.yaml"
      key: "image.repository"
      value: "jenkinsciinfra/jenkins-wiki-exporter"
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
    sourceID: jenkins-wiki-exporter
    name: "jenkinsci/jenkins Helm Chart"
    kind: yaml
    spec:
      file: "charts/jenkins-wiki-exporter/values.yaml"
      key: "image.tag"
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
