# Versioning and stability

Updatecli does not break your manifests on purpose. When something needs to change, the old
syntax is deprecated and keeps working. It is never removed without a year of warning.

## Semantic versioning

Updatecli uses [SemVer-style](https://semver.org/) version numbers with an extended compatibility policy. Deprecated manifest syntax and CLI flags may be removed in minor releases after at least 12 months of notice.

| Bump | Means |
|---|---|
| Minor | New plugins, new fields, new flags. May also drop a syntax that has been deprecated for over a year (see below). |
| Patch | Bug fixes. |
| Major | Reserved for a redesign of Updatecli itself. Routine deprecation removals do not need one. |

Upgrading is safe as long as your manifests do not rely on syntax listed on the
[deprecations page](https://www.updatecli.io/docs/help/deprecations/).

## What is stable

Two surfaces are covered by this promise:

* **The manifest API**: top level keys, resource `kind` names, their spec fields, the meaning
  of their values, and the templating functions available in a manifest.
* **The CLI**: command names, flag names and their behaviour, and `UPDATECLI_*` environment
  variables.

The manifest API is the project's first priority. A manifest that works today is expected to
keep working.

## What is not stable

* **The Go code** (`github.com/updatecli/updatecli/pkg/...`). It is public because Go offers no
  other way to structure a program, not because it is an API. Packages move and signatures
  change in any release. If you import Updatecli as a library, pin an exact version and expect
  work on every upgrade.
* **Anything behind `--experimental`.** It may change or be withdrawn at any time.
* **Log and report output.** It is written for people to read, not for scripts to parse.

## Deprecation

When a manifest key or a CLI flag is superseded:

1. The old form keeps working and warns on every run, naming its replacement.
2. It is listed on <https://www.updatecli.io/docs/help/deprecations/>.
3. It stays supported for **at least 12 months** from that announcement.

Once both conditions are met, meaning a year has passed and the deprecation is documented, the
old form may be dropped in a regular release. A major version bump is not required for this, so
the deprecation page is the page to watch.

`updatecli manifest upgrade --save` rewrites most deprecated syntax for you.

## Pinning

A manifest can declare the minimum version it needs, which turns a confusing failure on an
older binary into a clear one:

```yaml
name: Example
version: '0.118.0'
sources:
  #...
targets:
  #...
```

## Found a break?

If an upgrade breaks a manifest that worked before, it is a bug, not an intentional change.
Please open an issue with the manifest, both versions, and the output of each.
