package v1alpha1

// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +groupName=portal.solo.io

// Controller resources
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// Leader election lease (namespace-scoped) and events for leader election recording
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create

// Portal controller and deployer required permissions
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;list;watch;create;patch;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;patch;update;delete;deletecollection
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;patch;update;delete;deletecollection

// Portal translator required permissions for watching Gateway API resources
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=grpcroutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=listenersets,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.x-k8s.io,resources=xlistenersets,verbs=get;list;watch

// ApiProduct KRT pipeline watches ListenerSet and EnterpriseListenerSet for server config resolution
// +kubebuilder:rbac:groups=enterprise.solo.io,resources=enterpriselistenersets,verbs=get;list;watch

// Portal API source controller required permissions
// +kubebuilder:rbac:groups=gateway.kgateway.dev,resources=backends,verbs=get;list;watch

// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=referencegrants,verbs=get;list;watch

// CommonCollections watches core resources for service discovery and routing
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.kgateway.dev,resources=gatewayextensions,verbs=get;list;watch

// InitPlugins (EKTP plugin) watches policy-related resources for route auth resolution
// +kubebuilder:rbac:groups=extauth.solo.io,resources=authconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=ratelimit.solo.io,resources=ratelimitconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=waf.solo.io,resources=wafpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=tcproutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=tlsroutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch
