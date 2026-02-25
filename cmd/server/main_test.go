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
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"

	ocprometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stakater/prometheus-msteams/pkg/service"
	"github.com/stakater/prometheus-msteams/pkg/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
func TestRunApplication(t *testing.T) {
	// Not a unit test - this is an end-to-end test that runs the application
	// with different flags to ensure it starts without errors.
	// For actual e2e tests, see test/e2e/e2e_test.go"
	t.SkipNow()
	tests := []struct {
		name  string
		flags []string
	}{
		{
			name:  "run application without errors",
			flags: []string{},
		},
		{
			name: "run application with version flag",
			flags: []string{
				"-version",
			},
		},
		{
			name: "run application with json logs enabled",
			flags: []string{
				"-log-format", "json",
			},
		},
		{
			name: "run application with fmt logs enabled",
			flags: []string{
				"-log-format", "fmt",
			},
		},
		{
			name: "run application with debug logs enabled",
			flags: []string{
				"-debug",
			},
		},
		{
			name: "run application with custom HTTP address",
			flags: []string{
				"-http-addr", ":8080",
			},
		},
		{
			name: "run application with auto escape underscores disabled",
			flags: []string{
				"-auto-escape-underscores=false",
			},
		},
		{
			name: "run application with jaeger trace enabled",
			flags: []string{
				"-jaeger-trace=true",
			},
		},
		{
			name: "run application with jaeger trace enabled and custom agent",
			flags: []string{
				"-jaeger-trace=true",
				"-jaeger-agent=custom-agent:6831",
			},
		},
		{
			name: "run application with default request uri",
			flags: []string{
				"-teams-request-uri", "/alertmanager",
			},
		},
		{
			name: "run application with default webhook connector",
			flags: []string{
				"-teams-incoming-webhook-url", "default",
			},
		},
		{
			name: "run application with config file",
			flags: []string{
				"-config-file", "./testdata/test-config.yaml",
			},
		},
		{
			name: "run application with custom idle connection timeout",
			flags: []string{
				"-idle-conn-timeout=1s",
			},
		},
		{
			name: "run application with custom TLS handshake timeout",
			flags: []string{
				"-tls-handshake-timeout=1s",
			},
		},
		{
			name: "run application with custom max idle connections",
			flags: []string{
				"-max-idle-conns=1",
			},
		},
		{
			name: "run application with insecure skip verify enabled",
			flags: []string{
				"-insecure-skip-verify=true",
			},
		},
		{
			name: "run application with max retry count of 1",
			flags: []string{
				"-max-retry-count=1",
			},
		},
		{
			name: "run application with max retry count of 5",
			flags: []string{
				"-max-retry-count=5",
			},
		},
		{
			name: "run application with strict validation of webhook url enabled",
			flags: []string{
				"-validate-webhook-url=true",
			},
		},
	}
	defaultArgs := []string{
		"prometheus-msteams",
		"-template-file", "../../default-message-workflow-card.tmpl",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = defaultArgs
			os.Args = append(os.Args, tt.flags...)
			err := Run()
			assert.NoError(t, err)
		})
	}
}
*/

func TestParseFlagsDefaults(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args with just program name
	os.Args = []string{"prometheus-msteams"}

	cfg, err := parseFlags()
	require.NoError(t, err)

	assert.False(t, cfg.Version)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.False(t, cfg.DebugLogs)
	assert.False(t, cfg.JaegerTrace)
	assert.Equal(t, ":2000", cfg.HTTPAddr)
	assert.Equal(t, service.Workflow, cfg.WebhookType)
	assert.True(t, cfg.EscapeUnderscores)
	assert.Equal(t, 3, cfg.RetryMax)
	assert.False(t, cfg.ValidateWebhookURL)
}

