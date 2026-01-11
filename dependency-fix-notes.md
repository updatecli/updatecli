# Dependency fix notes (yaml-jsonpath fork)

## Why each change was made

### go.mod
- Added a `replace` for `github.com/vmware-labs/yaml-jsonpath` pointing to `github.com/helm-unittest/yaml-jsonpath v0.4.0` with a short comment.
  - Reason: the fork is maintained, but its `go.mod` still declares the module path as `github.com/vmware-labs/yaml-jsonpath`, so we must keep the original module path in `require` and use `replace` to redirect it. This avoids the "module declares its path as ... but was required as ..." error.
  - The comment documents the module path mismatch and why the replace exists.

- Indirect dependency cleanup performed by `go mod tidy`:
  - Removed `github.com/dprotaso/go-yit` and `github.com/onsi/ginkgo` (indirect).
    - Reason: they are no longer needed by the module graph after the yaml-jsonpath fork replacement and tidy.
  - Updated `github.com/sergi/go-diff` from the pseudo-version to `v1.4.0` (indirect).
    - Reason: the forked yaml-jsonpath depends on `github.com/sergi/go-diff v1.4.0`, so tidy selected that version to satisfy the new graph.

### go.sum
- Added checksum entries for `github.com/helm-unittest/yaml-jsonpath v0.4.0`.
  - Reason: required by the new replace target; Go needs checksums for module integrity.

- Removed checksum entries for `github.com/vmware-labs/yaml-jsonpath v0.3.2`.
  - Reason: the module graph now resolves to the fork at `v0.4.0` via replace, so the old version is no longer used.

- Removed checksums for indirect modules that are no longer in the graph (e.g., `github.com/dprotaso/go-yit`, `github.com/onsi/ginkgo`, and older transitive versions).
  - Reason: `go mod tidy` prunes unused entries to keep the sums consistent with the current dependency graph.

### pkg/plugins/resources/yaml/condition.go
- Switched YAML import from `gopkg.in/yaml.v3` to `go.yaml.in/yaml/v3`.
  - Reason: `github.com/helm-unittest/yaml-jsonpath` uses `go.yaml.in/yaml/v3` types in its API. The change aligns the node types passed into `yamlpath.Find`, fixing the compile error caused by mixing two different yaml.v3 packages.

### pkg/plugins/resources/yaml/source.go
- Same import change as above (`go.yaml.in/yaml/v3`).
  - Reason: ensures the YAML nodes passed to yaml-jsonpath are the exact types expected by the forked module.

### pkg/plugins/resources/yaml/target_yamlpath.go
- Same import change as above (`go.yaml.in/yaml/v3`).
  - Reason: keeps the yamlpath target implementation consistent with the fork?s type expectations and fixes the compile-time mismatch.

## Summary
- The fork `helm-unittest/yaml-jsonpath` keeps the original module path (`github.com/vmware-labs/yaml-jsonpath`) in its own `go.mod`, so the correct fix is a `replace` while keeping imports as `github.com/vmware-labs/yaml-jsonpath`.
- The fork depends on `go.yaml.in/yaml/v3`, so internal yamlpath calls must use that package to avoid type mismatches.
- `go mod tidy` updated and pruned indirect dependencies to reflect the new module graph.
