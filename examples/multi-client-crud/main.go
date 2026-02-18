/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	upstreamkgateway "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	upstreamclientset "github.com/kgateway-dev/kgateway/v2/pkg/client/clientset/versioned"
	upstreamtypedkgateway "github.com/kgateway-dev/kgateway/v2/pkg/client/clientset/versioned/typed/v1alpha1/kgateway"
	enterprisekgatewayv1alpha1 "github.com/solo-io/kgateway-client/v2/api/v1alpha1/enterprisekgateway"
	enterprisekgatewayclientset "github.com/solo-io/kgateway-client/v2/clientset/versioned"
	enterprisetypedkgateway "github.com/solo-io/kgateway-client/v2/clientset/versioned/typed/v1alpha1/enterprisekgateway"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	gatewaytypedv1 "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/typed/apis/v1"
)

const (
	defaultNamespace = "default"

	gatewayClassName = "example-gatewayclass"

	gatewayName                    = "demo-gateway"
	httpRouteName                  = "demo-http-route"
	kgatewayTrafficPolicyName      = "demo-kgateway-traffic-policy"
	enterpriseTrafficPolicyName    = "demo-enterprisekgateway-traffic-policy"
	updatedLabelKey                = "examples.solo.io/updated"
	updatedLabelValue              = "true"
	updatedEnterpriseTargetRefName = "demo-gateway-updated"
)

func main() {
	var kubeconfig *string
	var namespace string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to kubeconfig")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig")
	}
	flag.StringVar(&namespace, "namespace", defaultNamespace, "namespace to manage example resources in")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	gatewayClient, err := gatewayclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	upstreamClient, err := upstreamclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	enterpriseClient, err := enterprisekgatewayclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	gateways := gatewayClient.GatewayV1().Gateways(namespace)
	httpRoutes := gatewayClient.GatewayV1().HTTPRoutes(namespace)
	upstreamTrafficPolicies := upstreamClient.GatewayKgateway().TrafficPolicies(namespace)
	enterpriseTrafficPolicies := enterpriseClient.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(namespace)

	fmt.Println("Creating resources across all 3 clientsets...")
	createGateway(ctx, gateways, namespace)
	createHTTPRoute(ctx, httpRoutes, namespace)
	createUpstreamTrafficPolicy(ctx, upstreamTrafficPolicies, namespace)
	createEnterpriseTrafficPolicy(ctx, enterpriseTrafficPolicies, namespace)

	fmt.Println("Updating each resource...")
	updateGateway(ctx, gateways)
	updateHTTPRoute(ctx, httpRoutes)
	updateUpstreamTrafficPolicy(ctx, upstreamTrafficPolicies)
	updateEnterpriseTrafficPolicy(ctx, enterpriseTrafficPolicies)

	fmt.Println("Listing resources...")
	listResources(ctx, gateways, httpRoutes, upstreamTrafficPolicies, enterpriseTrafficPolicies)

	fmt.Println("Deleting resources...")
	deleteResource("HTTPRoute", httpRouteName, func() error {
		return httpRoutes.Delete(ctx, httpRouteName, metav1.DeleteOptions{})
	})
	deleteResource("TrafficPolicy", kgatewayTrafficPolicyName, func() error {
		return upstreamTrafficPolicies.Delete(ctx, kgatewayTrafficPolicyName, metav1.DeleteOptions{})
	})
	deleteResource("EnterpriseKgatewayTrafficPolicy", enterpriseTrafficPolicyName, func() error {
		return enterpriseTrafficPolicies.Delete(ctx, enterpriseTrafficPolicyName, metav1.DeleteOptions{})
	})
	deleteResource("Gateway", gatewayName, func() error {
		return gateways.Delete(ctx, gatewayName, metav1.DeleteOptions{})
	})

	fmt.Println("Done.")
}

func createGateway(ctx context.Context, gateways gatewaytypedv1.GatewayInterface, namespace string) {
	_, err := gateways.Create(ctx, newGateway(namespace), metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Printf("Gateway %q already exists\n", gatewayName)
			return
		}
		panic(err)
	}
	fmt.Printf("Created Gateway %q\n", gatewayName)
}

