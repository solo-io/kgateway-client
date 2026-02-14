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

package fakeclient

import (
	"context"
	"testing"

	upstreamkgateway "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	enterprisekgatewayv1alpha1 "github.com/solo-io/kgateway-client/api/v1alpha1/enterprisekgateway"
	clientsetfake "github.com/solo-io/kgateway-client/clientset/versioned/fake"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const testNamespace = "test-ns"

func TestFakeClientsetEnterpriseKgatewayTrafficPolicyCRUD(t *testing.T) {
	ctx := context.Background()
	seed := newEnterpriseKgatewayTrafficPolicy("existing")

	client := clientsetfake.NewSimpleClientset(seed)
	trafficPoliciesClient := client.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(testNamespace)

	list, err := trafficPoliciesClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 seeded EnterpriseKgatewayTrafficPolicy, got %d", len(list.Items))
	}

	created, err := trafficPoliciesClient.Create(ctx, newEnterpriseKgatewayTrafficPolicy("created"), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Name != "created" {
		t.Fatalf("unexpected name: %s", created.Name)
	}

	got, err := trafficPoliciesClient.Get(ctx, "created", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if got.Labels == nil {
		got.Labels = map[string]string{}
	}
	got.Labels["examples.solo.io/updated"] = "true"
	if _, err := trafficPoliciesClient.Update(ctx, got, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if err := trafficPoliciesClient.Delete(ctx, "created", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = trafficPoliciesClient.Get(ctx, "created", metav1.GetOptions{})
	if !k8serrors.IsNotFound(err) {
		t.Fatalf("expected NotFound after delete, got: %v", err)
	}
}

func newEnterpriseKgatewayTrafficPolicy(name string) *enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy {
	return &enterprisekgatewayv1alpha1.EnterpriseKgatewayTrafficPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: enterprisekgatewayv1alpha1.SchemeGroupVersion.String(),
			Kind:       "EnterpriseKgatewayTrafficPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
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
