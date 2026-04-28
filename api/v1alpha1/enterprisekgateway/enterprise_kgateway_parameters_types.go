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
	ExtAuth *ExtAuthConfiguration `json:"extauth,omitempty"`
	// +optional
	RateLimiter *RateLimiterConfiguration `json:"ratelimiter,omitempty"`
	// +optional
	ExtCache *DeploymentConfiguration `json:"extCache,omitempty"`

	// WAF configures the WAF server.
	// +optional
	WAF *WAFConfiguration `json:"waf,omitempty"`
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

// RateLimiterConfiguration configures the RateLimit server deployment.
type RateLimiterConfiguration struct {
	DeploymentConfiguration `json:",inline"`

	// Redis configures the Redis connection for the RateLimit server.
	// When specified, the RateLimit server connects to this Redis instance
	// instead of the managed ext-cache Redis.
	//
	// +optional
	Redis *RedisClientConfig `json:"redis,omitempty"`

	// ServiceAccountName sets the serviceAccountName on the generated RateLimit
	// Deployment. Use this to attach a ServiceAccount configured for AWS
	// credentials, such as through IRSA or EKS Pod Identity, for AWS
	// ElastiCache IAM authentication.
	//
	// +optional
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
}

// ExtAuthConfiguration configures the ExtAuth server deployment.
type ExtAuthConfiguration struct {
	DeploymentConfiguration `json:",inline"`

	// SessionRedis configures the server-level default Redis connection for
	// ExtAuth session storage (OAuth2/OIDC). When specified, individual
	// AuthConfig CRs do not need to repeat connection details in their
	// RedisOptions fields, though per-AuthConfig overrides are still supported.
	//
	// +optional
	SessionRedis *RedisClientConfig `json:"sessionRedis,omitempty"`

	// ServiceAccountName sets the serviceAccountName on the generated ExtAuth
	// Deployment. Use this to attach a ServiceAccount configured for AWS
	// credentials, such as through IRSA or EKS Pod Identity, for AWS
	// ElastiCache IAM authentication.
	//
	// +optional
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
}

// RedisClientConfig is a reusable Redis connection configuration structure
// shared by both the RateLimit and ExtAuth sections of EnterpriseKgatewayParameters.
//
// +kubebuilder:validation:XValidation:rule="!has(self.certs) || (has(self.socketType) && self.socketType == 'tls')",message="certs can only be set when socketType is 'tls'"
type RedisClientConfig struct {
	// Address is the Redis server address in "host:port" format, or a Unix
	// socket path when socketType is "unix".
	// Examples: "redis.example.com:6379", "my-redis.default.svc.cluster.local:6379", "/var/run/redis/redis.sock"
	//
	// +required
	Address string `json:"address"`

	// DB is the Redis database index.
	// Defaults to 0 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	DB *int32 `json:"db,omitempty"`

	// SocketType specifies the connection type: "tcp", "tls", or "unix".
	// Defaults to "tcp" if not specified.
	//
	// +optional
	// +kubebuilder:validation:Enum=tcp;tls;unix
	SocketType *string `json:"socketType,omitempty"`

	// Clustered enables Redis Cluster mode.
	// When true, the client uses a ClusterClient that handles MOVED/ASK redirects.
	//
	// +optional
	Clustered *bool `json:"clustered,omitempty"`

	// Certs configures TLS certificates for the Redis connection.
	// Only applicable when SocketType is "tls".
	//
	// +optional
	Certs *RedisCerts `json:"certs,omitempty"`

	// Auth configures authentication for the Redis connection.
	// Exactly one of secretRef or aws should be specified.
	// If not specified, no authentication is used.
	//
	// +optional
	Auth *RedisAuth `json:"auth,omitempty"`

	// Connection configures connection pool and timeout tuning parameters.
	// When not specified, Redis client library defaults apply.
	//
	// +optional
	Connection *RedisConnectionConfig `json:"connection,omitempty"`
}

// RedisCerts configures TLS certificates for the Redis connection.
type RedisCerts struct {
	// CACertSecretRef references a Kubernetes Secret containing the CA certificate
	// for verifying the Redis server's TLS certificate.
	//
	// +optional
	CACertSecretRef *corev1.SecretReference `json:"caCertSecretRef,omitempty"`

	// CACertKey is the key within the Secret that contains the CA certificate in PEM format.
	// Defaults to "ca.crt" if not specified.
	//
	// +optional
	CACertKey *string `json:"caCertKey,omitempty"`
}

// RedisAuth configures authentication for a Redis connection.
// This is a discriminated union: specify exactly one of secretRef or aws.
// +kubebuilder:validation:ExactlyOneOf=secretRef;aws
type RedisAuth struct {
	// SecretRef configures static credential authentication using a Kubernetes Secret
	// containing username and password.
	//
	// +optional
	SecretRef *RedisSecretAuth `json:"secretRef,omitempty"`

	// AWS configures AWS ElastiCache IAM authentication.
	// No static credentials are stored; tokens are generated at runtime
	// using AWS credentials available to the pod.
	//
	// +optional
	AWS *RedisAWSAuth `json:"aws,omitempty"`
}

