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
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	upstreamkgateway "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	enterprisekgatewayv1alpha1 "github.com/solo-io/kgateway-client/api/v1alpha1/enterprisekgateway"
	clientset "github.com/solo-io/kgateway-client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	defaultNamespace = "default"
	resourceName     = "demo-enterprisekgateway-traffic-policy"
)

func main() {
	var kubeconfig *string
	var namespace string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", defaultNamespace, "namespace where the EnterpriseKgatewayTrafficPolicy will be managed")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	kgatewayClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	trafficPoliciesClient := kgatewayClient.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(namespace)

	trafficPolicy := newDemoEnterpriseKgatewayTrafficPolicy(namespace)

	fmt.Println("Creating EnterpriseKgatewayTrafficPolicy...")
	result, err := trafficPoliciesClient.Create(context.TODO(), trafficPolicy, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created EnterpriseKgatewayTrafficPolicy %q.\n", result.GetName())

	prompt()
	fmt.Println("Updating EnterpriseKgatewayTrafficPolicy...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, getErr := trafficPoliciesClient.Get(context.TODO(), resourceName, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get latest version of EnterpriseKgatewayTrafficPolicy: %w", getErr)
		}

		if latest.Labels == nil {
			latest.Labels = map[string]string{}
		}
		latest.Labels["examples.solo.io/updated"] = "true"

		_, updateErr := trafficPoliciesClient.Update(context.TODO(), latest, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %w", retryErr))
	}
	fmt.Println("Updated EnterpriseKgatewayTrafficPolicy...")

	prompt()
	fmt.Printf("Listing EnterpriseKgatewayTrafficPolicies in namespace %q:\n", namespace)
	list, err := trafficPoliciesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, policy := range list.Items {
		fmt.Printf(" * %s (entExtAuth.disable=%t)\n", policy.Name, isEntExtAuthDisabled(&policy))
	}

	prompt()
	fmt.Println("Deleting EnterpriseKgatewayTrafficPolicy...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := trafficPoliciesClient.Delete(context.TODO(), resourceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted EnterpriseKgatewayTrafficPolicy.")
}

func newDemoEnterpriseKgatewayTrafficPolicy(namespace string) *enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy {
	return &enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: enterprisekgatewayv1alpha1.SchemeGroupVersion.String(),
			Kind:       "EnterpriseKgatewayTrafficPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
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

func isEntExtAuthDisabled(policy *enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy) bool {
	return policy.Spec.EntExtAuth != nil && policy.Spec.EntExtAuth.Disable != nil
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
