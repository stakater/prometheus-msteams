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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
	"github.com/prometheus/alertmanager/template"

	"github.com/stakater/prometheus-msteams/pkg/card"
	"github.com/stakater/prometheus-msteams/pkg/service"
	"github.com/stakater/prometheus-msteams/pkg/testutils"
	"github.com/stakater/prometheus-msteams/pkg/transport"
)

type alert struct {
	requestPath   string
	promAlertFile string
}

// compareToGoldenFileGinkgo is a Ginkgo-compatible version of testutils.CompareToGoldenFile
// It uses Gomega assertions instead of testing.T
func compareToGoldenFileGinkgo(v any, file string, update bool) {
	gotBytes, err := json.MarshalIndent(v, "", "  ")
	Expect(err).NotTo(HaveOccurred(), "failed to marshal value to JSON")

	gp := filepath.Join("testdata", file)
	dir := filepath.Dir(gp)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0750)
	}
	if _, err := os.Stat(gp); os.IsNotExist(err) {
		_ = os.WriteFile(gp, []byte{}, 0600)
	}

	if update {
		GinkgoWriter.Println("updating golden file:", gp)
		err := os.WriteFile(gp, gotBytes, 0600)
		Expect(err).NotTo(HaveOccurred(), "failed to update golden file")
	}

	filePath := filepath.Clean(gp)
	want, err := os.ReadFile(filePath)
	Expect(err).NotTo(HaveOccurred(), "failed reading the golden file")

	Expect(gotBytes).To(Equal(want))

	/*
		if string(want) != string(gotBytes) {
			a, err := jd.ReadJsonString(string(gotBytes))
			Expect(err).NotTo(HaveOccurred(), "failed to parse got JSON")
			b, err := jd.ReadJsonString(string(want))
			Expect(err).NotTo(HaveOccurred(), "failed to parse want JSON")

			result := fmt.Sprintf(
				"\\ngot:\\n%s\\nwant:\\n%s\\ndiff:\\n%s",
				string(gotBytes),
				string(want),
				a.Diff(b).Render(),
			)
			Fail(result)
		}
	*/
}

var _ = Describe("Server E2E Tests", func() {

	Context("Templated Card Service", func() {
		var (
			templatePath = "../../default-message-workflow-card.tmpl"
			tmpl         *template.Template
			cardCreator  card.Converter

			routes []transport.Route
			alerts []alert
		)

		BeforeEach(func() {
			var err error
			By("parsing the template file")
			tmpl, err = card.ParseTemplateFile(templatePath)
			Expect(err).NotTo(HaveOccurred(), "should successfully parse template file")

			By("creating card converter")
			cardCreator = card.NewTemplatedCardCreator(tmpl, false, logger)

			By("setting up routes with Workflow webhook")
			routes = []transport.Route{
				{
					RequestPath: "/alertmanager",
					Service: service.NewLoggingService(
						logger,
						service.NewSimpleService(
							cardCreator, http.DefaultClient, testWebhookURL, service.Workflow,
						),
					),
				},
			}

			By("setting up test alerts")
			alerts = []alert{
				{
					requestPath:   "/alertmanager",
					promAlertFile: "../data/prom_post_request.json",
				},
				{
					requestPath:   "/alertmanager",
					promAlertFile: "../data/prom_post_request_linebreak.json",
				},
			}
		})

		It("should successfully process Prometheus alerts and send to Teams", func() {
			By("creating the test server")
			srv := transport.NewServer(logger.GetLogger(), routes, []transport.DynamicRoute{})
			testSrv := httptest.NewServer(srv)
			defer testSrv.Close()

			By("processing each alert")
			for _, a := range alerts {
				By(fmt.Sprintf("loading webhook message from %s", a.promAlertFile))
				wm, err := testutils.ParseWebhookJSONFromFile(a.promAlertFile)
				Expect(err).NotTo(HaveOccurred(), "should successfully parse webhook JSON file")

				By("marshaling webhook message")
				b, err := json.Marshal(wm)
				Expect(err).NotTo(HaveOccurred(), "should successfully marshal webhook message")

				By(fmt.Sprintf("creating POST request to %s", a.requestPath))
				req, err := http.NewRequest(
					"POST",
					fmt.Sprintf("%s%s", testSrv.URL, a.requestPath),
					bytes.NewBuffer(b),
				)
				Expect(err).NotTo(HaveOccurred(), "should successfully create HTTP request")

				By("sending the request")
				resp, err := http.DefaultClient.Do(req)
				Expect(err).NotTo(HaveOccurred(), "should successfully send HTTP request")
				defer resp.Body.Close()

				By("validating response status code")
				Expect(resp.StatusCode).To(Equal(200), "should return 200 OK")

				By("decoding response body")
				var prs []service.PostResponse
				err = json.NewDecoder(resp.Body).Decode(&prs)
				Expect(err).NotTo(HaveOccurred(), "should successfully decode response")

				if !isIntegrationTest {
					By("normalizing webhook URLs for comparison")
					for i := range prs {
						Expect(prs[i].WebhookURL).NotTo(BeEmpty(), "webhook URL should not be empty")
						prs[i].WebhookURL = "" // Clear dynamic port
					}
				}

				By("comparing response to golden file")
				compareToGoldenFileGinkgo(prs, goldenFile, *update)
			}
		})
	})
})
