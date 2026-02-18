package enterpriseagentgateway

import (
	upstreamagent "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/agentgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/solo-io/kgateway-client/v2/api/v1alpha1/shared"
)

// +kubebuilder:rbac:groups=enterpriseagentgateway.solo.io,resources=enterpriseagentgatewaypolicies,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=enterpriseagentgateway.solo.io,resources=enterpriseagentgatewaypolicies/status,verbs=get;update;patch

// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=".status.ancestors[*].conditions[?(@.type=='Accepted')].status",description="Solo Enterprise for agentgateway policy acceptance status"
// +kubebuilder:printcolumn:name="Attached",type=string,JSONPath=".status.ancestors[*].conditions[?(@.type=='Attached')].status",description="Solo Enterprise for agentgateway policy attachment status"

// EnterpriseAgentgatewayPolicy is a Solo Enterprise for agentgateway policy
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterpriseagentgateway,eagw},path=enterpriseagentgatewaypolicies,shortName=eagpol
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=gloo,app.kubernetes.io/name=enterpriseagentgatewaypolicy}
type EnterpriseAgentgatewayPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the agentgateway policy
	// +required
	Spec EnterpriseAgentgatewayPolicySpec `json:"spec"`

	// Status is the status of the agentgateway policy
	// +optional
	Status gwv1.PolicyStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// EnterpriseAgentgatewayPolicySpec defines the desired state of EnterpriseAgentgatewayPolicy
