# kgateway-client

Go clients for [Solo Enterprise for kgateway](https://www.solo.io/products/gloo-gateway).

**Note:** The [clients](https://github.com/kgateway-dev/kgateway/tree/main/pkg/client/clientset/versioned)
for the open source [kgateway project](https://kgateway.dev/) are **not** included in this repository.

> This is an automatically published staged repository for [Solo Enterprise for kgateway](https://www.solo.io/products/gloo-gateway).
> The repository is read-only for importing and is not used for direct contributions.

## What's included

- The `clientset` package contains typed clients for Solo Enterprise for kgateway APIs.

## Versioning

Versioning aligns with Solo Enterprise for kgateway releases, for example `v2.x.y`.

The HEAD of `main` in `kgateway-client` tracks the synced HEAD of `main` from
Solo Enterprise for kgateway.

### Tagging Strategy

Tags in this repository are created by automation and use a leading `v`
prefix, for example `v2.2.0-beta.4`.

- The source sync workflow in `solo-io/gloo-gateway` includes `Source-Tag`
  metadata when a source release tag is being propagated.
- Source tag syncs are pushed to per-tag branches such as
  `sync/tag-2.2.0-beta.10`.
- This repo's `sync-source-tag-to-release-tag.yaml` workflow reads that
  metadata from the pushed `sync/tag-*` commit and creates the matching tag in
  `kgateway-client`.
- After creating or retargeting the tag, that workflow explicitly dispatches
  the validation workflows using the created tag as the ref.
- The tag is created from the pushed tag-branch commit in this repository, not
  from a commit in the source repository.
- If the source metadata does not include a leading `v`, the workflow adds it
  so published tags follow normal Go module tagging conventions.
- If a corrected sync for the same source tag is merged later, the workflow can
  retarget the existing tag to the newer tag-branch commit.
- The `sync/gloo-gateway-clientset` branch is a long-lived automation branch
  used to open or update sync PRs against `main`; it is intentionally reused
  across sync runs and is not auto-deleted after merges.
- `sync/tag-*` branches are per-tag automation branches used for validation and
  tag publication; they are not merged into `main`.

This means `main` can move ahead of the most recent published tag, while tags
identify specific synced release points that are safe to consume from Go.

## Build Validation Suite

Use the ref matrix test suite to validate compilation/test health for `main` and
all repository tags:

```sh
make validate-refs
```

The suite runs `go test ./...` in isolated git worktrees for each ref and
returns non-zero if any ref fails.

Use the example matrix test suite to compile/test each directory under
`examples/` for `main` and all repository tags:

```sh
make validate-examples
```

Use the example e2e suite to run live-cluster example validation (kind + CRDs)
for `main` and all repository tags:

```sh
make validate-examples-e2e
```

You can also pass explicit refs:

```sh
make validate-refs REFS="main v2.2.0-beta.2 v2.2.0-beta.4"
make validate-examples REFS="main v2.2.0-beta.2 v2.2.0-beta.4"
make validate-examples-e2e REFS="main v2.2.0-beta.4"
```

## How to get it

To get the latest version, use Go 1.16+:

```sh
go get github.com/solo-io/kgateway-client/v2@latest
```

To get a specific version:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.2.0-beta.4
```

See [INSTALL.md](INSTALL.md) for installation details and troubleshooting.

## How to use it

If your application runs in a Pod in the cluster, use the in-cluster
[example](examples/in-cluster-client-configuration).

If your application runs outside the cluster, use the out-of-cluster
[example](examples/out-of-cluster-client-configuration).

Additional examples are listed in [examples/README.md](examples/README.md).

## Dependency management

Using `kgateway-client` automatically locks your project to compatible versions
of the kgateway and Gateway API client libraries. Go's Minimum Version Selection
(MVS) ensures you get the correct versions without manual pinning.

### go.mod

You only need to require `kgateway-client` - transitive dependencies are
resolved automatically:

```go
module your-project

go 1.23

require github.com/solo-io/kgateway-client/v2 v2.2.0-beta.4
```

After `go mod tidy`, your `go.mod` will include the resolved dependencies:

```go
require (
    github.com/solo-io/kgateway-client/v2 v2.2.0-beta.4
    github.com/kgateway-dev/kgateway/v2 v2.3.0-beta.1  // version from kgateway-client
    k8s.io/apimachinery v0.34.3
    k8s.io/client-go v0.34.1
    sigs.k8s.io/gateway-api v1.4.1                     // version from kgateway-client
)
```

The `kgateway` and `gateway-api` versions are determined by `kgateway-client`.
When you upgrade `kgateway-client`, these dependencies update automatically:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.2.0-beta.5
go mod tidy
```

### Imports

Use the Solo Enterprise for kgateway client alongside the upstream kgateway
client and Gateway API client:

```go
import (
    // Solo Enterprise for kgateway client
    entkgatewayclient "github.com/solo-io/kgateway-client/v2/clientset/versioned"

    // Upstream kgateway client (version managed by kgateway-client)
    kgatewayclient "github.com/kgateway-dev/kgateway/v2/pkg/client/clientset/versioned"

    // Gateway API client (version managed by kgateway-client)
    gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"

    // Standard k8s
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/tools/clientcmd"
)
```

### Usage

```go
config, _ := clientcmd.BuildConfigFromFlags("", "~/.kube/config")

// Solo Enterprise for kgateway client
entClient, _ := entkgatewayclient.NewForConfig(config)
entPolicies, _ := entClient.EnterprisekgatewayEnterprisekgateway().
    EnterpriseKgatewayTrafficPolicies("default").
    List(ctx, metav1.ListOptions{})

// Upstream kgateway client
kgClient, _ := kgatewayclient.NewForConfig(config)
policies, _ := kgClient.GatewayKgateway().
    TrafficPolicies("default").
    List(ctx, metav1.ListOptions{})

// Gateway API client
gwClient, _ := gatewayclient.NewForConfig(config)
gateways, _ := gwClient.GatewayV1().Gateways("default").List(ctx, metav1.ListOptions{})
routes, _ := gwClient.GatewayV1().HTTPRoutes("default").List(ctx, metav1.ListOptions{})
```

For a runnable end-to-end example that uses all 3 clients together, see
[examples/multi-client-crud](examples/multi-client-crud).

For additional installation details and troubleshooting, see [INSTALL.md](INSTALL.md).
