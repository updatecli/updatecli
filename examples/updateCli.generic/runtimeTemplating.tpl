---
sources:
  stable:
    kind: jenkins
    depends_on:
      - weekly
    name: 'Get Latest Jenkins stable version and depends on {{ pipeline "sources.weekly.name" }}'
#    spec:
#      github:
#        token: {{ requiredEnv .github.token }}
#        username: {{ .github.username }}
  weekly:
    kind: jenkins
    name: Get Latest Jenkins weekly version
#    spec:
#      release: weekly
#      github:
#        token: {{ requiredEnv .github.token }}
#        username: {{ .github.username }}
conditions:
  stabledockerImage:
    name: 'Is docker image jenkins/jenkins:{{ pipeline "sources.stable.kind" }} published?'
    kind: dockerImage
    sourceID: stable
    spec:
      image: jenkins/jenkins
      tag: '{{ source "stable" }}-jdk11'
  weeklydockerImage:
    name: 'Is docker image tag{{ context "Sources.weekly.Output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ context "Sources.weekly.Output" }}-jdk11'
  weekly2dockerImage:
    name: 'Is docker image tag{{ context "Sources.weekly.Output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ context "Sources.weekly.Output" }}-jdk11'
  weekly3dockerImage:
    name: 'Is docker image tag{{ context "Sources.weekly.Output" }} published?'
    kind: dockerImage
    sourceID: weekly
    spec:
      image: jenkins/jenkins
      tag: '{{ context "Sources.weekly.Output" }}-jdk11'
targets:
  imageTag:
    name: 'Update jenkins/jenkins docker tag to {{ context "Sources.weekly.Output" }}-jdk11'
    kind: yaml
    sourceID: stable
    spec:
      file: "charts/jenkins/values.yaml"
      key: "jenkins.controller.tag"
      value: '{{ context "Sources.weekly.Output" }}-jdk11' 
    scm:
      git:
        url: "git@github.com:olblak/charts.git"
        branch: master
        user: olblak
        email: me@olblak.com
