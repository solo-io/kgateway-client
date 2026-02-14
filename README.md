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
go get github.com/solo-io/kgateway-client@latest
```

To get a specific version:

```sh
go get github.com/solo-io/kgateway-client@v2.1.1
```

See [INSTALL.md](INSTALL.md) for installation details and troubleshooting.

## How to use it

If your application runs in a Pod in the cluster, use the in-cluster
[example](examples/in-cluster-client-configuration).

If your application runs outside the cluster, use the out-of-cluster
[example](examples/out-of-cluster-client-configuration).

Additional examples are listed in [examples/README.md](examples/README.md).

## Dependency management

For dependency installation details, see [INSTALL.md](INSTALL.md).
