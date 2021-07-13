---
sources:
  default:
   kind: jenkins
   spec:
     # already default value
     # release: stable
     github:
       token: {{ requiredEnv .github.token }}
       username: {{ .github.username }}
conditions:
  jenkinsVersion:
    name: Test jenkinsversion 2.263.3
    kind: jenkins
    spec:
      version: "2.263.3"
  defaultjenkinsStable:
    name: Test jenkins stable version
    kind: jenkins
  jenkinsStable:
    kind: jenkins
    name: Test jenkins stable version a second time
    spec:
      release: stable
  imageTag:
    name: "jenkins/jenkins docker image set"
    kind: yaml
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
  dockerImage:
    name: Test jenkins docker image
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
