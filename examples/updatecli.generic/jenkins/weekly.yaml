---
sources:
  default:
    # Get latest jenkins weekly version with changelog from github
    kind: jenkins
    spec:
      release: weekly
      github:
        token: {{ requiredEnv .github.token }}
        username: {{ .github.username }}
conditions:
  # Test that a specific Jenkins version exist
  jenkinsVersion:
    kind: jenkins
    spec:
      version: "2.275"
  # Test that the version from source is a weekly release
  jenkinsWeekly:
    kind: jenkins
    spec:
      release: weekly
  # Test that our yaml file is correctly set to a jenkins image
  imageTag:
    name: "jenkins/jenkins docker image set"
    kind: yaml
    disableSourceInput: true
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.image"
      value: "jenkinsciinfra/jenkins-weekly"
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
  # Test that there is a dockeri image with the correct version
  dockerImage:
    kind: dockerImage
    transformers:
      - addSuffix: "-jdk11"
    spec:
      image: jenkins/jenkins
targets:
  imageTag:
    name: "jenkins/jenkins docker tag"
    kind: yaml
    transformers:
      - addSuffix: "-jdk11"
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.tag"
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
