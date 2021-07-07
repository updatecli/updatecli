---
sources:
  stable:
    kind: jenkins
    depends_on:
      - weekly
    name: Get Latest Jenkins stable version
    spec:
      github:
        token: {{ requiredEnv .github.token }}
        username: {{ .github.username }}
  weekly:
    kind: jenkins
    name: Get Latest Jenkins weekly version
    spec:
      release: weekly
      github:
        token: {{ requiredEnv .github.token }}
        username: {{ .github.username }}
conditions:
  stabledockerImage:
    name: 'Is docker image jenkins/jenkins:{{ pipeline "Sources.stable.output" }} published?'
    kind: dockerImage
    sourceID: stable
    spec:
      image: jenkins/jenkins
      tag: '{{ pipeline "Sources.stable.output" }}-jdk11'
  weeklydockerImage:
    name: 'Is docker image tag{{ pipeline "Sources.weekly.output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ pipeline "Sources.weekly.output" }}-jdk11'
  weekly2dockerImage:
    name: 'Is docker image tag{{ pipeline "Sources.weekly.output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ pipeline "Sources.weekly.output" }}-jdk11'
  weekly3dockerImage:
    name: 'Is docker image tag{{ pipeline "Sources.weekly.output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ pipeline "Sources.weekly.output" }}-jdk11'
targets:
  imageTag:
    name: 'Update jenkins/jenkins docker tag to {{ pipeline "Sources.weekly.output" }}-jdk11'
    kind: yaml
    sourceID: stable
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.tag"
      value: '{{ pipeline "Sources.weekly.output" }}-jdk11' 
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