func TestParseFlagsWorkflowWebhookUsesCorrectTemplate(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"prometheus-msteams",
	}

	cfg, err := parseFlags()
	require.NoError(t, err)

	assert.Equal(t, service.Workflow, cfg.WebhookType)
	assert.Equal(t, "./default-message-workflow-card.tmpl", cfg.TemplateFile)
}

func TestSetupLoggerJSONFormat(t *testing.T) {
	cfg := Config{
		LogFormat: "json",
		DebugLogs: false,
	}

	logger := setupLogger(cfg)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.GetLogger())
}

func TestSetupLoggerFmtFormat(t *testing.T) {
	cfg := Config{
		LogFormat: "fmt",
		DebugLogs: true,
	}

	logger := setupLogger(cfg)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.GetLogger())
}

func TestSetupHTTPClient(t *testing.T) {
	cfg := Config{
		DebugLogs:                     false,
		RetryMax:                      3,
		HTTPClientMaxIdleConn:         100,
		HTTPClientIdleConnTimeout:     90 * time.Second,
		HTTPClientTLSHandshakeTimeout: 30 * time.Second,
		InsecureSkipVerify:            false,
	}

	client := setupHTTPClient(cfg)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}

func TestCheckDuplicateRequestPathNoDuplicates(t *testing.T) {
	routes := []transport.Route{
		{RequestPath: "/path1"},
		{RequestPath: "/path2"},
		{RequestPath: "/path3"},
	}

	err := checkDuplicateRequestPath(routes)
	assert.NoError(t, err)
}

func TestCheckDuplicateRequestPathWithDuplicates(t *testing.T) {
	routes := []transport.Route{
		{RequestPath: "/path1"},
		{RequestPath: "/path2"},
		{RequestPath: "/path1"}, // Duplicate
	}

	err := checkDuplicateRequestPath(routes)
	assert.Error(t, err)
	assert.Equal(t, "found duplicate use of request path '/path1'", err.Error())
}

func TestCheckDuplicateRequestPathEmptyRoutes(t *testing.T) {
	routes := []transport.Route{}

	err := checkDuplicateRequestPath(routes)
	assert.NoError(t, err)
}

func TestSetupHTTPClientInsecureSkipVerify(t *testing.T) {
	cfg := Config{
		DebugLogs:                     true,
		RetryMax:                      5,
		HTTPClientMaxIdleConn:         50,
		HTTPClientIdleConnTimeout:     60 * time.Second,
		HTTPClientTLSHandshakeTimeout: 20 * time.Second,
		InsecureSkipVerify:            true,
	}

	client := setupHTTPClient(cfg)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
}

func TestSetupTracerDisabled(t *testing.T) {
	cfg := Config{
		JaegerTrace: false,
	}
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})

	err := setupTracer(cfg, logger)
	assert.NoError(t, err)
}

func TestSetupConverterValidTemplate(t *testing.T) {
	cfg := Config{
		TemplateFile:      "../../default-message-workflow-card.tmpl",
		EscapeUnderscores: true,
	}
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})

	converter, err := setupConverter(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, converter)
}

func TestSetupConverterInvalidTemplate(t *testing.T) {
	cfg := Config{
		TemplateFile:      "./nonexistent-template.tmpl",
		EscapeUnderscores: false,
	}
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})

	converter, err := setupConverter(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, converter)
}

func TestParseTeamsConfigFileInvalidPath(t *testing.T) {
	_, err := parseTeamsConfigFile("./nonexistent-config.yaml")
	assert.Error(t, err)
}

func TestParseTeamsConfigFileValidConfig(t *testing.T) {
	config, err := parseTeamsConfigFile("./testdata/test-config.yaml")
	require.NoError(t, err)

	assert.Len(t, config.Connectors, 2)
	assert.Len(t, config.ConnectorsWithCustomTemplates, 1)

	// Check first connector
	assert.Contains(t, config.Connectors[0], "/alertmanager")

	// Check templated connector
	assert.Equal(t, "/custom", config.ConnectorsWithCustomTemplates[0].RequestPath)
	assert.True(t, config.ConnectorsWithCustomTemplates[0].EscapeUnderscores)
}

