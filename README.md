# kgateway-client

Go Clients for [Solo Enterprise for kgateway](https://www.solo.io/products/gloo-gateway).

**Note:** The [clients](https://github.com/kgateway-dev/kgateway/tree/main/pkg/client/clientset/versioned)
for the open source [kgateway project](https://kgateway.dev/) are **NOT** included in this repository.

> ⚠️ **This is an automatically published staged repository for [Solo Enterprise for kgateway](https://www.solo.io/products/gloo-gateway)**.
> This repository is read-only for importing, and not used for direct contributions.

## What's included

* The `clientset` package contains the clientset to access the Solo Enterprise for kgateway APIs.

## Versioning

Versioning aligns with Solo Enterprise for kgateway releases, e.g. `v2.x.y`.

### Branches and tags

A new tag is created for each increment in the minor or patch version number.
See [semver](http://semver.org/) for definitions of major, minor, and patch.

The HEAD of the main branch in kgateway-client will track the HEAD of the main branch in the Solo Enterprise for kgateway repo.

## How to get it

To get the latest version, use go1.16+ and fetch using the `go get` command. For example:

```sh
go get solo-io/kgateway-client@latest
```

To get a specific version, use go1.11+ and fetch the desired version using the `go get` command. For example:

```sh
go get solo-io/kgateway-client@v2.1.1
```

See [INSTALL.md](/INSTALL.md) for detailed instructions and troubleshooting.

## How to use it

If your application runs in a Pod in the cluster, please refer to the
in-cluster [example](examples/in-cluster-client-configuration), otherwise please
refer to the out-of-cluster [example](examples/out-of-cluster-client-configuration).

## Dependency management

For details on how to correctly use a dependency management for installing kgateway-client, please see [INSTALL.md](INSTALL.md).
