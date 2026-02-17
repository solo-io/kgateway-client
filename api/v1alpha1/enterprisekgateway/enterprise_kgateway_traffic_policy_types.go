package enterprisekgateway

import (
	upstream "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/solo-io/kgateway-client/api/v1alpha1/shared"
)

// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewaytrafficpolicies,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewaytrafficpolicies/status,verbs=get;update;patch

// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=".status.ancestors[*].conditions[?(@.type=='Accepted')].status",description="Solo Enterprise for kgateway Traffic policy acceptance status"
// +kubebuilder:printcolumn:name="Attached",type=string,JSONPath=".status.ancestors[*].conditions[?(@.type=='Attached')].status",description="Solo Enterprise for kgateway Traffic policy attachment status"

// EnterpriseKgatewayTrafficPolicy is a traffic policy that can be applied to a route
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprisekgateway,ekgw},path=enterprisekgatewaytrafficpolicies,shortName=ekgtp
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprisekgateway,app.kubernetes.io/name=enterprisekgatewaytrafficpolicy}
type EnterpriseKgatewayTrafficPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the traffic policy
	// +required
	Spec EnterpriseKgatewayTrafficPolicySpec `json:"spec"`

	// Status is the status of the traffic policy
	// +optional
	Status gwv1.PolicyStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// EnterpriseKgatewayTrafficPolicySpec defines the desired state of EnterpriseKgatewayTrafficPolicy
//
// +kubebuilder:validation:AtMostOneOf=extAuth;entExtAuth
// +kubebuilder:validation:AtMostOneOf=rateLimit;entRateLimit
// +kubebuilder:validation:AtMostOneOf=transformation;entTransformation
type EnterpriseKgatewayTrafficPolicySpec struct {
	upstream.TrafficPolicySpec `json:",inline"`

	// EntRateLimit defines the Enterprise rate limit configuration for the traffic policy
	// +optional
	EntRateLimit *EntRateLimit `json:"entRateLimit,omitempty"`

	// EntExtAuth defines the Enterprise external authorization configuration for the traffic policy
	// +optional
	EntExtAuth *EntExtAuth `json:"entExtAuth,omitempty"`

	// EntTransformation defines the Enterprise transformation configuration for the traffic policy
	// +optional
	EntTransformation *EntTransformation `json:"entTransformation,omitempty"`

	// EntJWT allows for configuration of JWT authentication
	// +optional
	EntJWT *StagedJWT `json:"entJWT,omitempty"`

	// EntRBAC provides config for RBAC rules based on JWT claims resulting from authentication with `entJWT` configs
	// +optional
	EntRBAC *EntRBAC `json:"entRBAC,omitempty"`

	// EntWAF defines the Web Application Firewall configuration
	// +optional
	EntWAF *EntWAF `json:"entWAF,omitempty"`
}

type EntRateLimit struct {
	// Global rate limit configuration
	// +required
	Global GlobalRateLimit `json:"global"`
}

type GlobalRateLimit struct {
	// ExtensionRef references a GatewayExtension that provides the global rate limit service.
	// If not set, defaults to the rate limit service named 'rate-limit' in the same namespace as
	// the Solo Enterprise for kgateway control plane. In this case no reference grant is required.
	// +optional
	ExtensionRef *upstreamshared.NamespacedObjectReference `json:"extensionRef,omitempty"`

	// RateLimitConfigRefs is a list of references to the RateLimitConfig resources containing the
	// rate limit configurations.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +required
	RateLimitConfigRefs []shared.RateLimitConfigRef `json:"rateLimitConfigRefs"`
}

// EnterpriseKgatewayTrafficPolicyList is a list of EnterpriseKgatewayTrafficPolicy resources
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprisekgateway,ekgw},path=enterprisekgatewaytrafficpolicies
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprisekgateway,app.kubernetes.io/name=enterprisekgatewaytrafficpolicy}
type EnterpriseKgatewayTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnterpriseKgatewayTrafficPolicy `json:"items"`
}

// +kubebuilder:validation:ExactlyOneOf=authConfigRef;disable
type EntExtAuth struct {
	// AuthConfigRef references the AuthConfig we want the external-auth server will use to make auth
	// decisions.
	// +optional
	AuthConfigRef *shared.AuthConfigRef `json:"authConfigRef,omitempty"`

	// ExtensionRef references a GatewayExtension that provides the external authorization service.
	// If not set, defaults to the provisioned ext-auth-service for the GatewayClass of the parent Gateway
	// this policy is being used in.
	// Reference grants are not required for cross-namespace extension references.
	// +optional
	ExtensionRef *upstreamshared.NamespacedObjectReference `json:"extensionRef,omitempty"`

	// Disable all external authorization filters.
	// Can be used to disable external authorization policies applied at a higher level in the config hierarchy.
	// +optional
	Disable *upstreamshared.PolicyDisable `json:"disable,omitempty"`
}

// +kubebuilder:validation:ExactlyOneOf=wafPolicyRef;disable
// +kubebuilder:validation:AtMostOneOf=wafServer;disable
type EntWAF struct {
	// WAFPolicyRef references the WAFPolicy we want to use for the traffic policy
	// +optional
	WAFPolicyRef *shared.WAFPolicyRef `json:"wafPolicyRef,omitempty"`

	// WAFServer is a reference to the external processing gRPC service that will be used to process requests
	// when WAF is enabled.
	// If not set, defaults to the extproc service named 'waf-extproc' in the same namespace as
	// the Solo Enterprise for kgateway control plane.
	// +optional
	WAFServer *gwv1.BackendObjectReference `json:"wafServer,omitempty"`

	// Disable WAF.
	// Can be used to disable WAF policies applied at a higher level in the config hierarchy.
	// +optional
	Disable *upstreamshared.PolicyDisable `json:"disable,omitempty"`
}
