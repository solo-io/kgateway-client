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

### Branches and tags

A new tag is created for each minor or patch increment.
See [semver](https://semver.org/) for definitions of major, minor, and patch.

The HEAD of the `main` branch in `kgateway-client` tracks the HEAD of `main` in
Solo Enterprise for kgateway.

## How to get it

To get the latest version, use Go 1.16+:

```sh
go get github.com/solo-io/kgateway-client/v2@latest
```

To get a specific version:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.1.1
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

require github.com/solo-io/kgateway-client/v2 v2.1.1
```

After `go mod tidy`, your `go.mod` will include the resolved dependencies:

```go
require (
    github.com/solo-io/kgateway-client/v2 v2.1.1
    github.com/kgateway-dev/kgateway/v2 v2.2.0      // version from kgateway-client
    k8s.io/apimachinery v0.35.1
    k8s.io/client-go v0.35.1
    sigs.k8s.io/gateway-api v1.4.1                  // version from kgateway-client
)
```

The `kgateway` and `gateway-api` versions are determined by `kgateway-client`.
When you upgrade `kgateway-client`, these dependencies update automatically:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.2.0
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

For installation troubleshooting, see [INSTALL.md](INSTALL.md).
