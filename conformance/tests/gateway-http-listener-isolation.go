/* Copyright 2022 The Kubernetes Authors.

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

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayHTTPListenerIsolation)
}

var GatewayHTTPListenerIsolation = suite.ConformanceTest{
	ShortName:   "GatewayHTTPListenerIsolation",
	Description: "Listener isolation for HTTP Listeners with multiple Listeners and HTTPRoutes",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportHTTPRoute,
	},
	Manifests: []string{"tests/gateway-http-listener-isolation.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		// This test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwNN := types.NamespacedName{Name: "gateway-http-listener-isolation", Namespace: ns}

		routes := []types.NamespacedName{
			{Namespace: ns, Name: "attaches-to-empty-hostname"},
			{Namespace: ns, Name: "attaches-to-wildcard-example-com"},
			{Namespace: ns, Name: "attaches-to-wildcard-foo-example-com"},
		}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routes...)
		for _, routeNN := range routes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}

		testCases := []http.ExpectedResponse{
			{
				Request:   http.Request{Host: "bar.com", Path: "/empty-hostname"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "bar.com", Path: "/wildcard-example-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "bar.com", Path: "/foo-wildcard-example-com"},
				Response: http.Response{StatusCode: 404},
			},
		}

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
