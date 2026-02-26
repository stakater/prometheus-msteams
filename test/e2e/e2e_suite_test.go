/*
Copyright 2026 Stakater.

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
package e2e

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive

	"github.com/stakater/prometheus-msteams/pkg/utility"
)

var (
	update            = flag.Bool("update", false, "update .golden files")
	logger            *utility.Logger
	teamsSrv          *httptest.Server
	testWebhookURL    string
	isIntegrationTest bool
	goldenFile        string
)

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting prometheus-msteams e2e suite\n")
	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
	By("setting up the e2e test suite")

	By("creating logger")
	logger = utility.NewLogger(utility.LogFormatJSON, false)

	By("setting up mock Microsoft Teams server")
	teamsSrv = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			logger.Log("request", string(b))
			w.WriteHeader(200)
			_, _ = w.Write([]byte("1"))
		}),
	)

	By("determining test webhook URL")
	if v := os.Getenv("INTEGRATION_TEST_WEBHOOK_URL"); len(v) > 0 {
		GinkgoWriter.Printf("Running integration test with webhook: %s\n", v)
		testWebhookURL = v
		isIntegrationTest = true
		goldenFile = "TestServer/templated_card_service_test/integration_resp.json"
	} else {
		testWebhookURL = teamsSrv.URL
		isIntegrationTest = false
		goldenFile = "TestServer/templated_card_service_test/resp.json"
	}
})

var _ = AfterSuite(func() {
	By("tearing down the e2e test suite")
	By("cleaning up mock Teams server")
	if teamsSrv != nil {
		teamsSrv.Close()
	}
})
