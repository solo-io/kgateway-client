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

	// PortalConditionResolvedRefs indicates whether references in the Portal spec were successfully resolved.
	PortalConditionResolvedRefs = "ResolvedRefs"

	// PortalReasonResolvedRefs indicates all referenced resources were resolved.
	PortalReasonResolvedRefs = "ResolvedRefs"
	// PortalReasonInvalid indicates one or more references on the Portal spec could not be resolved.
	PortalReasonInvalid = "Invalid"
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
	ApiProductRefs []PortalApiProductRef `json:"apiProductRefs,omitempty"` // nolint:kubeapilinter // arrayofstruct - required fields inherited from embedded NamespacedObjectReference

	// visibility controls the access requirements for the portal's API catalog.
	// +optional
	Visibility *PortalVisibility `json:"visibility,omitempty"`
}

// PortalApiProductRef references an ApiProduct exposed by the Portal and
// optionally attaches per-product VisibilityPolicy references
type PortalApiProductRef struct {
	upstreamshared.NamespacedObjectReference `json:",inline"`

	// visibilityPolicyRefs references VisibilityPolicy resources in the Portal's namespace
	// Multiple policies are unioned (OR across policies; OR across each policy's
	// groups; AND within a group).
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +optional
	VisibilityPolicyRefs []VisibilityPolicyReference `json:"visibilityPolicyRefs,omitempty"`
}

// PortalParametersReference identifies a PortalParameters resource in the same namespace.
type PortalParametersReference struct {
	// name is the name of the PortalParameters resource.
	// +required
	Name gwv1.ObjectName `json:"name"`
}

// VisibilityPolicyReference references a VisibilityPolicy in the same namespace as the referencing Portal.
type VisibilityPolicyReference struct {
	// name is the name of the VisibilityPolicy resource.
	// +required
	Name gwv1.ObjectName `json:"name"`
}

// PortalVisibility controls the access requirements for viewing the portal's API catalog
//
// +kubebuilder:validation:XValidation:rule="!(has(self.public) && self.public == true && has(self.visibilityPolicyRefs))",message="public cannot be true when visibilityPolicyRefs are defined"
type PortalVisibility struct {
	// public controls whether the portal's API catalog is publicly accessible without authentication.
	// When true, unauthenticated users can browse the API catalog.
	// When false (default), users must be logged in to view the catalog content.
	// Cannot be true when visibilityPolicyRefs is set, since claim-based gating
	// requires an authenticated user to have the necessary claims.
	// +optional
	Public *bool `json:"public,omitempty"`

	// visibilityPolicyRefs references VisibilityPolicy resources in the Portal's
	// namespace which are applied to every ApiProduct exposed by this Portal.
	// These policies are combined with per-product visibilityPolicyRefs declared on apiProductRefs[].
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +optional
	VisibilityPolicyRefs []VisibilityPolicyReference `json:"visibilityPolicyRefs,omitempty"`
}

// ClaimGroup is an AND of claims that must all be present on a subject's
// identity to satisfy this group.
type ClaimGroup struct {
	// claimGroup lists the claims that must all be present for this group
	// to grant access.
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +required
	Claims []Claim `json:"claimGroup"`
}

// Claim is a key/value pair to match against a subject's identity claims.
type Claim struct {
	// key is the claim name to match.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Key string `json:"key"`

	// value is the claim value to match.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Value string `json:"value"`
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
