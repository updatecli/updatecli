---
###
# This strategy will:
#
#  Source:
#  =======
#
#  Retrieve the helm version from its github release located on https://github.com/helm/helm
#    => v3.4.2
#
#  Conditions:
#  ===========
#
#  Then it will test one condition:
#  If the dockerfile 'docker/Dockerfile' is located on the git repository https://github.com/olblak/charts 
#  has the instruction ENV[1][0] set to "HELM_VERSION". ENV[1][0] is a custom syntax to represent 
#  a two-dimensional array where the first element represents a specific Dockerfile instruction identifier
#  starting from "0" at the beginning of the document, so we are looking for the second INSTRUCTION "ENV".
#  The second element represents an instruction argument position. In this case, we want to check that ENV key
#  is set to "HELM_VERSION"
#
#  Targets:
#  ========
#
#  If the condition is met, which is to be sure that the ENV key set to "HELM_VERSION" exist, then we'll 
#  are going to update its value if needed based on the version retrieved from the source.
#  The syntax is the same for the condition excepted that this time we are looking for ENV[1][1]
#  which means that the second argument of the second ENV instruction.
#
###


source:
  name: Get Latest helm release version
  kind: githubRelease
  spec:
    owner: "helm"
    repository: "helm"
    token: {{ requiredEnv .github.token }}
    username: olblak
    version: latest
conditions:
  isENVSet:
    name: Is ENV HELM_VERSION set
    kind: dockerfile
    spec:
      file: docker/Dockerfile
      Instruction: ENV[1][0]
      Value: "HELM_VERSION"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
targets:
  updateENVHELMVERSION:
    name: Update HELM_VERSION
    kind: dockerfile
    spec:
      file: docker/Dockerfile
      Instruction: ENV[1][1]
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "charts"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: "master"
