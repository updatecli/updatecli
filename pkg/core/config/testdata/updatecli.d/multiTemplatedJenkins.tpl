{{- range $release := .releases }}
---
name: Get latest {{ $release.type }} Jenkins version
pipelineid: jenkins/latest
{{- end }}