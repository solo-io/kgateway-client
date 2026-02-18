# Authenticating outside the cluster

This example shows how to configure `kgateway-client` to authenticate to the
Kubernetes API from an application running outside the cluster.

It uses your kubeconfig file (the same config `kubectl` uses) to initialize the
client.

## Running this example

Make sure `kubectl` points to a cluster:

```sh
kubectl get nodes
```

Install the Solo Enterprise for kgateway CRDs first (see
`examples/README.md#cluster-prerequisites`) and verify the
`EnterpriseKgatewayTrafficPolicy` CRD:

```sh
kubectl get crd enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io
```

Build and run the example:

```sh
cd examples/out-of-cluster-client-configuration
go build -o app .
./app
```

Optionally provide a specific kubeconfig file and namespace:

```sh
./app -kubeconfig=$HOME/.kube/config -namespace=default
```

Expected output (repeats every 10 seconds):

```text
Created EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy" in namespace "default"
There are 1 EnterpriseKgatewayTrafficPolicies in namespace "default"
Found EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy" in namespace "default"
```

Press <kbd>Ctrl</kbd>+<kbd>C</kbd> to stop.

## Cleanup

```sh
kubectl delete enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
  example-enterprisekgateway-traffic-policy -n default --ignore-not-found
```

If you installed CRDs only for this example, remove them with:

```sh
helm uninstall enterprise-kgateway-crds -n kgateway-system --ignore-not-found
```
