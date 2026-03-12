package enterprisekgateway

// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +groupName=enterprisekgateway.solo.io
// +versionName=v1alpha1

// Portal plugin required permissions for watching ApiProduct resources
// +kubebuilder:rbac:groups=portal.solo.io,resources=apiproducts,verbs=get;list;watch