func createHTTPRoute(ctx context.Context, routes gatewaytypedv1.HTTPRouteInterface, namespace string) {
	_, err := routes.Create(ctx, newHTTPRoute(namespace), metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Printf("HTTPRoute %q already exists\n", httpRouteName)
			return
		}
		panic(err)
	}
	fmt.Printf("Created HTTPRoute %q\n", httpRouteName)
}

func createUpstreamTrafficPolicy(ctx context.Context, policies upstreamtypedkgateway.TrafficPolicyInterface, namespace string) {
	_, err := policies.Create(ctx, newUpstreamTrafficPolicy(namespace), metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Printf("TrafficPolicy %q already exists\n", kgatewayTrafficPolicyName)
			return
		}
		panic(err)
	}
	fmt.Printf("Created TrafficPolicy %q\n", kgatewayTrafficPolicyName)
}

func createEnterpriseTrafficPolicy(
	ctx context.Context,
	policies enterprisetypedkgateway.EnterpriseKgatewayTrafficPolicyInterface,
	namespace string,
) {
	_, err := policies.Create(ctx, newEnterpriseTrafficPolicy(namespace), metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Printf("EnterpriseKgatewayTrafficPolicy %q already exists\n", enterpriseTrafficPolicyName)
			return
		}
		panic(err)
	}
	fmt.Printf("Created EnterpriseKgatewayTrafficPolicy %q\n", enterpriseTrafficPolicyName)
}

func updateGateway(ctx context.Context, gateways gatewaytypedv1.GatewayInterface) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := gateways.Get(ctx, gatewayName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if latest.Labels == nil {
			latest.Labels = map[string]string{}
		}
		latest.Labels[updatedLabelKey] = updatedLabelValue
		_, err = gateways.Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("failed to update Gateway %q: %w", gatewayName, retryErr))
	}
	fmt.Printf("Updated Gateway %q\n", gatewayName)
}

func updateHTTPRoute(ctx context.Context, routes gatewaytypedv1.HTTPRouteInterface) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := routes.Get(ctx, httpRouteName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if latest.Labels == nil {
			latest.Labels = map[string]string{}
		}
		latest.Labels[updatedLabelKey] = updatedLabelValue
		_, err = routes.Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("failed to update HTTPRoute %q: %w", httpRouteName, retryErr))
	}
	fmt.Printf("Updated HTTPRoute %q\n", httpRouteName)
}

func updateUpstreamTrafficPolicy(ctx context.Context, policies upstreamtypedkgateway.TrafficPolicyInterface) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := policies.Get(ctx, kgatewayTrafficPolicyName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if latest.Labels == nil {
			latest.Labels = map[string]string{}
		}
		latest.Labels[updatedLabelKey] = updatedLabelValue
		_, err = policies.Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("failed to update TrafficPolicy %q: %w", kgatewayTrafficPolicyName, retryErr))
	}
	fmt.Printf("Updated TrafficPolicy %q\n", kgatewayTrafficPolicyName)
}

func updateEnterpriseTrafficPolicy(
	ctx context.Context,
	policies enterprisetypedkgateway.EnterpriseKgatewayTrafficPolicyInterface,
) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := policies.Get(ctx, enterpriseTrafficPolicyName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if latest.Labels == nil {
			latest.Labels = map[string]string{}
		}
		latest.Labels[updatedLabelKey] = updatedLabelValue
		if len(latest.Spec.TargetRefs) > 0 {
			latest.Spec.TargetRefs[0].Name = gwv1.ObjectName(updatedEnterpriseTargetRefName)
		}
		_, err = policies.Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("failed to update EnterpriseKgatewayTrafficPolicy %q: %w", enterpriseTrafficPolicyName, retryErr))
	}
	fmt.Printf("Updated EnterpriseKgatewayTrafficPolicy %q\n", enterpriseTrafficPolicyName)
}

