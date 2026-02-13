package enterprisekgateway

import (
	upstream "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	"github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewayparameters,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=enterprisekgateway.solo.io,resources=enterprisekgatewayparameters/status,verbs=get;update;patch

// EnterpriseKgatewayParameters contains configuration that is used to dynamically
// provision Solo Enterprise for kgateway's data plane (Envoy proxy instance),
// and enterprise ExtAuth and RateLimiter extensions
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprisekgateway,ekgw},path=enterprisekgatewayparameters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprisekgateway,app.kubernetes.io/name=enterprisekgatewayparameters}
type EnterpriseKgatewayParameters struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the gateway parameters
	// +required
	Spec EnterpriseKgatewayParametersSpec `json:"spec"`

	// Status is the status of the gateway parameters
	// +optional
	Status EnterpriseKgatewayParametersStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// EnterpriseKgatewayParametersSpec defines the desired state of EnterpriseKgatewayParameters
type EnterpriseKgatewayParametersSpec struct {
	// Kubernetes configuration for the proxy.
	// +optional
	Kube *EnterpriseKgatewayKubernetesProxyConfig `json:"kube,omitempty"`
}

type EnterpriseKgatewayKubernetesProxyConfig struct {
	// Use a Kubernetes deployment as the proxy workload type. Currently, this is the only
	// supported workload type.
	//
	// +optional
	Deployment *upstream.ProxyDeployment `json:"deployment,omitempty"`

	// Configuration for the container running Envoy.
	// If agentgateway is enabled, the EnvoyContainer values will be ignored.
	//
	// +optional
	EnvoyContainer *upstream.EnvoyContainer `json:"envoyContainer,omitempty"`

	// Configuration for the container running the Secret Discovery Service (SDS).
	//
	// +optional
	SdsContainer *upstream.SdsContainer `json:"sdsContainer,omitempty"`

	// Configuration for the pods that will be created.
	//
	// +optional
	PodTemplate *upstream.Pod `json:"podTemplate,omitempty"`

	// Configuration for the Kubernetes Service that exposes the Envoy proxy over
	// the network.
	//
	// +optional
	Service *upstream.Service `json:"service,omitempty"`

	// Configuration for the Kubernetes ServiceAccount used by the Envoy pod.
	//
	// +optional
	ServiceAccount *upstream.ServiceAccount `json:"serviceAccount,omitempty"`

	// Configuration for the Istio integration.
	//
	// +optional
	Istio *upstream.IstioIntegration `json:"istio,omitempty"`

	// Configuration for the stats server.
	//
	// +optional
	Stats *upstream.StatsConfig `json:"stats,omitempty"`

	// OmitDefaultSecurityContext is used to control whether or not
	// `securityContext` fields should be rendered for the various generated
	// Deployments/Containers that are dynamically provisioned by the deployer.
	//
	// When set to true, no `securityContexts` will be provided and will left
	// to the user/platform to be provided.
	//
	// This should be enabled on platforms such as Red Hat OpenShift where the
	// `securityContext` will be dynamically added to enforce the appropriate
	// level of security.
	//
	// +optional
	OmitDefaultSecurityContext *bool `json:"omitDefaultSecurityContext,omitempty"`

	// GatewayParametersOverlays contains overlay fields for proxy deployment resources.
	// These allow applying strategic merge patches and creating HPA/PDB/VPA resources.
	upstream.GatewayParametersOverlays `json:",inline"`

	// SharedExtensions defines extensions that are shared across all Gateways of the same GatewayClass
	// +optional
	SharedExtensions *Extensions `json:"sharedExtensions,omitempty"`
}

type EnterpriseKgatewayParametersStatus struct{}

// EnterpriseKgatewayParametersList is a list of EnterpriseKgatewayParameters resources
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={enterprisekgateway,ekgw},path=enterprisekgatewayparameters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=enterprisekgateway,app.kubernetes.io/name=enterprisekgatewayparameters}
type EnterpriseKgatewayParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []EnterpriseKgatewayParameters `json:"items"`
}