func TestParseTeamsConfigFileEmptyFile(t *testing.T) {
	// Create a temporary empty YAML file
	tmpFile, err := os.CreateTemp("", "empty-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("")
	tmpFile.Close()

	config, err := parseTeamsConfigFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Empty(t, config.Connectors)
	assert.Empty(t, config.ConnectorsWithCustomTemplates)
}

func TestSetupTracer(t *testing.T) {
	cfg := Config{
		LogFormat: "json",
		DebugLogs: false,
	}
	logger := setupLogger(cfg)

	err := setupTracer(cfg, logger)
	require.NoError(t, err)
}

func TestSetupPrometheusExporter(t *testing.T) {
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})

	exporter, err := setupPrometheusExporter(logger)
	require.NoError(t, err)
	assert.NotNil(t, exporter)
}

func TestSetupServer(t *testing.T) {
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})

	// Don't call setupPrometheusExporter here to avoid duplicate view registration
	// Just create a mock exporter for testing purposes
	pe, err := ocprometheus.NewExporter(
		ocprometheus.Options{
			Registry: prometheus.NewRegistry(), // Use new registry to avoid conflicts
		},
	)
	require.NoError(t, err)

	routes := []transport.Route{
		{
			RequestPath: "/test",
		},
	}
	dRoutes := []transport.DynamicRoute{}
	tc := PromTeamsConfig{}

	handler := setupServer(logger, routes, dRoutes, tc, pe)
	assert.NotNil(t, handler)
}

func TestSetupRoutesWithConnectors(t *testing.T) {
	cfg := Config{
		TemplateFile:       "../../default-message-workflow-card.tmpl",
		EscapeUnderscores:  true,
		RequestURI:         "/alertmanager",
		TeamsWebhookURL:    "https://test.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/testid/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=testtoken",
		ValidateWebhookURL: false,
		WebhookType:        service.Workflow,
	}

	tc := PromTeamsConfig{}
	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})
	converter, err := setupConverter(cfg, logger)
	require.NoError(t, err)

	httpClient := setupHTTPClient(cfg)

	routes, dRoutes, err := setupRoutes(cfg, tc, logger, converter, httpClient)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(routes), 1)
	assert.Len(t, dRoutes, 1) // Dynamic route should always be present
}

func TestSetupRoutesWithConfigFile(t *testing.T) {
	cfg := Config{
		TemplateFile:       "../../default-message-workflow-card.tmpl",
		EscapeUnderscores:  true,
		ValidateWebhookURL: false,
		WebhookType:        service.Workflow,
	}

	tc := PromTeamsConfig{
		Connectors: []map[string]string{
			{"/path1": "https://webhook1.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/id1/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=token1"},
			{"/path2": "https://webhook2.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/id2/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=token2"},
		},
	}

	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})
	converter, err := setupConverter(cfg, logger)
	require.NoError(t, err)

	httpClient := setupHTTPClient(cfg)

	routes, dRoutes, err := setupRoutes(cfg, tc, logger, converter, httpClient)
	require.NoError(t, err)

	assert.Len(t, routes, 2)  // Two connectors from config
	assert.Len(t, dRoutes, 1) // Dynamic route
}

func TestSetupRoutesWithTemplatedConnectors(t *testing.T) {
	cfg := Config{
		TemplateFile:       "../../default-message-workflow-card.tmpl",
		EscapeUnderscores:  true,
		ValidateWebhookURL: false,
		WebhookType:        service.Workflow,
	}

	tc := PromTeamsConfig{
		ConnectorsWithCustomTemplates: []ConnectorWithCustomTemplate{
			{
				RequestPath:       "/custom1",
				TemplateFile:      "../../default-message-workflow-card.tmpl",
				WebhookURL:        "https://custom1.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/id1/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=customtoken",
				EscapeUnderscores: true,
			},
		},
	}

	logger := setupLogger(Config{LogFormat: "json", DebugLogs: false})
	converter, err := setupConverter(cfg, logger)
	require.NoError(t, err)

	httpClient := setupHTTPClient(cfg)

	routes, dRoutes, err := setupRoutes(cfg, tc, logger, converter, httpClient)
	require.NoError(t, err)

	assert.Len(t, routes, 1)  // One templated connector
	assert.Len(t, dRoutes, 1) // Dynamic route
}

