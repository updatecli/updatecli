---
###
# This strategy will:
#
# Source:
# =======

#   Retrieve the version from the Jenkins helm chart repository located on "https://charts.jenkins.io"
#     => 2.7.1
#
# Conditions:
# ===========
#
#   Then it will test two conditions.
#   1 - Test a helmchart condition, "Is the prometheus helm chart version "11.16.5" is available from https://prometheus-community.github.io/helm-charts?
#     => Yes, proceed, No then abort
#   2 - Test a yaml condition, "Do we have an yaml file named "charts/jenkins/requirements.yaml" with the key dependencies that contains an array where the first element is set to "jenkins" ?"
#     => Yes, proceed, No then abort
#
#  Targets:
#  ========
#
#  If conditions are all met, then updatecli will update (if needed) the first element of the key
#  "dependencies" to "2.7.1" for the file "charts/jenkins/requirements.yaml"
#  from the github repository olblak/chart then commit the change to a temporary branch then open
#  a pull request targeting main
#
# Remark: The specificity in this example is that we are using a go template
# so we could reuse information accross the yaml file or use environment variable which contains the github token
#
###

source:
  kind: helmChart
  spec:
    url: https://charts.jenkins.io
    name: jenkins
conditions:
  isPrometheuseHelmChartVersionAvailable:
    name: "Test if the prometheus helm chart is available"
    kind: helmChart
    spec:
      url: https://prometheus-community.github.io/helm-charts
      name: prometheus
      version: 11.16.5
  chartVersion:
    name: "jenkinsci/jenkins Helm Chart used"
    kind: yaml
    spec:
      file: "charts/jenkins/requirements.yaml"
      key: "dependencies[0].name"
      value: "jenkins"
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
    name: "jenkinsci/jenkins Helm Chart"
    kind: yaml
    spec:
      file: "charts/jenkins/requirements.yaml"
      key: "dependencies[0].version"
    scm:
      github:
        user: "{{ .github.user }}"
        email: "{{ .github.email }}"
        owner: "{{ .github.owner }}"
        repository: "{{ .github.repository }}"
        token: "{{ requiredEnv .github.token }}"
        username: "{{ .github.username }}"
        branch: "{{ .github.branch }}"
