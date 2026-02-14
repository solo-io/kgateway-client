# kgateway-client Examples

This directory contains examples that cover common `kgateway-client` use cases.

## Cluster prerequisites

Examples that talk to a live cluster require the Solo Enterprise for kgateway
CRDs, including
`enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io`.

Set a version that matches your `kgateway-client` release:

```sh
KGW_VERSION=2.1.1
```

Install CRDs (minimum required for the examples):

```sh
helm upgrade --install enterprise-kgateway-crds \
  oci://us-docker.pkg.dev/solo-public/enterprise-kgateway/charts/enterprise-kgateway-crds \
  --version "${KGW_VERSION}" \
  --namespace kgateway-system \
  --create-namespace
```

Verify the CRD is present:

```sh
kubectl get crd enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io
```

CRD-only install is sufficient for these client examples and does not require a
product license key.

Optional: run the full Enterprise kgateway Helm install in HA mode (2
controller replicas). A valid Solo Enterprise for kgateway license key is
required for this install. Install the CRD chart above first.

```sh
export LICENSE_KEY="<your-license-key>"
```

```sh
helm upgrade --install enterprise-kgateway \
  oci://us-docker.pkg.dev/solo-public/enterprise-kgateway/charts/enterprise-kgateway \
  --version "${KGW_VERSION}" \
  --namespace kgateway-system \
  --set licensing.licenseKey="${LICENSE_KEY}" \
  --set controller.replicaCount=2 \
  --set controller.podDisruptionBudget.minAvailable=1
```

Optionally, use an existing Kubernetes secret for the license key:

```sh
kubectl create secret generic enterprise-kgateway-license \
  --namespace kgateway-system \
  --from-literal=license-key="${LICENSE_KEY}"

helm upgrade --install enterprise-kgateway \
  oci://us-docker.pkg.dev/solo-public/enterprise-kgateway/charts/enterprise-kgateway \
  --version "${KGW_VERSION}" \
  --namespace kgateway-system \
  --set licensing.createSecret=false \
  --set licensing.secretName=enterprise-kgateway-license \
  --set controller.replicaCount=2 \
  --set controller.podDisruptionBudget.minAvailable=1
```

Reference docs:

- [Installation](https://docs.solo.io/kgateway/2.1.x/setup/install/)
- [Helm values reference](https://docs.solo.io/kgateway/2.1.x/reference/helm/kgateway/)
- [Licensing](https://docs.solo.io/gateway/2.1.x/setup/product-licensing/)

## Configuration

- [**Authenticate in cluster**](./in-cluster-client-configuration): Configure a
  client from inside a Kubernetes Pod.
- [**Authenticate out of cluster**](./out-of-cluster-client-configuration):
  Configure a client from outside the cluster using kubeconfig.

## Basics

- [**Managing resources with API**](./create-update-delete-enterprisekgatewaytrafficpolicy):
  Create, get, update, list, and delete an `EnterpriseKgatewayTrafficPolicy`
  resource.

## Testing

- [**Fake Client**](./fake-client): Use the generated fake clientset in unit
  tests.
