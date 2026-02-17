# Authenticating inside the cluster

This example shows how to configure `kgateway-client` to authenticate to the
Kubernetes API from an application running inside the cluster.

It uses `rest.InClusterConfig()`, which reads the service account token mounted
in the Pod at `/var/run/secrets/kubernetes.io/serviceaccount`.

## Running this example

Build a container image from the repository root:

```sh
docker build -t kgateway-in-cluster:dev -f examples/in-cluster-client-configuration/Dockerfile .
```

If you are using a local cluster runtime such as kind, load the local image
into the cluster nodes:

```sh
kind load docker-image kgateway-in-cluster:dev
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
  --verb=get,list,watch,create \
  --resource=enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
  -n default

kubectl create rolebinding default-kgateway-client-view \
  --role=kgateway-client-view \
  --serviceaccount=default:default \
  -n default
```

Run the image in a Pod:

```sh
kubectl run --rm -i demo \
  --image=kgateway-in-cluster:dev \
  --image-pull-policy=IfNotPresent \
  --env="NAMESPACE=default"
```

If your cluster cannot access local images, push the image to a registry your
cluster can pull from and use that image reference in `kubectl run`.

Expected output (repeats every 10 seconds):

```text
Created EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy" in namespace "default"
There are 1 EnterpriseKgatewayTrafficPolicies in namespace "default"
Found EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy" in namespace "default"
```

Press <kbd>Ctrl</kbd>+<kbd>C</kbd> to stop.

## Cleanup

```sh
kubectl delete pod demo --ignore-not-found
kubectl delete rolebinding default-kgateway-client-view -n default --ignore-not-found
kubectl delete role kgateway-client-view -n default --ignore-not-found
kubectl delete enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
  example-enterprisekgateway-traffic-policy -n default --ignore-not-found
```