// +kubebuilder:validation:ExactlyOneOf=targetRefs;targetSelectors
// +kubebuilder:validation:XValidation:rule="has(self.traffic) || has(self.frontend) || has(self.backend)",message="At least one of traffic, frontend, or backend must be provided."
// +kubebuilder:validation:XValidation:rule="!has(self.backend) || !has(self.backend.mcp) || ((!has(self.targetRefs) || !self.targetRefs.exists(t, t.kind == 'Service')) && (!has(self.targetSelectors) || !self.targetSelectors.exists(t, t.kind == 'Service')))",message="backend.mcp may not be used with a Service target"
// +kubebuilder:validation:XValidation:rule="!has(self.backend) || !has(self.backend.ai) || ((!has(self.targetRefs) || !self.targetRefs.exists(t, t.kind == 'Service')) && (!has(self.targetSelectors) || !self.targetSelectors.exists(t, t.kind == 'Service')))",message="backend.ai may not be used with a Service target"
// +kubebuilder:validation:XValidation:rule="has(self.frontend) && has(self.targetRefs) ? self.targetRefs.all(t, t.kind == 'Gateway' && !has(t.sectionName)) : true",message="the 'frontend' field can only target a Gateway"
// +kubebuilder:validation:XValidation:rule="has(self.frontend) && has(self.targetSelectors) ? self.targetSelectors.all(t, t.kind == 'Gateway' && !has(t.sectionName)) : true",message="the 'frontend' field can only target a Gateway"
// +kubebuilder:validation:XValidation:rule="has(self.traffic) && has(self.targetRefs) ? self.targetRefs.all(t, t.kind in ['Gateway', 'HTTPRoute', 'XListenerSet']) : true",message="the 'traffic' field can only target a Gateway, XListenerSet, or HTTPRoute"
// +kubebuilder:validation:XValidation:rule="has(self.traffic) && has(self.targetSelectors) ? self.targetSelectors.all(t, t.kind in ['Gateway', 'HTTPRoute', 'XListenerSet']) : true",message="the 'traffic' field can only target a Gateway, XListenerSet, or HTTPRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) && has(self.traffic) && has(self.traffic.phase) && self.traffic.phase == 'PreRouting' ? self.targetRefs.all(t, t.kind in ['Gateway', 'XListenerSet']) : true",message="the 'traffic.phase=PreRouting' field can only target a Gateway or XListenerSet"
// +kubebuilder:validation:XValidation:rule="has(self.targetSelectors) && has(self.traffic) && has(self.traffic.phase) && self.traffic.phase == 'PreRouting' ? self.targetSelectors.all(t, t.kind in ['Gateway', 'XListenerSet']) : true",message="the 'traffic.phase=PreRouting' field can only target a Gateway or XListenerSet"
type EnterpriseAgentgatewayPolicySpec struct {
	// targetRefs specifies the target resources by reference to attach the policy to.
	//
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(r, (r.kind == 'Service' && r.group == '') || (r.kind == 'AgentgatewayBackend' && r.group == 'agentgateway.dev') || (r.kind in ['Gateway', 'HTTPRoute'] && r.group == 'gateway.networking.k8s.io') || (r.kind == 'XListenerSet' && r.group == 'gateway.networking.x-k8s.io'))",message="targetRefs may only reference Gateway, HTTPRoute, XListenerSet, Service, or AgentgatewayBackend resources"
	// +kubebuilder:validation:XValidation:message="Only one Kind of targetRef can be set on one policy",rule="self.all(l1, !self.exists(l2, l1.kind != l2.kind))"
	// +optional
	TargetRefs []upstreamshared.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs,omitempty"`

	// targetSelectors specifies the target selectors to select resources to attach the policy to.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(r, (r.kind == 'Service' && r.group == '') || (r.kind == 'AgentgatewayBackend' && r.group == 'agentgateway.dev') || (r.kind in ['Gateway', 'HTTPRoute'] && r.group == 'gateway.networking.k8s.io') || (r.kind == 'XListenerSet' && r.group == 'gateway.networking.x-k8s.io'))",message="targetRefs may only reference Gateway, HTTPRoute, XListenerSet, Service, or AgentgatewayBackend resources"
	// +kubebuilder:validation:XValidation:message="Only one Kind of targetRef can be set on one policy",rule="self.all(l1, !self.exists(l2, l1.kind != l2.kind))"
	// +optional
	TargetSelectors []upstreamshared.LocalPolicyTargetSelectorWithSectionName `json:"targetSelectors,omitempty"`

	// frontend settings for incoming traffic. Targets Gateways only. Field-level merge.
	// +optional
	Frontend *EnterpriseAgentgatewayPolicyFrontend `json:"frontend,omitempty"`

	// traffic settings for request processing. Targets Gateway/Listener(ListenerSet)/Route/RouteRule. Field-level merge; precedence: Gateway < Listener < Route < RouteRule.
	// +optional
	Traffic *EnterpriseAgentgatewayPolicyTraffic `json:"traffic,omitempty"`

	// backend settings for connecting to destinations. Targets Gateway/Listener(ListenerSet)/Route/RouteRule/Service/Backend. Applies per backend; precedence: Gateway < Listener < Route < RouteRule < Backend/Service.
	// +optional
	Backend *EnterpriseAgentgatewayPolicyBackend `json:"backend,omitempty"`
}

type EnterpriseAgentgatewayPolicyFrontend struct {
	upstreamagent.Frontend `json:",inline"`
}

// +kubebuilder:validation:AtLeastOneOf=tcp;tls;http;auth;mcp;ai;tokenExchange
type EnterpriseAgentgatewayPolicyBackend struct {
	upstreamagent.BackendSimple `json:",inline"`

	// ai specifies settings for AI workloads. This is only applicable when connecting to a Backend of type 'ai'.
	// +optional
	AI *upstreamagent.BackendAI `json:"ai,omitempty"`

	// mcp specifies settings for MCP workloads. This is only applicable when connecting to a Backend of type 'mcp'.
	// +optional
	MCP *upstreamagent.BackendMCP `json:"mcp,omitempty"`
	// Perform token exchange before sending requests to the backend.
	// For this to work, token exchange settings need to be pre-configured for the dataplane.
	// +optional
	TokenExchange *TokenExchangeCfg `json:"tokenExchange,omitempty"`
}

type TokenExchangeMode string

const (
	// TokenExchangeModeElicitationOnly indicates that token exchange should not be performed, just eliciting the user to do the auth flow.
	TokenExchangeModeElicitationOnly TokenExchangeMode = "ElicitationOnly"
	// TokenExchangeModeExchangeOnly indicates that elicitation errors should not be returned to the user.
	TokenExchangeModeExchangeOnly TokenExchangeMode = "ExchangeOnly"
)

