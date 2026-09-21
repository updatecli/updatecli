# Versioning and stability

Updatecli does not break your manifests on purpose. When something needs to change, the old
syntax is deprecated and keeps working.

## Semantic versioning

Updatecli follows [semantic versioning](https://semver.org/).

| Bump | Means |
|---|---|
| Patch | Bug fixes. Nothing is added, nothing is taken away. |
| Minor | New plugins, new fields, new flags. Everything that worked before still works. |
| Major | The only release allowed to remove a stable surface, rename it, or change what it means. |

## What is stable

Two surfaces are covered by this promise:

* **The manifest API**: top level keys, resource `kind` names, their spec fields, the meaning
  of their values, and the templating functions available in a manifest.
* **The CLI**: command names, flag names and their behaviour, and `UPDATECLI_*` environment
  variables.

The manifest API is the project's first priority. A manifest that works today is expected to
keep working.

## What is not stable

* **The Go code** (`github.com/updatecli/updatecli/pkg/...`). This project is not a library, any Go code may change in any release.
* **Anything behind `--experimental`.** It may change or be withdrawn at any time.
* **Log and report output.** It is written for people to read, not for scripts to parse.

## Deprecation

When a manifest key or a CLI flag is superseded:

1. The old form keeps working and warns on every run, naming its replacement.
2. It is listed on <https://www.updatecli.io/docs/help/deprecations/>.

## Minimum Version

A manifest can declare the minimum version it needs using the key `version`.

```yaml
name: This is an example manifest
version: '0.118.0'
sources:
  #...
targets:
  #... 
```

## Found a break?

If an upgrade within a major version breaks a manifest that worked before, it is a bug,
not an intentional change. Please open an issue with the manifest, both versions, and the
output of each.
