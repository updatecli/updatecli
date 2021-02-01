---
source:
  kind: file
  spec:
    file: https://releases.hashicorp.com/terraform/0.14.5/terraform_0.14.5_SHA256SUMS
    line: 
      Hasincludes: 
        - "arm64"
conditions:
  condition0:
    name: condition0
    kind: file
    spec:
      file: https://releases.hashicorp.com/terraform/0.14.5/terraform_0.14.5_SHA256SUMS
      line: 
        Hasincludes:
          - ".zip"
  condition1:
    name: condition0
    kind: file
    spec:
      file: https://releases.hashicorp.com/terraform/0.14.5/terraform_0.14.5_SHA256SUMS
      line: 
        includes: 
          - "arm64"
      content: "d3cab7d777eec230b67eb9723f3b271cd43e29c688439e4c67e3398cdaf6406b  terraform_0.14.5_linux_arm64.zip"
targets:
  file1:
    name: target1
    kind: file
    spec:
      file: TODO
      line: linux_arm64.zip