type TokenExchangeCfg struct {
	// Mode indicates the token exchange mode to use.
	// Defaults to empty, which allows for both Elicitation and Exchange
	// +optional
	// +kubebuilder:validation:Enum=ElicitationOnly;ExchangeOnly
	Mode *TokenExchangeMode `json:"mode,omitempty"`

	// Oidc configures which OAuth/OIDC client Secret should be used for elicitations and token storage
	// associated with this backend.
	// The token exchange server keys provider selection by the "resource" value used in elicitation APIs.
	// When this policy targets an AgentgatewayBackend, the backend name (and also "namespace/name") is
	// used as the resource key for mapping to this SecretName.
	// +optional
	Oidc *TokenExchangeOidcConfig `json:"oidc,omitempty"`
}

type TokenExchangeOidcConfig struct {
	// SecretName is the name of the Kubernetes Secret in the agentgateway installation namespace
	// containing the OAuth client configuration (authorize_url, access_token_url, client_id, etc.).
	// This field is required unless Mode is set to ExchangeOnly.
	// +required
	SecretName string `json:"secretName"`
}

// EnterpriseAgentgatewayPolicyTraffic defines the desired state of EnterpriseAgentgatewayPolicyTraffic
// +kubebuilder:validation:AtLeastOneOf=transformation;extProc;extAuth;rateLimit;cors;csrf;headerModifiers;hostRewrite;timeouts;retry;authorization;jwtAuthentication;basicAuthentication;apiKeyAuthentication;entRateLimit;entExtAuth
// +kubebuilder:validation:AtMostOneOf=extAuth;entExtAuth
// +kubebuilder:validation:AtMostOneOf=rateLimit;entRateLimit
type EnterpriseAgentgatewayPolicyTraffic struct {
	upstreamagent.Traffic `json:",inline"`

	// RateLimit defines the Enterprise rate limit configuration for the traffic policy
	// +optional
	EntRateLimit *EnterpriseAgentgatewayRateLimit `json:"entRateLimit,omitempty"`

	// ExtAuth defines the Enterprise external authorization configuration for the traffic policy
	// +optional
	EntExtAuth *EnterpriseAgentgatewayExtAuth `json:"entExtAuth,omitempty"`
}

type EnterpriseAgentgatewayRateLimit struct {
	// Global rate limit configuration
	// +required
	Global AgwGlobalRateLimit `json:"global"`
}

type AgwGlobalRateLimit struct {
	// domain for the limit; defaults to 'solo.io'.
	// +optional
	Domain *ShortString `json:"domain,omitempty"`

	// BackendRef to the rate limit service (Service or Backend). Defaults to 'rate-limit' in the control-plane namespace.
	// +optional
	BackendRef *gwv1.BackendObjectReference `json:"backendRef,omitempty"`

	// RateLimitConfigRefs references RateLimitConfig resources.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +required
	RateLimitConfigRefs []shared.RateLimitConfigRef `json:"rateLimitConfigRefs"`
}

type EnterpriseAgentgatewayExtAuth struct {
	// AuthConfigRef used by the external-auth server.
	// +optional
	AuthConfigRef *shared.AuthConfigRef `json:"authConfigRef,omitempty"`

	// BackendRef to the external authorization service (Service or Backend). Defaults to the provisioned ext-auth-service; cross-namespace grants not required.
	// +optional
	BackendRef *gwv1.BackendObjectReference `json:"backendRef,omitempty"`
}

// EnterpriseAgentgatewayPolicyList is a list of EnterpriseAgentgatewayPolicy resources
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterpriseagentgateway,eagw},path=enterpriseagentgatewaypolicies
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=gloo,app.kubernetes.io/name=enterpriseagentgatewaypolicy}
type EnterpriseAgentgatewayPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnterpriseAgentgatewayPolicy `json:"items"`
}

// +kubebuilder:validation:MaxLength=256
type ShortString = string