func TestValidateWebhookValidWorkflow(t *testing.T) {
	url := "https://test.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/testid/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=testtoken"
	err := validateWebhook(service.Workflow, url)
	assert.NoError(t, err)
}

func TestValidateWebhookInvalidHTTP(t *testing.T) {
	url := "http://test.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/testid/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=testtoken"
	err := validateWebhook(service.Workflow, url)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "https://")
}

func TestValidateWebhookWorkflowInvalidFormat(t *testing.T) {
	url := "https://invalid.com/not-a-workflow"
	err := validateWebhook(service.Workflow, url)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected format")
}

func TestValidateWebhook(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name    string
		args    args
		webhook service.WebhookType
		wantErr bool
	}{
		{name: "workflow webhook", webhook: service.Workflow, args: args{u: "https://example.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/b008d545fb784/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=Ogxlm1IT-Hs"}, wantErr: false},
		{name: "workflow webhook with different sig", webhook: service.Workflow, args: args{u: "https://example.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/testid/triggers/manual/paths/invoke?api-version=2&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=abc123_def-456"}, wantErr: false},
		{name: "missing https", webhook: service.Workflow, args: args{u: "example.cd.environment.api.powerplatform.com/powerautomate/test"}, wantErr: true},
		{name: "only http", webhook: service.Workflow, args: args{u: "http://example.cd.environment.api.powerplatform.com/powerautomate/test"}, wantErr: true},
		{name: "https but invalid format", webhook: service.Workflow, args: args{u: "https://example.com"}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := validateWebhook(tt.webhook, tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("validateWebhook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractWebhookFromRequest(t *testing.T) {
	workflowWebhook := "example.cd.environment.api.powerplatform.com/powerautomate/automations/direct/workflows/b008d545fb784/triggers/manual/paths/invoke?api-version=1&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=Ogxlm1IT-Hs"
	tests := []struct {
		name          string
		request       *http.Request
		webhookResult string
		wantErr       bool
	}{
		{
			name:          "workflow webhook in URL path",
			request:       newDummyRequest(fmt.Sprintf("/_dynamicwebhook/%s", workflowWebhook), ""),
			webhookResult: fmt.Sprintf("https://%s", workflowWebhook),
			wantErr:       false,
		},
		{
			name:          "workflow webhook in auth header",
			request:       newDummyRequest("/_dynamicwebhook/", fmt.Sprintf("webhook %s", workflowWebhook)),
			webhookResult: fmt.Sprintf("https://%s", workflowWebhook),
			wantErr:       false,
		},
		{
			name:    "invalid bearer",
			request: newDummyRequest("/_dynamicwebhook/", fmt.Sprintf("invalid-bearer %s", workflowWebhook)),
			wantErr: true,
		},
		{
			name:    "missing webhook and header",
			request: newDummyRequest("/_dynamicwebhook/", ""),
			wantErr: true,
		},
	}
	prefix := "/_dynamicwebhook/"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook, err := extractWebhookFromRequest(tt.request, prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractWebhookFromRequest() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.webhookResult != webhook {
				t.Errorf("extractWebhookFromRequest() webhook (\"%s\") does not match expected webhook (\"%s\")", webhook, tt.webhookResult)
			}
		})
	}
}

func newDummyRequest(urlPath, authHeader string) *http.Request {
	url := fmt.Sprintf("http://localhost:2000%s", urlPath)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req
}
