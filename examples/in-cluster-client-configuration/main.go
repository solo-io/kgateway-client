/*
Copyright 2025.

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
	"fmt"
	"os"
	"time"

	upstreamkgateway "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	enterprisekgatewayv1alpha1 "github.com/solo-io/kgateway-client/api/v1alpha1/enterprisekgateway"
	clientset "github.com/solo-io/kgateway-client/clientset/versioned"
	typedenterprisekgatewayv1alpha1 "github.com/solo-io/kgateway-client/clientset/versioned/typed/v1alpha1/enterprisekgateway"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	defaultNamespace   = "default"
	exampleResourceKey = "example-enterprisekgateway-traffic-policy"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	kgatewayClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = defaultNamespace
	}

	trafficPoliciesClient := kgatewayClient.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(namespace)
	ctx := context.TODO()

	if err := ensureExampleEnterpriseKgatewayTrafficPolicy(ctx, trafficPoliciesClient, namespace); err != nil {
		panic(err.Error())
	}

	for {
		trafficPolicies, err := trafficPoliciesClient.List(ctx, metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d EnterpriseKgatewayTrafficPolicies in namespace %q\n", len(trafficPolicies.Items), namespace)

		_, err = trafficPoliciesClient.Get(ctx, exampleResourceKey, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			fmt.Printf("EnterpriseKgatewayTrafficPolicy %q in namespace %q not found; creating it\n", exampleResourceKey, namespace)
			if ensureErr := ensureExampleEnterpriseKgatewayTrafficPolicy(ctx, trafficPoliciesClient, namespace); ensureErr != nil {
				panic(ensureErr.Error())
			}
		} else if statusError, isStatus := err.(*k8serrors.StatusError); isStatus {
			fmt.Printf("Error getting EnterpriseKgatewayTrafficPolicy %q in namespace %q: %v\n", exampleResourceKey, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found EnterpriseKgatewayTrafficPolicy %q in namespace %q\n", exampleResourceKey, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}

func ensureExampleEnterpriseKgatewayTrafficPolicy(
	ctx context.Context,
	trafficPoliciesClient typedenterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicyInterface,
	namespace string,
) error {
	_, err := trafficPoliciesClient.Get(ctx, exampleResourceKey, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !k8serrors.IsNotFound(err) {
		return err
	}

	_, err = trafficPoliciesClient.Create(ctx, newExampleEnterpriseKgatewayTrafficPolicy(namespace), metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	fmt.Printf("Created EnterpriseKgatewayTrafficPolicy %q in namespace %q\n", exampleResourceKey, namespace)
	return nil
}

func newExampleEnterpriseKgatewayTrafficPolicy(namespace string) *enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy {
	return &enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: enterprisekgatewayv1alpha1.SchemeGroupVersion.String(),
			Kind:       "EnterpriseKgatewayTrafficPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      exampleResourceKey,
			Namespace: namespace,
		},
		Spec: enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicySpec{
			TrafficPolicySpec: upstreamkgateway.TrafficPolicySpec{
				TargetRefs: []upstreamshared.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: upstreamshared.LocalPolicyTargetReference{
							Group: gwv1.Group("gateway.networking.k8s.io"),
							Kind:  gwv1.Kind("Gateway"),
							Name:  gwv1.ObjectName("example-gateway"),
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
