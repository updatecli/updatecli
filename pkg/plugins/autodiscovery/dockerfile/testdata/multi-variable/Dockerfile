# Single variable case - SHOULD generate a pipeline
ARG node_version=18.0.0
FROM node:${node_version}

# Multiple variables case in a single keyword position - SHOULD NOT generate a pipeline
ARG base_version=20
ARG variant=alpine
FROM node:${base_version}-${variant}

# Multiple variables case a different keyword positions - SHOULD generate a pipeline
ARG my_version=20
ARG my_platform=linux/amd64
FROM --platform=${my_platform} node:${my_version}
