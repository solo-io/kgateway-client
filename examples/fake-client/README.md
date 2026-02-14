# Fake Client Example

This example demonstrates how to use the generated fake clientset from
`kgateway-client` in unit tests.

It covers:

- Creating a fake client with seeded `EnterpriseKgatewayTrafficPolicy` objects.
- Performing `Create`, `Get`, `Update`, `List`, and `Delete` operations.
- Asserting Kubernetes-style API errors (for example, `IsNotFound`).

## Running

```sh
go test -v ./examples/fake-client
```
