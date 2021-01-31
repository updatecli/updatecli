source:
  name: "Get latest jenkins weekly version"
  kind: jenkins
  transformers:
    - prefix: "alpha-"
    - suffix: "-jdk11"
    - trimSuffix: "-jdk11"
    - suffix: "-jdk11"
    - replacer:
        from: "-jdk11"
        to: "-jdk15"
    - replacers:
        - from: "-jdk15"
          to: "-jdk17"
  spec:
    release: weekly
targets:
  targetID:
    name: "Update file file"
    kind: file
    spec:
      file: TODO