func listResources(
	ctx context.Context,
	gateways gatewaytypedv1.GatewayInterface,
	httpRoutes gatewaytypedv1.HTTPRouteInterface,
	upstreamTrafficPolicies upstreamtypedkgateway.TrafficPolicyInterface,
	enterpriseTrafficPolicies enterprisetypedkgateway.EnterpriseKgatewayTrafficPolicyInterface,
) {
	gatewayList, err := gateways.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Gateway count in namespace: %d\n", len(gatewayList.Items))

	httpRouteList, err := httpRoutes.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("HTTPRoute count in namespace: %d\n", len(httpRouteList.Items))

	upstreamPolicyList, err := upstreamTrafficPolicies.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("TrafficPolicy count in namespace: %d\n", len(upstreamPolicyList.Items))

	enterprisePolicyList, err := enterpriseTrafficPolicies.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("EnterpriseKgatewayTrafficPolicy count in namespace: %d\n", len(enterprisePolicyList.Items))
}

func deleteResource(kind string, name string, deleteFunc func() error) {
	if err := deleteFunc(); err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Printf("%s %q already deleted\n", kind, name)
			return
		}
		panic(err)
	}
	fmt.Printf("Deleted %s %q\n", kind, name)
}

func newGateway(namespace string) *gwv1.Gateway {
	return &gwv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gwv1.GroupVersion.String(),
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatewayName,
			Namespace: namespace,
		},
		Spec: gwv1.GatewaySpec{
			GatewayClassName: gwv1.ObjectName(gatewayClassName),
			Listeners: []gwv1.Listener{
				{
					Name:     gwv1.SectionName("http"),
					Protocol: gwv1.HTTPProtocolType,
					Port:     gwv1.PortNumber(80),
				},
			},
		},
	}
}

func newHTTPRoute(namespace string) *gwv1.HTTPRoute {
	pathPrefix := gwv1.PathMatchPathPrefix
	pathValue := "/"
	return &gwv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gwv1.GroupVersion.String(),
			Kind:       "HTTPRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      httpRouteName,
			Namespace: namespace,
		},
		Spec: gwv1.HTTPRouteSpec{
			CommonRouteSpec: gwv1.CommonRouteSpec{
				ParentRefs: []gwv1.ParentReference{
					{
						Name: gwv1.ObjectName(gatewayName),
					},
				},
			},
			Rules: []gwv1.HTTPRouteRule{
				{
					Matches: []gwv1.HTTPRouteMatch{
						{
							Path: &gwv1.HTTPPathMatch{
								Type:  &pathPrefix,
								Value: &pathValue,
							},
						},
					},
				},
			},
		},
	}
}

func newUpstreamTrafficPolicy(namespace string) *upstreamkgateway.TrafficPolicy {
	return &upstreamkgateway.TrafficPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: upstreamkgateway.SchemeGroupVersion.String(),
			Kind:       "TrafficPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kgatewayTrafficPolicyName,
			Namespace: namespace,
		},
		Spec: upstreamkgateway.TrafficPolicySpec{
			TargetRefs: []upstreamshared.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: upstreamshared.LocalPolicyTargetReference{
						Group: gwv1.Group("gateway.networking.k8s.io"),
						Kind:  gwv1.Kind("Gateway"),
						Name:  gwv1.ObjectName(gatewayName),
					},
				},
			},
			ExtAuth: &upstreamkgateway.ExtAuthPolicy{
				Disable: &upstreamshared.PolicyDisable{},
			},
		},
	}
}

func newEnterpriseTrafficPolicy(namespace string) *enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy {
	return &enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: enterprisekgatewayv1alpha1.SchemeGroupVersion.String(),
			Kind:       "EnterpriseKgatewayTrafficPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      enterpriseTrafficPolicyName,
			Namespace: namespace,
		},
		Spec: enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicySpec{
			TrafficPolicySpec: upstreamkgateway.TrafficPolicySpec{
				TargetRefs: []upstreamshared.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: upstreamshared.LocalPolicyTargetReference{
							Group: gwv1.Group("gateway.networking.k8s.io"),
							Kind:  gwv1.Kind("Gateway"),
							Name:  gwv1.ObjectName(gatewayName),
						},
					},
				},
			},
			EntExtAuth: &enterprisekgatewayv1alpha1.EntExtAuth{
				Disable: &upstreamshared.PolicyDisable{},
			},
		},
	}
}
