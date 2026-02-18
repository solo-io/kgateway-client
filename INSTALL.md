# Installing kgateway-client

## Using the latest version

If you want the latest published version of this library, use Go 1.16+ and run:

```sh
go get github.com/solo-io/kgateway-client/v2@latest
```

This records `github.com/solo-io/kgateway-client/v2` in your module dependencies.

## Using a specific version

If you want a specific release, fetch that tag directly:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.1.1
```

`kgateway-client` versions align with Solo Enterprise for kgateway releases
(`v2.x.y`).

## Verifying installation

Run these commands in your module:

```sh
go mod tidy
go list -m github.com/solo-io/kgateway-client/v2
```

## Troubleshooting

### Wrong module path

If you see package resolution errors, confirm you are using the full module path:

```go
import "github.com/solo-io/kgateway-client/v2/clientset/versioned"
```

### Older Go versions

If you use an older Go toolchain and `@latest` resolves unexpectedly, pin an
explicit version:

```sh
go get github.com/solo-io/kgateway-client/v2@v2.1.1
```

### Dependency conflicts

If your project pulls incompatible Kubernetes dependencies, inspect your module graph:

```sh
go mod graph | grep "github.com/solo-io/kgateway-client/v2@"
```

Then pin compatible versions in your `go.mod` as needed.
