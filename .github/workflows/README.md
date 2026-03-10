# Workflows

This directory contains the GitHub Actions workflows used to validate
`kgateway-client`, support automated syncs from
`solo-io/gloo-gateway`, and propagate release tags.

## Validation workflows

### `ref-validation.yaml`

Runs the lightweight validation suite.

- On `pull_request` to `main`:
  - `validate` tests the latest repo tag if one exists, otherwise `main`
  - `validate-examples` runs the default example compile/test matrix (`main` and all tags)
- On `push` to `main`:
  - validates `main`
- On `push` to `sync/tag-*`:
  - validates only the pushed branch commit
- On `push` to `v*` tags:
  - validates only the pushed tag
- On `workflow_dispatch`:
  - validates `main` and all tags by default
  - accepts optional space-separated `refs`

The jobs call:

- `./hack/test-ref-matrix.sh`
- `./hack/test-example-matrix.sh`

These checks are intended to catch compile or basic test regressions without
standing up a cluster.

### `example-e2e-validation.yaml`

Runs the live-cluster example e2e suite.

- On `pull_request` to `main`:
  - tests only the PR head SHA
- On `push` to `main`:
  - tests `main` and all tags
- On `push` to `sync/tag-*`:
  - tests only the pushed branch commit
- On `push` to `v*` tags:
  - tests `main` and all tags
- On `workflow_dispatch`:
  - tests `main` and all tags by default
  - accepts optional space-separated `refs`

The job installs `kubectl`, `helm`, and `kind`, then runs:

- `./hack/test-example-e2e-matrix.sh`

That script creates a kind cluster per ref, installs the upstream `kgateway`
and Gateway API CRDs, generates the enterprise CRD manifests from the checked
out ref's API types, skips `examples/fake-client`, and runs the live examples.

## Sync workflows

### `sync-pr-ci-automerge.yaml`

Validates and merges trusted sync PRs opened against `main`.

- `sync-pr-ci`
  - runs on `pull_request`
  - only for branch `sync/gloo-gateway-clientset`
  - compile-checks `./api/...` and `./clientset/...`
- `merge-sync-pr`
  - runs on `pull_request_target`
  - only for branch `sync/gloo-gateway-clientset`
  - only when the PR author matches `SYNC_PR_AUTHOR_LOGIN`
  - uses the configured sync GitHub App token
  - checks out the base repository and delegates the wait-and-merge logic to
    `./hack/merge-sync-pr.sh`
  - waits for the expected sync PR checks to pass
  - merges the PR directly as the sync app with administrator privileges so the
    configured ruleset bypass applies
  - leaves the sync branch in place with `--delete-branch=false`

Required repo configuration:

- Variable: `SYNC_APP_ID`
- Variable: `SYNC_PR_AUTHOR_LOGIN`
- Secret: `SYNC_APP_PRIVATE_KEY`

The sync app must also have the permissions and ruleset bypass needed to merge
trusted sync PRs.

### `sync-source-tag-to-release-tag.yaml`

Creates or retargets repo tags from pushed `sync/tag-*` branches.

- Runs on `push` to `sync/tag-*`
- Reads `Source-Tag:` metadata from the pushed commit message/body
- Verifies that the `Source-Tag` matches the pushed branch name
- Creates a matching target tag
- Adds a leading `v` when the source tag does not already include it
- Retargets an existing tag if the tag already exists and points at a different
  commit

This workflow is paired with the source repo sync workflow, which includes
`Source-Tag` metadata in sync tag branch commits when a source tag is being
propagated.

## Notes

- The fixed sync branch is `sync/gloo-gateway-clientset`.
- The branch is intentionally reused across sync runs and is not auto-deleted on
  merge.
- Per-tag sync branches follow the `sync/tag-*` pattern, are validated on push,
  and are not auto-merged to `main`.
- The validation and e2e workflows are independent of the source sync workflow;
  they only define the checks and post-merge behavior in this repo.
