package v1alpha1

import (
	upstream "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:rbac:groups=portal.solo.io,resources=portalparameters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=portal.solo.io,resources=portalparameters/status,verbs=get;update;patch

// PortalParameters defines operational configuration for a Portal deployment,
// including data store settings and postgres connection details.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={portal},path=portalparameters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app=gloo,app.kubernetes.io/name=portalparameters}
type PortalParameters struct {
	metav1.TypeMeta `json:",inline"`
	//nolint:kubeapilinter // consistent with Portal type
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired operational configuration for the Portal
	// +required
	Spec PortalParametersSpec `json:"spec"`

	// Status defines the observed state of the PortalParameters
	// +optional
	//nolint:kubeapilinter
	Status PortalParametersStatus `json:"status,omitempty"`
}

// PortalParametersSpec defines the desired operational configuration for a Portal deployment.
//
// +kubebuilder:validation:XValidation:rule="!has(self.authServer) || (has(self.store) && has(self.store.postgres))",message="authServer requires spec.store.postgres; authServer is only supported with a postgres store"
type PortalParametersSpec struct {
	// store configures the data store for the portal backend.
	// Exactly one store type must be specified. If omitted, defaults to in-memory.
	// +optional
	Store *StoreConfig `json:"store,omitempty"`

	// webServer configures the portal web server deployment.
	// +optional
	WebServer *PortalWebServer `json:"webServer,omitempty"`

	// authServer configures the portal auth server deployment, which serves the
	// data-plane /v1/metadata endpoint that ExtAuth calls to validate API keys and
	// OAuth tokens. The portal auth server is opt-in: when this block is omitted, no
	// dedicated Deployment is rendered and ExtAuth continues to call /v1/metadata on
	// the portal web server (existing behavior). To enable, set this block — its
	// presence is the opt-in signal.
	// +optional
	AuthServer *PortalAuthServer `json:"authServer,omitempty"`

	// idpServerURL is the URL of an Identity Provider SPI server for OAuth client management.
	// When set, OAuth credential create/delete operations are delegated to this external server
	// (e.g., Keycloak via gloo-portal-idp-connect). When omitted, oauth endpoint will return a 500.
	// +optional
	IdpServerURL *string `json:"idpServerURL,omitempty"`

	// adminAuth configures how Portal admin privileges are derived from a caller's JWT claims.
	// When omitted, a user is an admin if the JWT's "groups" claim contains "admin".
	// +optional
	AdminAuth *AdminAuthConfig `json:"adminAuth,omitempty"`
}

const (
	// DefaultAdminGroupClaim is the default JWT claim the portal web server reads group
	// membership from when AdminAuthConfig.GroupClaim is unset.
	DefaultAdminGroupClaim = "groups"

	// DefaultAdminGroup is the default group value that grants Portal admin when
	// AdminAuthConfig.AdminGroups is unset.
	DefaultAdminGroup = "admin"
)

// AdminAuthConfig configures admin-group mapping for the portal web server. Admin status is
// derived from the caller's validated JWT. The web server reads the configured claim and grants
// admin when it contains any of the configured group values.
type AdminAuthConfig struct {
	// groupClaim is the JWT claim to read group membership from. Defaults to "groups".
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	GroupClaim *string `json:"groupClaim,omitempty"`

	// adminGroups are the group values that grant Portal admin. A caller is an admin when the
	// groupClaim contains any of these values, compared as an exact, case-sensitive match.
	// Defaults to ["admin"].
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=253
	// +listType=atomic
	// +optional
	AdminGroups []string `json:"adminGroups,omitempty"`
}

// PortalWebServer configures the portal web server deployment.
type PortalWebServer struct {
	// replicas is the number of portal web server pods.
	// If omitted, defaults to 1. If using an HPA, do not set this field.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// resources configures CPU and memory requests/limits for the portal web server container.
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// container configures the portal web server container.
	// +optional
	Container *PortalWebServerContainer `json:"container,omitempty"`

	// LogLevel sets the log level for the portal web server process. One of:
	// "error", "warn", "info", "debug", "trace". When unset, the binary's
	// default level (info) is used.
	// +kubebuilder:validation:Enum=error;warn;info;debug;trace
	// +optional
	LogLevel *string `json:"logLevel,omitempty"`

	// GatewayParametersOverlays contains overlay fields for portal web server resources.
	// These allow applying strategic merge patches and creating HPA/PDB/VPA resources.
	upstream.GatewayParametersOverlays `json:",inline"`
}

// PortalWebServerContainer configures the portal web server container.
type PortalWebServerContainer struct {
	// image overrides the portal web server container image.
	// Individual fields (registry, repository, tag, pullPolicy) can be set
	// independently; unset fields retain the chart defaults.
	// +optional
	Image *upstream.Image `json:"image,omitempty"`
}

// PortalAuthServer configures the portal auth server deployment. The portal auth
// server serves the data-plane /v1/metadata endpoint for ExtAuth credential
// validation. It runs the same image as the portal web server with --role=portal-auth,
// which skips the migrator, the Portal/ApiDoc informers, and all admin routes.
//
// Running this as a separate Deployment isolates the high-QPS data plane from
// admin CRUD traffic — independent scaling, independent fault domain, independent
// DB connection pool.
type PortalAuthServer struct {
	// replicas is the number of portal auth server pods.
	// If omitted, defaults to 1. If using an HPA, do not set this field.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// resources configures CPU and memory requests/limits for the portal auth server container.
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// container configures the portal auth server container. The image is shared
	// with the portal web server; this field exists to allow per-role image overrides
	// (for example, pinning the auth server to a specific tag during a rolling upgrade).
	// +optional
	Container *PortalWebServerContainer `json:"container,omitempty"`

	// LogLevel sets the log level for the portal auth server process. One of:
	// "error", "warn", "info", "debug", "trace". When unset, the binary's
	// default level (info) is used.
	// +kubebuilder:validation:Enum=error;warn;info;debug;trace
	// +optional
	LogLevel *string `json:"logLevel,omitempty"`

	// postgres configures the postgres connection for the portal auth server.
	// When unset, the portal auth server uses spec.store.postgres for the database config.
	// +optional
	Postgres *PostgresStoreConfig `json:"postgres,omitempty"`

	// GatewayParametersOverlays contains overlay fields for portal auth server resources.
	// These allow applying strategic merge patches and creating HPA/PDB/VPA resources
	// scoped to the portal auth Deployment only.
	upstream.GatewayParametersOverlays `json:",inline"`
}

// StoreConfig configures the data store for the portal backend.
// Exactly one store type must be specified.
//
// +kubebuilder:validation:ExactlyOneOf=memory;postgres
type StoreConfig struct {
	// memory selects the in-memory data store.
	// +optional
	Memory *MemoryStoreConfig `json:"memory,omitempty"`

	// postgres selects a PostgreSQL data store.
	// +optional
	Postgres *PostgresStoreConfig `json:"postgres,omitempty"`
}

// MemoryStoreConfig configures an in-memory data store for the portal backend.
// This store is ephemeral and data is lost on pod restart.
type MemoryStoreConfig struct{}

// PostgresStoreConfig configures a PostgreSQL data store for the portal backend.
type PostgresStoreConfig struct {
	// secretRef references a Secret containing postgres connection credentials.
	// Required keys: host, database, username, password
	// Optional keys: port (default: 5432), sslmode (default: require)
	// +required
	SecretRef LocalSecretReference `json:"secretRef,omitzero"`

	// tls configures TLS for the postgres connection.
	// +optional
	TLS *PostgresTLSConfig `json:"tls,omitempty"`

	// schema is the PostgreSQL schema name for portal tables.
	// Defaults to "portal" if not specified.
	// +kubebuilder:validation:Pattern=`^[a-zA-Z_][a-zA-Z0-9_]*$`
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:XValidation:rule="!self.startsWith('pg_')",message="schema name must not start with 'pg_'"
	// +kubebuilder:validation:XValidation:rule="self != 'information_schema'",message="schema name must not be 'information_schema'"
	// +kubebuilder:validation:XValidation:rule="self != 'public'",message="schema name must not be 'public'"
	// +optional
	Schema *string `json:"schema,omitempty"`
}

// LocalSecretReference identifies a Secret in the same namespace.
type LocalSecretReference struct {
	// name is the name of the Secret.
	// +required
	Name gwv1.ObjectName `json:"name"`
}

// +kubebuilder:validation:Enum=disable;require;verify-ca;verify-full
type PostgresSSLMode string

const (
	PostgresSSLModeDisable    PostgresSSLMode = "disable"
	PostgresSSLModeRequire    PostgresSSLMode = "require"
	PostgresSSLModeVerifyCA   PostgresSSLMode = "verify-ca"
	PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)

// PostgresTLSConfig configures TLS settings for the postgres connection.
// +kubebuilder:validation:XValidation:rule="!has(self.mode) || self.mode == 'disable' || self.mode == 'require' || has(self.caCertSecretRef)",message="caCertSecretRef is required when mode is verify-ca or verify-full"
// +kubebuilder:validation:XValidation:rule="!has(self.clientCertSecretRef) || !has(self.mode) || self.mode != 'disable'",message="clientCertSecretRef cannot be set when mode is disable"
type PostgresTLSConfig struct {
	// mode is the SSL mode for postgres: disable, require, verify-ca, verify-full.
	// Overrides the sslmode key in the credentials secret if both are set.
	// +optional
	Mode *PostgresSSLMode `json:"mode,omitempty"`

	// caCertSecretRef references a Secret containing a CA certificate (key: "ca.crt")
	// for verifying the postgres server. Required for verify-ca and verify-full modes.
	// +optional
	CACertSecretRef *LocalSecretReference `json:"caCertSecretRef,omitempty"`

	// clientCertSecretRef references a Secret containing a client certificate and key
	// (keys: "tls.crt", "tls.key") for mutual TLS authentication with postgres.
	// +optional
	ClientCertSecretRef *LocalSecretReference `json:"clientCertSecretRef,omitempty"`
}

// PortalParametersStatus defines the observed state of PortalParameters.
type PortalParametersStatus struct{}

// PortalParametersList contains a list of PortalParameters resources
// +kubebuilder:object:root=true
type PortalParametersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortalParameters `json:"items"`
}
