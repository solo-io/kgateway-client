# Create, Update & Delete EnterpriseKgatewayTrafficPolicy

This example demonstrates the fundamental operations for managing
`EnterpriseKgatewayTrafficPolicy` resources with `kgateway-client`, including
`Create`, `List`, `Update`, and `Delete`.

You can adapt this pattern to manage other Solo Enterprise for kgateway API
resources in this repository.

## Running this example

Make sure you have a Kubernetes cluster and `kubectl` is configured:

```sh
kubectl get nodes
```

Install the Solo Enterprise for kgateway CRDs first (see
`examples/README.md#cluster-prerequisites`) and verify
`EnterpriseKgatewayTrafficPolicy` is available:

```sh
kubectl get crd enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io
```

Compile this example on your workstation:

```sh
cd examples/create-update-delete-enterprisekgatewaytrafficpolicy
go build -o ./app
```

Run the application with your kubeconfig:

```sh
./app
# or specify kubeconfig/namespace
./app -kubeconfig=$HOME/.kube/config -namespace=default
```

The application runs these operations in order:

1. Create an `EnterpriseKgatewayTrafficPolicy` named `demo-enterprisekgateway-traffic-policy`.
2. Update it using `RetryOnConflict` by:
   - adding label `examples.solo.io/updated=true`
   - changing `spec.targetRefs[0].name` from `example-gateway` to `example-gateway-updated`
3. List `EnterpriseKgatewayTrafficPolicy` objects in the namespace.
4. Delete `demo-enterprisekgateway-traffic-policy`.

Each step pauses for <kbd>Return</kbd> so you can inspect state with `kubectl`.

Example output:

```text
Creating EnterpriseKgatewayTrafficPolicy...
Created EnterpriseKgatewayTrafficPolicy "demo-enterprisekgateway-traffic-policy".
-> Press Return key to continue.

Updating EnterpriseKgatewayTrafficPolicy...
Updated EnterpriseKgatewayTrafficPolicy (label examples.solo.io/updated="true", first targetRef.name="example-gateway-updated", generation=2).
-> Press Return key to continue.

Listing EnterpriseKgatewayTrafficPolicies in namespace "default":
 * demo-enterprisekgateway-traffic-policy (targetRef.name=example-gateway-updated, examples.solo.io/updated="true", entExtAuth.disable=true)
-> Press Return key to continue.

Deleting EnterpriseKgatewayTrafficPolicy...
Deleted EnterpriseKgatewayTrafficPolicy.
```

## Cleanup

If the program is interrupted before delete completes, remove the resource with:

```sh
kubectl delete enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io demo-enterprisekgateway-traffic-policy -n default
```
