---
source:
  kind: jenkins
  spec:
    release: stable
    github:
      token: {{ requiredEnv .github.token }}
      username: {{ .github.username }}
conditions:
  jenkinsVersion:
    kind: jenkins
    spec:
      version: "2.275"
  jenkinsStable:
    kind: jenkins
    spec:
      release: stable
  imageTag:
    name: "jenkins/jenkins docker image set"
    kind: yaml
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.image"
      value: "jenkins/jenkins"
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
  dockerImage:
    kind: dockerImage
    postfix: "-jdk11"
    spec:
      image: jenkins/jenkins
targets:
  imageTag:
    name: "jenkins/jenkins docker tag"
    kind: yaml
    postfix: "-jdk11"
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.tag"
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
