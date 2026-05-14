package v1alpha1

import (
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Portal condition types and reasons reuse Gateway API constants for consistency.
const (
	// PortalConditionProgrammed is the condition type indicating if portal resources are deployed.
	// Reuses Gateway API's Programmed condition for consistency.
	PortalConditionProgrammed = string(gwv1.GatewayConditionProgrammed)

	// PortalConditionApiProductStatus is the condition type indicating the overall health
	// of the Portal's ApiProduct references: whether they resolved, are permitted, and are healthy.
	PortalConditionApiProductStatus = "ApiProductStatus"

	// PortalReasonApiProductsReady indicates all resolved ApiProducts are fully healthy.
	PortalReasonApiProductsReady = "ApiProductsReady"
	// PortalReasonPartiallyAccepted indicates at least one resolved ApiProduct is healthy,
	// but not all are fully accepted (some have Ready: False, Reason: PartiallyAccepted,
	// or some references failed to resolve).
	PortalReasonPartiallyAccepted = "PartiallyAccepted"
	// PortalReasonApiProductError indicates all resolved ApiProducts have Ready: False.
	PortalReasonApiProductError = "ApiProductError"
	// PortalReasonNoApiProducts indicates no ApiProduct references are configured or none resolved.
	PortalReasonNoApiProducts = "NoApiProducts"
)

// +kubebuilder:rbac:groups=portal.solo.io,resources=portals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=portal.solo.io,resources=portals/status,verbs=get;update;patch

// Portal represents a developer portal deployment that references a PortalParameters resource
// for operational configuration.
//
// The Portal controller deploys only the backend web server. Users are responsible for:
//
//   - Deploying the frontend UI (see dev-portal-starter or custom implementation)
//   - Creating Gateway and HTTPRoutes to expose the portal
//   - Configuring authentication (AuthConfig, traffic policies) if needed
//
// See: https://docs.solo.io/kgateway/latest/portal/guides/frontend-portal/
//
// To prevent naming collisions, a Portal cannot be named "{name}" if an existing Gateway
// in the same namespace is already named "portal-{name}".
// This is because both resources will attempt to manage data plane components (such as
// Deployments and Services) using the "portal-{name}" identifier.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={portal},path=portals
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=gloo,app.kubernetes.io/name=portals}
// +kubebuilder:printcolumn:name="Programmed",type=string,JSONPath=`.status.conditions[?(@.type=='Programmed')].status`
// +kubebuilder:printcolumn:name="ProductsStatusReason",type=string,JSONPath=`.status.conditions[?(@.type=='ApiProductStatus')].reason`
type Portal struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of the Portal
	// +required
	Spec PortalSpec `json:"spec"`

	// status defines the observed state of the Portal
	// +optional
	Status PortalStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// PortalSpec defines the desired state of Portal
type PortalSpec struct {
	// parametersRef references the PortalParameters resource in the same namespace
	// that defines the operational configuration for this portal.
	// If not specified, the controller uses default in-memory configuration.
	// +optional
	ParametersRef *PortalParametersReference `json:"parametersRef,omitempty"`

	// apiProductRefs lists the ApiProducts visible in this portal.
	// Only the referenced ApiProducts will be exposed to end users via the portal's API catalog.
	// If empty, no API products are exposed.
	// +listType=atomic
	// +optional
	ApiProductRefs []upstreamshared.NamespacedObjectReference `json:"apiProductRefs,omitempty"`

	// visibility controls the access requirements for the portal's API catalog.
	// +optional
	Visibility *PortalVisibility `json:"visibility,omitempty"`
}

// PortalParametersReference identifies a PortalParameters resource in the same namespace.
type PortalParametersReference struct {
	// name is the name of the PortalParameters resource.
	// +required
	Name gwv1.ObjectName `json:"name"`
}

// PortalVisibility controls the access requirements for viewing the portal's API catalog
type PortalVisibility struct {
	// public controls whether the portal's API catalog is publicly accessible without authentication.
	// When true, unauthenticated users can browse the API catalog.
	// When false (default), users must be logged in to view the catalog content.
	// +optional
	Public *bool `json:"public,omitempty"`
}

// PortalStatus defines the observed state of Portal
type PortalStatus struct {
	// conditions represent the latest available observations of the Portal state
	// +optional
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// PortalList contains a list of Portal resources
// +kubebuilder:object:root=true
type PortalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Portal `json:"items"`
}
