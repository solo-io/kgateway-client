# Multi-Client CRUD Example

This example demonstrates using all 3 clientsets in the same program:

- Gateway API client (`sigs.k8s.io/gateway-api/pkg/client/clientset/versioned`)
- Upstream kgateway client (`github.com/kgateway-dev/kgateway/v2/pkg/client/clientset/versioned`)
- Solo Enterprise for kgateway client (`github.com/solo-io/kgateway-client/v2/clientset/versioned`)

It performs create/get/update/list/delete operations for:

- `Gateway` and `HTTPRoute` (Gateway API)
- `TrafficPolicy` (upstream kgateway)
- `EnterpriseKgatewayTrafficPolicy` (Solo Enterprise for kgateway)

## Running this example

Make sure `kubectl` points to a cluster:

```sh
kubectl get nodes
```

Install the required CRDs first:

```sh
KGW_VERSION=2.1.1

# Enterprise kgateway CRDs (includes EnterpriseKgatewayTrafficPolicy and TrafficPolicy)
helm upgrade --install enterprise-kgateway-crds \
  oci://us-docker.pkg.dev/solo-public/enterprise-kgateway/charts/enterprise-kgateway-crds \
  --version "${KGW_VERSION}" \
  --namespace kgateway-system \
  --create-namespace

# Gateway API CRDs (version aligned with this repo's dependency)
GW_API_VERSION=v1.4.1

kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/"${GW_API_VERSION}"/standard-install.yaml
```

Then verify the resources used in this example are available:

```sh
kubectl get crd gateways.gateway.networking.k8s.io
kubectl get crd httproutes.gateway.networking.k8s.io
kubectl get crd trafficpolicies.gateway.kgateway.dev
kubectl get crd enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io
```

Build and run:

```sh
cd examples/multi-client-crud
go build -o app .
./app
```

Optionally provide a specific kubeconfig and namespace:

```sh
./app -kubeconfig=$HOME/.kube/config -namespace=default
```

Sample output:

```text
Creating resources across all 3 clientsets...
Created Gateway "demo-gateway"
Created HTTPRoute "demo-http-route"
Created TrafficPolicy "demo-kgateway-traffic-policy"
Created EnterpriseKgatewayTrafficPolicy "demo-enterprisekgateway-traffic-policy"
Updating each resource...
Updated Gateway "demo-gateway"
Updated HTTPRoute "demo-http-route"
Updated TrafficPolicy "demo-kgateway-traffic-policy"
Updated EnterpriseKgatewayTrafficPolicy "demo-enterprisekgateway-traffic-policy"
Listing resources...
Gateway count in namespace: 1
HTTPRoute count in namespace: 1
TrafficPolicy count in namespace: 1
EnterpriseKgatewayTrafficPolicy count in namespace: 1
Deleting resources...
Deleted HTTPRoute "demo-http-route"
Deleted TrafficPolicy "demo-kgateway-traffic-policy"
Deleted EnterpriseKgatewayTrafficPolicy "demo-enterprisekgateway-traffic-policy"
Deleted Gateway "demo-gateway"
Done.
```

## Cleanup

If the program is interrupted before delete completes, remove resources with:

```sh
kubectl delete httproutes.gateway.networking.k8s.io demo-http-route -n default --ignore-not-found
kubectl delete gateways.gateway.networking.k8s.io demo-gateway -n default --ignore-not-found
kubectl delete trafficpolicies.gateway.kgateway.dev demo-kgateway-traffic-policy -n default --ignore-not-found
kubectl delete enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io demo-enterprisekgateway-traffic-policy -n default --ignore-not-found
```

If you installed CRDs only for this example, remove them with:

```sh
GW_API_VERSION=v1.4.1

helm uninstall enterprise-kgateway-crds -n kgateway-system --ignore-not-found
kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/"${GW_API_VERSION}"/standard-install.yaml --ignore-not-found
```
