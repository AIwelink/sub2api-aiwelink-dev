# AIWeLink Versioning Design

## Goal

Give every AIWeLink release an unambiguous version that identifies both the
official Sub2API baseline and the AIWeLink revision, while ensuring the built-in
update and rollback features only install AIWeLink releases.

## Version Model

AIWeLink versions use this format:

```text
<upstream-major>.<upstream-minor>.<upstream-patch>-<aiwelink-revision>
```

Examples:

- `0.1.170-1`: first AIWeLink release based on Sub2API `0.1.170`.
- `0.1.170-2.4`: a later AIWeLink revision on the same official baseline.
- `0.1.171-1`: first AIWeLink release after moving to Sub2API `0.1.171`.

The AIWeLink revision is one or more dot-separated positive integers. Zero,
empty segments, leading or trailing dots, and additional suffixes are invalid.
Release tags add the conventional `v` prefix, for example `v0.1.170-1`.

Two source files make the relationship explicit:

- `backend/cmd/server/VERSION` contains the complete AIWeLink version.
- `backend/cmd/server/UPSTREAM_VERSION` contains the official Sub2API baseline.

For the first release under this design they contain `0.1.170-1` and `0.1.170`.
The validation rule requires the full version to start with the exact upstream
version followed by `-` and a valid AIWeLink revision.

## Build Metadata

Both version values are embedded in the server binary and may be overridden by
linker flags in release builds. Local builds fall back to the embedded files.
The full AIWeLink version remains the primary application version used by CLI
output, public settings, update checks, Docker labels, and release artifacts.

The CLI output identifies the product as AIWeLink and includes the official
baseline. The public settings API and admin update API expose both values with
stable fields:

```json
{
  "version": "0.1.170-1",
  "upstream_version": "0.1.170"
}
```

Existing `version` and `current_version` fields remain available so current
clients continue to work. New `upstream_version` fields are additive.

## Version Resolution And Validation

The version resolver keeps the current precedence:

1. An exact release tag on the checked-out commit.
2. An explicitly supplied build argument or linker value where supported.
3. The committed version files.

Exact tags must match `v<upstream>-<revision>` or `<upstream>-<revision>`. A
validation script checks the following before builds and releases:

- `VERSION` and `UPSTREAM_VERSION` are single-line values without whitespace.
- `UPSTREAM_VERSION` is a three-part numeric version.
- `VERSION` contains the same official baseline.
- The AIWeLink revision contains only dot-separated positive integers.
- For tagged releases, the tag without `v` equals `VERSION`.

Invalid metadata fails the build with an error naming the mismatched values.

## Update And Rollback

The built-in update service changes its release source from
`Wei-Shaw/sub2api` to `AIwelink/sub2api-aiwelink-dev`. It never downloads an
official Sub2API binary directly.

Version comparison first compares the three official baseline components and
then compares every AIWeLink revision component numerically. Missing revision
components compare as zero, so `0.1.170-2.4` is newer than `0.1.170-2` and older
than `0.1.170-3`. Only valid AIWeLink release tags participate in update and
rollback selection. The service scans recent releases and selects the highest
valid version instead of assuming GitHub's publication order is version order.
Drafts, prereleases, malformed tags, and official-only tags such as
`v0.1.171` are ignored.

Update assets, changelog links, release links, and rollback downloads all come
from the AIWeLink repository. This prevents the updater from replacing the
customized application with an official Sub2API build.

## Frontend Presentation

The existing version badge remains compact. It displays the primary value as:

```text
AIWeLink v0.1.170-1
```

The expanded admin version panel adds:

```text
Based on Sub2API v0.1.170
```

Chinese and English translations are added. Update and rollback controls keep
their existing behavior but link to the AIWeLink repository and use AIWeLink
release versions and image names.

## Release And Docker Integration

The release workflow accepts only tags matching the validated AIWeLink format.
Although the hyphen makes this look like a SemVer prerelease, AIWeLink tags are
product release identifiers and GitHub Releases are explicitly published as
stable (`prerelease: false`). This keeps them visible to update checks and the
GitHub latest-release surface. The workflow does not rewrite `VERSION` after a
release. Version changes must arrive in a reviewed release PR, avoiding
automated direct pushes to protected long-lived branches.

Release binaries receive both `Version` and `UpstreamVersion` linker values.
Docker images and OCI labels use the complete AIWeLink version. The private
registry publishes these tags:

```text
docker.aiwelink.cc/sub2api-aiwelink-dev:0.1.170-1
docker.aiwelink.cc/sub2api-aiwelink-dev:latest
```

The immutable version tag is the deployment and rollback reference. `latest`
is only a convenience pointer. GitHub Release assets use the same complete
version, so binary, release, and container provenance agree.

## Git Workflow

Implementation starts from `aiwelink-dev` on
`chore/aiwelink-versioning`. The branch is pushed and merged through a PR into
`aiwelink-dev`; neither `aiwelink-dev` nor `main` receives a direct development
push. Future official synchronization branches update `UPSTREAM_VERSION`, reset
the AIWeLink revision to `1`, pass validation, and use a Merge Commit into
`aiwelink-dev`.

Production releases continue through an `aiwelink-dev -> main` PR using a Merge
Commit. The release tag is created on the merged production commit.

## Failure Handling

- Invalid local version metadata stops builds before producing artifacts.
- A malformed remote release is skipped rather than treated as an update.
- Failure to contact the AIWeLink release API preserves the existing cached
  result and warning behavior.
- An update with missing or invalid assets fails before binary replacement.
- Existing source-build behavior remains available, but source builds do not
  offer an in-place binary update unless the current release rules permit it.

## Testing

Focused tests cover:

- accepted and rejected version formats;
- baseline and multi-part AIWeLink revision comparisons;
- exact-tag resolution and file fallback;
- build metadata propagation through dependency injection;
- public and admin API responses containing both versions;
- AIWeLink repository selection and malformed release filtering;
- frontend store compatibility and version badge rendering;
- release validation for matching and mismatching tags.

Verification includes backend unit tests, frontend tests, type checking, linting,
frontend production build, version validation, and a Docker build with the
complete AIWeLink version.
