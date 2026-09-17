package enterprisekgateway

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewaydestinationselectors,verbs=get;list;watch
// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewaydestinationselectors/status,verbs=get;update;patch

// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=".status.conditions[?(@.type=='Accepted')].status",description="Whether the destination selector was accepted by the control plane"

// EnterpriseKgatewayDestinationSelector selects a kgateway Backend by name for each request.
// Note: It must be the only backendRef in an HTTPRoute or GRPCRoute rule.
// Other route types are not supported and combining with multiple backendRefs is not supported.
//
// Any Backend in the selector's namespace can be selected. Adding a Backend does
// not require updating the selector. Cross-namespace destinations are not supported.
// Granting a route access to the selector grants access to all Backends in its namespace.
//
// Requests that select no Backend receive a 404. There is no fallback destination.
//
// Note: This API is experimental and may change in future releases.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprisekgateway,ekgw},path=enterprisekgatewaydestinationselectors,shortName=ekgds
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprisekgateway,app.kubernetes.io/name=enterprisekgatewaydestinationselector}
type EnterpriseKgatewayDestinationSelector struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines how to select a destination.
	// +required
	Spec EnterpriseKgatewayDestinationSelectorSpec `json:"spec"`

	// Status reports the selector's current state.
	// +optional
	Status EnterpriseKgatewayDestinationSelectorStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// EnterpriseKgatewayDestinationSelectorSpec defines how to select a destination.
type EnterpriseKgatewayDestinationSelectorSpec struct {
	// Header identifies the request header containing the Backend name.
	// +required
	Header DestinationHeader `json:"header"`
}

// DestinationHeader selects a Backend using a request header's value.
// The value must match a Backend name exactly.
type DestinationHeader struct {
	// Name is the request header containing the Backend name.
	//
	// The header is read during route resolution. Policies that set it, such as
	// JWT claimsToHeaders, must clear the route cache for the new value to take effect.
	// Users must ensure the header is sanitized.
	// +required
	Name gwv1.HeaderName `json:"name"`
}

// EnterpriseKgatewayDestinationSelectorStatus reports the selector's current state.
type EnterpriseKgatewayDestinationSelectorStatus struct {
	// Conditions reports the selector's status conditions.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +patchStrategy=merge
	// +patchMergeKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
type EnterpriseKgatewayDestinationSelectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnterpriseKgatewayDestinationSelector `json:"items"`
}