// RedisSecretAuth configures Redis authentication from a Kubernetes Secret.
type RedisSecretAuth struct {
	// Name is the name of the Kubernetes Secret containing Redis credentials.
	//
	// +required
	Name string `json:"name"`

	// Namespace is the namespace of the Secret. If not specified, the Secret
	// is assumed to be in the same namespace as the extension Deployment.
	//
	// +optional
	Namespace *string `json:"namespace,omitempty"`

	// PasswordKey is the key in the Secret that contains the Redis password.
	// Defaults to "password" if not specified.
	//
	// +optional
	PasswordKey *string `json:"passwordKey,omitempty"`

	// UsernameKey is the key in the Secret that contains the Redis username.
	// Defaults to "username" if not specified.
	//
	// +optional
	UsernameKey *string `json:"usernameKey,omitempty"`
}

// RedisAWSAuth configures AWS ElastiCache IAM authentication.
// Requires the pod to have AWS credentials, such as through IRSA or EKS Pod Identity.
type RedisAWSAuth struct {
	// Region is the AWS region of the ElastiCache cluster.
	//
	// +required
	Region string `json:"region"`

	// ClusterName is the ElastiCache replication group ID.
	//
	// +required
	ClusterName string `json:"clusterName"`

	// UserName is the ElastiCache user ID with IAM authentication enabled.
	//
	// +required
	UserName string `json:"userName"`

	// ServerlessCacheName is the AWS ElastiCache Serverless cache name used in
	// the IAM token signature. Set this for Serverless ElastiCache deployments;
	// the value is distinct from Address (which remains the Redis endpoint
	// host:port). Leave unset for provisioned ElastiCache clusters.
	//
	// +optional
	ServerlessCacheName *string `json:"serverlessCacheName,omitempty"`
}

// RedisConnectionConfig configures connection pool and timeout tuning for Redis.
type RedisConnectionConfig struct {
	// PoolSize is the maximum number of connections in the pool.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	PoolSize *int32 `json:"poolSize,omitempty"`

	// MinIdleConns is the minimum number of idle connections in the pool.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinIdleConns *int32 `json:"minIdleConns,omitempty"`

	// MaxIdleConns is the maximum number of idle connections in the pool.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxIdleConns *int32 `json:"maxIdleConns,omitempty"`

	// DialTimeout is the timeout for establishing new connections.
	//
	// +optional
	DialTimeout *metav1.Duration `json:"dialTimeout,omitempty"`

	// ReadTimeout is the timeout for reading a single command reply.
	//
	// +optional
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// WriteTimeout is the timeout for writing a single command.
	//
	// +optional
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// PoolTimeout is the time to wait for a connection from the pool
	// when all connections are busy.
	//
	// +optional
	PoolTimeout *metav1.Duration `json:"poolTimeout,omitempty"`

	// ConnMaxIdleTime is the maximum time a connection may be idle before
	// being closed.
	//
	// +optional
	ConnMaxIdleTime *metav1.Duration `json:"connMaxIdleTime,omitempty"`

	// ConnMaxLifetime is the maximum lifetime of a connection before it is
	// closed and recreated, regardless of activity.
	//
	// +optional
	ConnMaxLifetime *metav1.Duration `json:"connMaxLifetime,omitempty"`

	// MaxRetries is the maximum number of retries on failed commands.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxRetries *int32 `json:"maxRetries,omitempty"`

	// MinRetryBackoff is the minimum backoff interval between retries.
	//
	// +optional
	MinRetryBackoff *metav1.Duration `json:"minRetryBackoff,omitempty"`

	// MaxRetryBackoff is the maximum backoff interval between retries.
	//
	// +optional
	MaxRetryBackoff *metav1.Duration `json:"maxRetryBackoff,omitempty"`
}

type WAFConfiguration struct {
	DeploymentConfiguration `json:",inline"`

	// LogLevel is the log level for the WAF extproc server. If not set, defaults to "info".
	// +kubebuilder:validation:Enum=error;warn;info;debug;trace
	// +optional
	LogLevel *WAFLogLevel `json:"logLevel,omitempty"`

	// Admin configures the WAF admin server.
	// +optional
	Admin *WAFAdminConfiguration `json:"admin,omitempty"`
}

type WAFLogLevel string

const (
	WAFLogLevelError WAFLogLevel = "error"
	WAFLogLevelWarn  WAFLogLevel = "warn"
	WAFLogLevelInfo  WAFLogLevel = "info"
	WAFLogLevelDebug WAFLogLevel = "debug"
	WAFLogLevelTrace WAFLogLevel = "trace"
)

type WAFAdminConfiguration struct {
	// Enabled indicates whether the admin server is enabled. If not enabled, then no admin server will be deployed.
	// Defaults to false.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}
