package waf

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// WAFPolicyConditionReady is the condition type indicating whether the WAFPolicy was accepted by the WAF server.
	WAFPolicyConditionReady = "Ready"

	// WAFPolicyReasonAccepted indicates the WAFPolicy was accepted.
	WAFPolicyReasonAccepted = "Accepted"
	// WAFPolicyReasonError indicates an error occurred when trying to process the WAFPolicy.
	WAFPolicyReasonError = "Error"
)

// +kubebuilder:rbac:groups=waf.solo.io,resources=wafpolicies,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=waf.solo.io,resources=wafpolicies/status,verbs=get;update;patch

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprise},path=wafpolicies,shortName=wafpol
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprise,app.kubernetes.io/name=wafpolicy}

// WAFPolicy contains Web Application Firewall configuration that can be applied to one or more routes.
// This configuration is consumed by an [External Processing Server](https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto)
// that all WAF-enabled traffic will pass through.
// If using the bundled WAF extproc server, all configuration provided in the WAFPolicy must be supported by the [Coraza](https://coraza.io/) WAF engine.
type WAFPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the WAF policy
	// +required
	Spec WAFPolicySpec `json:"spec"`

	// Status is the status of the WAF policy
	// +optional
	Status WAFPolicyStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// WAFPolicyStatus contains WAF server acceptance status and per-ancestor (traffic policy) attachment status.
type WAFPolicyStatus struct {
	// Conditions contains information about WAF server acceptance of the WAFPolicy.
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Ancestors contains attachment information for referencing resources.
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Ancestors []gwv1.PolicyAncestorStatus `json:"ancestors,omitempty"`
}

// WAFPolicySpec contains Web Application Firewall configuration.
type WAFPolicySpec struct {
	// CoreRuleSet contains settings for the OWASP CoreRuleSet.
	// If set, then the OWASP CoreRuleSet rules will be loaded.
	// The bundled WAF extproc server uses the [v4 CoreRuleSet rules](https://github.com/coreruleset/coreruleset/tree/v4.0/main/rules).
	// +optional
	CoreRuleSet *CoreRuleSet `json:"coreRuleSet,omitempty"`

	// RuleEngineSettings are settings to configure the WAF rule engine.
	// For an example Coraza-compatible rule engine settings file, see: https://github.com/corazawaf/coraza-coreruleset/blob/v4.23.0/rules/%40coraza.conf-recommended
	// +required
	RuleEngineSettings DirectiveSource `json:"ruleEngineSettings"`

	// CustomDirectives is a list of custom directives to apply.
	// Custom directives will be applied after the CoreRuleSet rules and settings (if enabled) and WAF rule engine settings,
	// and can be used to modify/exclude CoreRuleSet rules or add custom rules, for example.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +optional
	CustomDirectives []DirectiveSource `json:"customDirectives,omitempty"`

	// CustomInterventionResponse is a custom response to send when a request is blocked by WAF.
	// If not set, returns the status code defined by the WAF rule that was triggered.
	// +optional
	CustomInterventionResponse *CustomInterventionResponse `json:"customInterventionResponse,omitempty"`
}

// CustomInterventionResponse defines the response returned when WAF blocks a request or response.
type CustomInterventionResponse struct {
	// StatusCode overrides the HTTP status code returned when WAF blocks a request or response.
	// If not set, the status code defined by the triggered WAF rule is used.
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=599
	// +optional
	StatusCode *int32 `json:"statusCode,omitempty"`

	// Headers sets the response headers returned when WAF blocks a request or response.
	// If not set, no header modifications are made.
	// +optional
	Headers *CustomInterventionResponseHeaders `json:"headers,omitempty"`

	// Body overrides the response body returned when WAF blocks a request or response.
	// If not set, the response body is empty.
	// +optional
	Body *string `json:"body,omitempty"`
}

// CustomInterventionResponseHeaders defines headers returned when WAF blocks a request.
type CustomInterventionResponseHeaders struct {
	// SetHeaders is the list of headers to set on the response.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	SetHeaders []CustomInterventionResponseHeader `json:"setHeaders,omitempty"`
}

// CustomInterventionResponseHeader defines a single header returned when WAF blocks a request.
type CustomInterventionResponseHeader struct {
	// Name is the header name.
	// +kubebuilder:validation:MinLength=1
	// +required
	Name string `json:"name"`

	// Value is the header value.
	// +required
	Value string `json:"value"`
}

// DirectiveSource is a set of directives (e.g. rules or settings) to provide to the WAF engine.
// +kubebuilder:validation:ExactlyOneOf=inline;configMap
type DirectiveSource struct {
	// Inline specifies custom directives as an inline string.
	// +optional
	Inline *string `json:"inline,omitempty"`

	// ConfigMap is a reference to a ConfigMap containing custom directives.
	// +optional
	ConfigMap *ConfigMapRef `json:"configMap,omitempty"`
}

// ConfigMapRef contains a reference to a ConfigMap, and optionally specific keys within the ConfigMap.
type ConfigMapRef struct {
	// Name is the name of the ConfigMap.
	// +required
	Name string `json:"name"`

	// Namespace is the namespace of the ConfigMap.
	// +required
	Namespace string `json:"namespace"`

	// Keys is a list of keys to use from the ConfigMap.
	// If not set, the values of all keys in the ConfigMap will be used, in lexicographic order by key.
	// +optional
	// +kubebuilder:validation:MinItems=1
	Keys []string `json:"keys,omitempty"`
}

// CoreRuleSet specifies custom settings for the OWASP CoreRuleSet.
type CoreRuleSet struct {
	// Settings are settings that apply to the CoreRuleSet.
	// For an example Coraza-compatible CoreRuleSet settings file, see: https://github.com/corazawaf/coraza-coreruleset/blob/v4.23.0/rules/%40crs-setup.conf.example
	// +required
	Settings DirectiveSource `json:"settings"`
}

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprise},path=wafpolicies
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprise,app.kubernetes.io/name=wafpolicy}
type WAFPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WAFPolicy `json:"items"`
}
