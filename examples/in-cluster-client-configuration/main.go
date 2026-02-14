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

	clientset "github.com/solo-io/kgateway-client/clientset/versioned"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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

	for {
		trafficPolicies, err := kgatewayClient.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d EnterpriseKgatewayTrafficPolicies in namespace %q\n", len(trafficPolicies.Items), namespace)

		_, err = kgatewayClient.EnterprisekgatewayEnterprisekgateway().EnterpriseKgatewayTrafficPolicies(namespace).Get(context.TODO(), exampleResourceKey, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			fmt.Printf("EnterpriseKgatewayTrafficPolicy %q in namespace %q not found\n", exampleResourceKey, namespace)
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