// +kubebuilder:validation:Enum=Shared;Dedicated
type Mode string

type Extensions struct {
	// +optional
	ExtAuth *DeploymentConfiguration `json:"extauth,omitempty"`
	// +optional
	RateLimiter *DeploymentConfiguration `json:"ratelimiter,omitempty"`
	// +optional
	ExtCache *DeploymentConfiguration `json:"extCache,omitempty"`
}

// DeploymentConfiguration configures the Kubernetes Deployment.
type DeploymentConfiguration struct {
	// Enabled indicates whether the extension is enabled. If not enabled, then no resources for this extension will be deployed.
	// If the extension was previously enabled and then disabled, the deployed resources will be garbage collected, regardless of
	// whether any other configuration still depends on it.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	PodTemplate *upstream.Pod `json:"pod,omitempty"`

	// +optional
	Container *ContainerConfiguration `json:"container,omitempty"`

	// The number of desired pods.
	// If omitted, behavior will be managed by the K8s control plane, and will default to 1.
	// If you are using an HPA, make sure to not explicitly define this.
	// K8s reference: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#replicas
	//
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// The deployment strategy to use to replace existing pods with new
	// ones. The Kubernetes default is a RollingUpdate with 25% maxUnavailable,
	// 25% maxSurge.
	//
	// E.g., to recreate pods, minimizing resources for the rollout but causing downtime:
	// strategy:
	//   type: Recreate
	// E.g., to roll out as a RollingUpdate but with non-default parameters:
	// strategy:
	//   type: RollingUpdate
	//   rollingUpdate:
	//     maxSurge: 100%
	//
	// +optional
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty"`

	// DeploymentOverlay allows specifying overrides for the generated Deployment resource.
	// Use this for advanced customization not covered by the typed config fields,
	// such as adding initContainers, sidecars, or removing security contexts for OpenShift.
	// +optional
	DeploymentOverlay *shared.KubernetesResourceOverlay `json:"deploymentOverlay,omitempty"`

	// ServiceOverlay allows specifying overrides for the generated Service resource.
	// +optional
	ServiceOverlay *shared.KubernetesResourceOverlay `json:"serviceOverlay,omitempty"`

	// ServiceAccountOverlay allows specifying overrides for the generated ServiceAccount resource.
	// +optional
	ServiceAccountOverlay *shared.KubernetesResourceOverlay `json:"serviceAccountOverlay,omitempty"`

	// PodDisruptionBudget allows creating a PodDisruptionBudget for this extension.
	// If absent, no PDB is created. If present, a PDB is created with its selector
	// automatically configured to target the extension Deployment.
	// The metadata and spec fields from this overlay are applied to the generated PDB.
	// +optional
	PodDisruptionBudget *shared.KubernetesResourceOverlay `json:"podDisruptionBudget,omitempty"`

	// HorizontalPodAutoscaler allows creating a HorizontalPodAutoscaler for this extension.
	// If absent, no HPA is created. If present, an HPA is created with its scaleTargetRef
	// automatically configured to target the extension Deployment.
	// The metadata and spec fields from this overlay are applied to the generated HPA.
	// +optional
	HorizontalPodAutoscaler *shared.KubernetesResourceOverlay `json:"horizontalPodAutoscaler,omitempty"`

	// VerticalPodAutoscaler allows creating a VerticalPodAutoscaler for this extension.
	// If absent, no VPA is created. If present, a VPA is created with its targetRef
	// automatically configured to target the extension Deployment.
	// The metadata and spec fields from this overlay are applied to the generated VPA.
	// +optional
	VerticalPodAutoscaler *shared.KubernetesResourceOverlay `json:"verticalPodAutoscaler,omitempty"`
}

type ContainerConfiguration struct {
	// The image. See https://kubernetes.io/docs/concepts/containers/images for
	// details.
	//
	// +optional
	Image *upstream.Image `json:"image,omitempty"`

	// The security context for this container. Note OmitSecurityContext and
	// FloatingUserId, two related settings. See
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core
	// for details.
	//
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}
