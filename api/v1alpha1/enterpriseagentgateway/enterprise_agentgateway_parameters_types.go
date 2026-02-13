package enterpriseagentgateway

import (
	upstreamagent "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/agentgateway"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups=enterpriseagentgateway.solo.io,resources=enterpriseagentgatewayparameters,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=enterpriseagentgateway.solo.io,resources=enterpriseagentgatewayparameters/status,verbs=get;update;patch

// EnterpriseAgentgatewayParameters contains configuration that is used to dynamically
// provision the agentgateway data plane with enterprise extensions like ExtAuth and RateLimiter.
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterpriseagentgateway,eagw},path=enterpriseagentgatewayparameters,shortName=eagpar
// +kubebuilder:subresource:status
type EnterpriseAgentgatewayParameters struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec EnterpriseAgentgatewayParametersSpec `json:"spec"`

	// +optional
	Status EnterpriseAgentgatewayParametersStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

type EnterpriseAgentgatewayParametersSpec struct {
	upstreamagent.AgentgatewayParametersSpec `json:",inline"`

	// SharedExtensions defines extensions that are shared across all Gateways of the same GatewayClass
	// +optional
	SharedExtensions *AgentgatewayExtensions `json:"sharedExtensions,omitempty"`

	// CA is the certificate authority configuration for Istio integration.
	// +optional
	CA *CA `json:"ca,omitempty"`

	// IstioClusterId is the ID of the cluster that this Istiod instance resides (default `Kubernetes`).
	// +optional
	IstioClusterId *string `json:"istioClusterId,omitempty"`
}

type AgentgatewayExtensions struct {
	// +optional
	ExtAuth *ExtensionDeployment `json:"extauth,omitempty"`
	// +optional
	RateLimiter *ExtensionDeployment `json:"ratelimiter,omitempty"`
	// +optional
	ExtCache *ExtensionDeployment `json:"extCache,omitempty"`
}

// The current conditions of the EnterpriseAgentgatewayParameters. This is not currently implemented.
type EnterpriseAgentgatewayParametersStatus struct{}

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterpriseagentgateway,eagw},path=enterpriseagentgatewayparameters
// +kubebuilder:subresource:status
type EnterpriseAgentgatewayParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []EnterpriseAgentgatewayParameters `json:"items"`
}

// ExtensionDeployment configures an extension deployment (extauth, ratelimiter, extcache).
type ExtensionDeployment struct {
	// Enabled indicates whether the extension is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the container image configuration.
	// +optional
	Image *upstreamagent.Image `json:"image,omitempty"`

	// Resources are the compute resources required by this container.
	// See https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Env are additional environment variables for the container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Deployment allows specifying overrides for the generated Deployment resource.
	// Use this for advanced customization not covered by the typed config fields,
	// such as adding initContainers, sidecars, or removing security contexts for OpenShift.
	// +optional
	Deployment *upstreamagent.KubernetesResourceOverlay `json:"deployment,omitempty"`

	// Service allows specifying overrides for the generated Service resource.
	// +optional
	Service *upstreamagent.KubernetesResourceOverlay `json:"service,omitempty"`

	// ServiceAccount allows specifying overrides for the generated ServiceAccount resource.
	// +optional
	ServiceAccount *upstreamagent.KubernetesResourceOverlay `json:"serviceAccount,omitempty"`

	// PodDisruptionBudget allows creating a PodDisruptionBudget for this extension.
	// If absent, no PDB is created. If present, a PDB is created with its selector
	// automatically configured to target the extension Deployment.
	// The metadata and spec fields from this overlay are applied to the generated PDB.
	// +optional
	PodDisruptionBudget *upstreamagent.KubernetesResourceOverlay `json:"podDisruptionBudget,omitempty"`

	// HorizontalPodAutoscaler allows creating a HorizontalPodAutoscaler for this extension.
	// If absent, no HPA is created. If present, an HPA is created with its scaleTargetRef
	// automatically configured to target the extension Deployment.
	// The metadata and spec fields from this overlay are applied to the generated HPA.
	// +optional
	HorizontalPodAutoscaler *upstreamagent.KubernetesResourceOverlay `json:"horizontalPodAutoscaler,omitempty"`
}

// CA is the certificate authority configuration for Istio integration.
type CA struct {
	// Address is the discovery address of the certificate authority.
	// Default is https://istiod.istio-system.svc:15012
	// +optional
	Address *string `json:"address,omitempty"`

	// TrustDomain is the trust domain of the certificate authority.
	// +optional
	TrustDomain *string `json:"trustDomain,omitempty"`
}
