# README

A scaffold bundle containing an Updatecli policy and default values/templates. Use this
bundle as a starting point for publishing a reusable Updatecli policy.

## REQUIREMENTS

- `updatecli` CLI installed (recommended: latest stable release).
- Access to an OCI registry (set the `OCI_REGISTRY` environment variable).
- Optional: `docker` (or another OCI client) for logging in and pushing bundles.

Before running, update these files to match your environment:
- `Policy.yaml` — policy metadata and `version`.
- `values.d/default.yaml` — default policy inputs.
- `updatecli.d/default.yaml` — pipeline configuration and SCM settings.

## QUICK USAGE

Show the policy (local):
```sh
updatecli manifest show --config updatecli.d --values values.yaml
```

Show the policy (registry):
```sh
updatecli manifest show $OCI_REGISTRY/<policy-name>:v1.0.0
```

Validate / dry-run (local):
```sh
updatecli diff --config updatecli.d --values values.yaml
```

Apply / enforce (local):
```sh
updatecli apply --config updatecli.d --values values.yaml
```

Notes:
- Any `--values` files specified at runtime override the defaults in this bundle.
- Replace `<policy-name>` and `v1.0.0` with your registry path and version.

## AUTHENTICATION

Authenticate with your OCI registry before publishing or pulling private bundles:
```sh
docker login "$OCI_REGISTRY"
```

`OCI_REGISTRY` can be any OCI-compliant registry (for example: Zot, Docker Hub, GitHub Container Registry).

## PUBLISH

Publish the bundle to an OCI registry (the `version` field in `Policy.yaml` controls the tag):
```sh
updatecli manifest push \
  --config updatecli.d \
  --values values.yaml \
  --policy Policy.yaml \
  --tag "$OCI_REGISTRY/<policy-name>" \
  .
```

After publishing, reference the bundle by tag:
```sh
updatecli manifest show "$OCI_REGISTRY/<policy-name>:v1.0.0"
```

## NEXT STEPS & LINKS

- Official docs: https://www.updatecli.io
- Compose docs (orchestrating multiple policies): https://www.updatecli.io/docs/core/compose/
- Sharing & reuse: https://www.updatecli.io/docs/core/shareandreuse/

## CONTRIBUTING

This README was generated from a template. Suggestions and issues are welcome:
https://github.com/updatecli/updatecli/issues