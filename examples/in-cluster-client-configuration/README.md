# Authenticating inside the cluster

This example shows how to configure `kgateway-client` to authenticate to the
Kubernetes API from an application running inside the cluster.

It uses `rest.InClusterConfig()`, which reads the service account token mounted
in the Pod at `/var/run/secrets/kubernetes.io/serviceaccount`.

## Running this example

Build a container image from the repository root:

```sh
docker build -t kgateway-in-cluster -f examples/in-cluster-client-configuration/Dockerfile .
```

Install the Solo Enterprise for kgateway CRDs first (see
`examples/README.md#cluster-prerequisites`) and verify the
`EnterpriseKgatewayTrafficPolicy` CRD:

```sh
kubectl get crd enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io
```

Grant the service account permission to read `EnterpriseKgatewayTrafficPolicy`
resources in the target namespace (`default` in this example):

```sh
kubectl create role kgateway-client-view \
  --verb=get,list,watch \
  --resource=enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
  -n default

kubectl create rolebinding default-kgateway-client-view \
  --role=kgateway-client-view \
  --serviceaccount=default:default \
  -n default
```

Run the image in a Pod:

```sh
kubectl run --rm -i demo --image=kgateway-in-cluster --env="NAMESPACE=default"
```

Expected output (repeats every 10 seconds):

```text
There are 1 EnterpriseKgatewayTrafficPolicies in namespace "default"
EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy" in namespace "default" not found
```

Press <kbd>Ctrl</kbd>+<kbd>C</kbd> to stop.

## Cleanup

```sh
kubectl delete pod demo --ignore-not-found
kubectl delete rolebinding default-kgateway-client-view -n default --ignore-not-found
kubectl delete role kgateway-client-view -n default --ignore-not-found
```
