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

// Package main is the application entry point
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	ocprometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/labstack/echo/v4"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/stakater/prometheus-msteams/pkg/card"
	"github.com/stakater/prometheus-msteams/pkg/service"
	"github.com/stakater/prometheus-msteams/pkg/transport"
	"github.com/stakater/prometheus-msteams/pkg/utility"
	"github.com/stakater/prometheus-msteams/pkg/version"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"

	_ "net/http/pprof" //nolint: gosec

	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v2"
)

// PromTeamsConfig is the struct representation of the config file.
type PromTeamsConfig struct {
	// Connectors
	// The key is the request path for Prometheus to post to.
	// The value is the Teams webhook url.
	Connectors                    []map[string]string           `yaml:"connectors"`
	ConnectorsWithCustomTemplates []ConnectorWithCustomTemplate `yaml:"connectors_with_custom_templates"`
}

// ConnectorWithCustomTemplate .
type ConnectorWithCustomTemplate struct {
	RequestPath       string `yaml:"request_path"`
	TemplateFile      string `yaml:"template_file"`
	WebhookURL        string `yaml:"webhook_url"`
	EscapeUnderscores bool   `yaml:"escape_underscores"`
}

func parseTeamsConfigFile(f string) (PromTeamsConfig, error) {
	filePath := filepath.Clean(f)
	b, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return PromTeamsConfig{}, err
	}
	var tc PromTeamsConfig
	if err = yaml.Unmarshal(b, &tc); err != nil {
		return PromTeamsConfig{}, err
	}
	return tc, nil
}

// New Webhook URL format : https://devblogs.microsoft.com/microsoft365dev/retirement-of-office-365-connectors-within-microsoft-teams/
var validWebhookPatternO365 = regexp.MustCompile(`^[a-z0-9]+\.webhook\.office\.com/webhookb2/[a-z0-9\-]+@[a-z0-9\-]+/IncomingWebhook/[a-z0-9]+/[a-z0-9\-]+(/[a-zA-Z0-9\-]+)?$`)
var validWebhookPatternWorkflow = regexp.MustCompile((`^[a-z0-9\-\.]+\.environment\.api\.powerplatform\.com/powerautomate/automations/direct/workflows/[\w]+/triggers/manual/paths/invoke\?api-version=\d+&sp=%2Ftriggers%2Fmanual%2Frun&sv=1\.0&sig=[a-zA-Z0-9\-_]+`))
var legacyWebhookPrefix = "outlook.office.com/webhook/" // old format is only valid until 11. april '21

func validateWebhook(workflowType service.WebhookType, u string) error {
	path := strings.TrimPrefix(u, "https://")
	if u == path {
		return fmt.Errorf("the webhook_url must start with 'https://'. url: '%s'", u)
	}

	switch workflowType {
	case service.O365:
		isValidTeamsHook := validWebhookPatternO365.MatchString(path) || strings.HasPrefix(path, legacyWebhookPrefix)
		if !isValidTeamsHook {
			return fmt.Errorf("the webhook_url has an unexpected format '%s'", u)
		}
	case service.Workflow:
		isValidTeamsHook := validWebhookPatternWorkflow.MatchString(path)
		if !isValidTeamsHook {
			return fmt.Errorf("the webhook_url has an unexpected format '%s'", u)
		}
		return nil
	}
	return nil
}

//nolint:gocyclo
func main() {
	// Parse configuration and flags
	cfg, err := parseFlags()
	if err != nil {
		// G705: XSS via taint analysis
		fmt.Fprintf(os.Stderr, "%q\n", err.Error()) //nolint:gosec
		os.Exit(1)
	}

	if cfg.Version {
		fmt.Println(version.VERSION)
		os.Exit(0)
	}

	// Setup logger
	logger := setupLogger(cfg)

	// Setup tracer
	if err := setupTracer(cfg, logger); err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	// Parse the config file if defined
	var tc PromTeamsConfig
	if cfg.ConfigFile != "" {
		tc, err = parseTeamsConfigFile(cfg.ConfigFile)
		if err != nil {
			logger.Err(err)
			os.Exit(1)
		}
	}

	// Setup converter
	defaultConverter, err := setupConverter(cfg, logger)
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	// Setup HTTP client
	httpClient := setupHTTPClient(cfg)

	// Setup routes
	routes, dRoutes, err := setupRoutes(cfg, tc, logger, defaultConverter, httpClient)
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	// Setup Prometheus exporter
	pe, err := setupPrometheusExporter(logger)
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	// Setup server
	handler := setupServer(logger, routes, dRoutes, tc, pe)

	// Setup run group
	var g run.Group
	{
		srv := http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           handler,
			ReadHeaderTimeout: 30 * time.Second,
		}
		g.Add(
			func() error {
				logger.Info(
					"listen_http_addr", cfg.HTTPAddr,
					"version", version.VERSION,
					"commit", version.COMMIT,
					"branch", version.BRANCH,
					"build_date", version.BUILDDATE,
				)
				return srv.ListenAndServe()
			},
			func(error) {
				if err != http.ErrServerClosed {
					if err := srv.Shutdown(context.Background()); err != nil {
						logger.Err(err)
					}
				}
			},
		)
	}
	{
		g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	}
	logger.Info("exit", g.Run())
}

func ocviews() []*view.View {
	clientKeys := []tag.Key{
		ochttp.KeyClientMethod, ochttp.KeyClientStatus, ochttp.KeyClientHost, ochttp.KeyClientPath,
	}
	serverKeys := []tag.Key{
		ochttp.StatusCode, ochttp.Method, ochttp.Path,
	}
	return []*view.View{
		// HTTP client metrics.
		{
			Name:        "http/client/sent_bytes",
			Measure:     ochttp.ClientSentBytes,
			Aggregation: view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304),
			Description: "Total bytes sent in request body (not including headers), by HTTP method and response status",
			TagKeys:     clientKeys,
		},
		{
			Name:        "http/client/received_bytes",
			Measure:     ochttp.ClientReceivedBytes,
			Aggregation: view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304),
			Description: "Total bytes received in response bodies (not including headers but including error responses with bodies), by HTTP method and response status",
			TagKeys:     clientKeys,
		},
		{
			Name:        "http/client/roundtrip_latency",
			Measure:     ochttp.ClientRoundtripLatency,
			Aggregation: view.Distribution(1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30),
			Description: "End-to-end latency, by HTTP method and response status",
			TagKeys:     clientKeys,
		},
		{
			Name:        "http/client/completed_count",
			Measure:     ochttp.ClientRoundtripLatency,
			Aggregation: view.Count(),
			Description: "Count of completed requests, by HTTP method and response status",
			TagKeys:     clientKeys,
		},
		// HTTP server metrics.
		{
			Name:        "http/server/request_count",
			Description: "Count of HTTP requests started",
			Measure:     ochttp.ServerRequestCount,
			Aggregation: view.Count(),
			TagKeys:     serverKeys,
		},
		{
			Name:        "http/server/request_bytes",
			Description: "Size distribution of HTTP request body",
			Measure:     ochttp.ServerRequestBytes,
			Aggregation: view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304),
			TagKeys:     serverKeys,
		},
		{
			Name:        "http/server/response_bytes",
			Description: "Size distribution of HTTP response body",
			Measure:     ochttp.ServerResponseBytes,
			Aggregation: view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304),
			TagKeys:     serverKeys,
		},
		{
			Name:        "http/server/latency",
			Description: "Latency distribution of HTTP requests",
			Measure:     ochttp.ServerLatency,
			Aggregation: view.Distribution(1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30),
			TagKeys:     serverKeys,
		},
	}
}

func checkDuplicateRequestPath(routes []transport.Route) error {
	added := map[string]bool{}
	for _, r := range routes {
		if _, ok := added[r.RequestPath]; ok {
			return fmt.Errorf("found duplicate use of request path '%s'", r.RequestPath)
		}
		added[r.RequestPath] = true
	}
	return nil
}

const bearerType = "webhook"

func extractWebhookFromRequest(request *http.Request, requestPathPrefix string) (string, error) {
	var pathAndQuery string

	// Query contains mandatory "api-version" for "worflow"
	if len(request.URL.Query()) > 0 {
		pathAndQuery = request.URL.Path + "?" + request.URL.RawQuery
	} else {
		pathAndQuery = request.URL.Path
	}

	pathAndQuery = strings.TrimPrefix(pathAndQuery, requestPathPrefix)
	if pathAndQuery != "" {
		return fmt.Sprintf("https://%s", pathAndQuery), nil
	}
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("neither webhook url nor bearer authorization present")
	}

	if !strings.HasPrefix(authHeader, bearerType) {
		return "", fmt.Errorf("invalid bearer on authorization")
	}
	pathAndQuery = strings.TrimPrefix(authHeader, fmt.Sprintf("%s ", bearerType))
	return fmt.Sprintf("https://%s", pathAndQuery), nil
}

// Config holds all application configuration
type Config struct {
	Version                       bool
	LogFormat                     string
	DebugLogs                     bool
	JaegerTrace                   bool
	JaegerAgentAddr               string
	HTTPAddr                      string
	RequestURI                    string
	TeamsWebhookURL               string
	TemplateFile                  string
	EscapeUnderscores             bool
	ConfigFile                    string
	HTTPClientIdleConnTimeout     time.Duration
	HTTPClientTLSHandshakeTimeout time.Duration
	HTTPClientMaxIdleConn         int
	InsecureSkipVerify            bool
	RetryMax                      int
	ValidateWebhookURL            bool
	WebhookType                   service.WebhookType
}

func parseFlags() (Config, error) {
	var (
		fs                            = flag.NewFlagSet("prometheus-msteams", flag.ExitOnError)
		promVersion                   = fs.Bool("version", false, "Print the version")
		logFormat                     = fs.String("log-format", "json", "json|fmt")
		debugLogs                     = fs.Bool("debug", false, "Set log level to debug mode.")
		jaegerTrace                   = fs.Bool("jaeger-trace", false, "Send traces to Jaeger.")
		jaegerAgentAddr               = fs.String("jaeger-agent", "localhost:6831", "Jaeger agent endpoint")
		httpAddr                      = fs.String("http-addr", ":2000", "HTTP listen address.")
		requestURI                    = fs.String("teams-request-uri", "", "The default request URI path where Prometheus will post to.")
		teamsWebhookURL               = fs.String("teams-incoming-webhook-url", "", "The default Microsoft Teams webhook connector.")
		templateFile                  = fs.String("template-file", "", "The Microsoft Teams Message Card template file.")
		escapeUnderscores             = fs.Bool("auto-escape-underscores", true, "Automatically replace all '_' with '\\_' from texts in the alert.")
		configFile                    = fs.String("config-file", "", "The connectors configuration file.")
		httpClientIdleConnTimeout     = fs.Duration("idle-conn-timeout", 90*time.Second, "The HTTP client idle connection timeout duration.")
		httpClientTLSHandshakeTimeout = fs.Duration("tls-handshake-timeout", 30*time.Second, "The HTTP client TLS handshake timeout.")
		httpClientMaxIdleConn         = fs.Int("max-idle-conns", 100, "The HTTP client maximum number of idle connections")
		insecureSkipVerify            = fs.Bool("insecure-skip-verify", false, "Disable validation of the server certificate.")
		retryMax                      = fs.Int("max-retry-count", 3, "The retry maximum for sending requests to the webhook")
		validateWebhookURL            = fs.Bool("validate-webhook-url", false, "Enforce strict validation of webhook url")
	)

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		return Config{}, err
	}

	if *templateFile == "" {
		*templateFile = "./default-message-workflow-card.tmpl"
	}

	return Config{
		Version:                       *promVersion,
		LogFormat:                     *logFormat,
		DebugLogs:                     *debugLogs,
		JaegerTrace:                   *jaegerTrace,
		JaegerAgentAddr:               *jaegerAgentAddr,
		HTTPAddr:                      *httpAddr,
		RequestURI:                    *requestURI,
		TeamsWebhookURL:               *teamsWebhookURL,
		TemplateFile:                  *templateFile,
		EscapeUnderscores:             *escapeUnderscores,
		ConfigFile:                    *configFile,
		HTTPClientIdleConnTimeout:     *httpClientIdleConnTimeout,
		HTTPClientTLSHandshakeTimeout: *httpClientTLSHandshakeTimeout,
		HTTPClientMaxIdleConn:         *httpClientMaxIdleConn,
		InsecureSkipVerify:            *insecureSkipVerify,
		RetryMax:                      *retryMax,
		ValidateWebhookURL:            *validateWebhookURL,
		WebhookType:                   service.Workflow,
	}, nil
}

func setupLogger(cfg Config) *utility.Logger {
	logger := utility.NewLogger(utility.LoggerFormat(cfg.LogFormat), cfg.DebugLogs)
	logger.Debug("webhook-type", cfg.WebhookType)
	return logger
}

func setupTracer(cfg Config, logger *utility.Logger) error {
	if !cfg.JaegerTrace {
		return nil
	}

	logger.Log("message", "jaeger tracing enabled")

	je, err := jaeger.NewExporter(
		jaeger.Options{
			AgentEndpoint: cfg.JaegerAgentAddr,
			ServiceName:   "prometheus-msteams",
		},
	)
	if err != nil {
		return err
	}

	trace.RegisterExporter(je)
	trace.ApplyConfig(
		trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		},
	)
	return nil
}

func setupConverter(cfg Config, logger *utility.Logger) (card.Converter, error) {
	tmpl, err := card.ParseTemplateFile(cfg.TemplateFile)
	if err != nil {
		return nil, err
	}
	converter := card.NewTemplatedCardCreator(tmpl, cfg.EscapeUnderscores, logger)
	converter = card.NewCreatorLoggingMiddleware(
		logger.With(
			"template_file", cfg.TemplateFile,
			"escaped_underscores", cfg.EscapeUnderscores,
		),
		converter,
	)
	return converter, nil
}

func setupHTTPClient(cfg Config) *http.Client {
	retryClient := retryablehttp.NewClient()
	if !cfg.DebugLogs {
		retryClient.Logger = nil
	}
	retryClient.RetryMax = cfg.RetryMax
	retryClient.HTTPClient = &http.Client{
		Transport: &ochttp.Transport{
			Base: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          cfg.HTTPClientMaxIdleConn,
				IdleConnTimeout:       cfg.HTTPClientIdleConnTimeout,
				TLSHandshakeTimeout:   cfg.HTTPClientTLSHandshakeTimeout,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint: gosec
			},
		},
	}
	return retryClient.StandardClient()
}

func setupRoutes(cfg Config, tc PromTeamsConfig, logger *utility.Logger, defaultConverter card.Converter, httpClient *http.Client) ([]transport.Route, []transport.DynamicRoute, error) {
	var routes []transport.Route
	var dRoutes []transport.DynamicRoute

	// Connectors from flags.
	if len(cfg.RequestURI) > 0 && len(cfg.TeamsWebhookURL) > 0 {
		tc.Connectors = append(
			tc.Connectors,
			map[string]string{
				cfg.RequestURI: cfg.TeamsWebhookURL,
			},
		)
	}

	// Dynamic uri handler: webhook uri is retrieved from request.URL
	var dr transport.DynamicRoute
	dr.RequestPath = "/_dynamicwebhook/*"
	dr.ServiceGenerator = func(c echo.Context) (service.Service, error) {
		webhook, err := extractWebhookFromRequest(c.Request(), "/_dynamicwebhook/")
		if err != nil {
			err = errors.Wrapf(err, "webhook extraction failed for /_dynamicwebhook/")
			logger.Err(err)
			return nil, err
		}

		err = validateWebhook(cfg.WebhookType, webhook)
		if cfg.ValidateWebhookURL && err != nil {
			err = errors.Wrapf(err, "webhook validation failed for /_dynamicwebhook/")
			logger.Err(err)
			return nil, err
		}

		var s service.Service
		s = service.NewSimpleService(defaultConverter, httpClient, webhook, cfg.WebhookType)
		s = service.NewLoggingService(logger, s)
		return s, nil
	}
	dRoutes = append(dRoutes, dr)

	configRoutes, err := connectorsFromConfig(tc, cfg, logger, defaultConverter, httpClient)
	if err != nil {
		return nil, nil, err
	}
	routes = append(routes, configRoutes...)

	templateRoutes, err := connectorsFromTemplate(tc, cfg, logger, httpClient)
	if err != nil {
		return nil, nil, err
	}
	routes = append(routes, templateRoutes...)

	if err := checkDuplicateRequestPath(routes); err != nil {
		return nil, nil, err
	}

	return routes, dRoutes, nil
}

func connectorsFromConfig(tc PromTeamsConfig, cfg Config, logger *utility.Logger, defaultConverter card.Converter, httpClient *http.Client) ([]transport.Route, error) {
	var routes []transport.Route
	// Connectors from config file.
	for _, c := range tc.Connectors {
		for uri, webhook := range c {
			err := validateWebhook(cfg.WebhookType, webhook)
			if cfg.ValidateWebhookURL && err != nil {
				return nil, err
			}

			var r transport.Route
			r.RequestPath = uri
			r.Service = service.NewSimpleService(defaultConverter, httpClient, webhook, cfg.WebhookType)
			r.Service = service.NewLoggingService(logger, r.Service)
			routes = append(routes, r)
		}
	}
	return routes, nil
}

func connectorsFromTemplate(tc PromTeamsConfig, cfg Config, logger *utility.Logger, httpClient *http.Client) ([]transport.Route, error) {
	var routes []transport.Route
	// Connectors with custom template files.
	for _, c := range tc.ConnectorsWithCustomTemplates {
		if len(c.RequestPath) == 0 {
			return nil, fmt.Errorf("one of the 'templated_connectors' is missing a 'request_path'")
		}
		if len(c.WebhookURL) == 0 {
			return nil, fmt.Errorf("the webhook_url is required for request_path '%s'", c.RequestPath)
		}
		err := validateWebhook(cfg.WebhookType, c.WebhookURL)
		if cfg.ValidateWebhookURL && err != nil {
			return nil, err
		}
		if len(c.TemplateFile) == 0 {
			return nil, fmt.Errorf("the template_file is required for request_path '%s'", c.RequestPath)
		}

		tmpl, err := card.ParseTemplateFile(c.TemplateFile)
		if err != nil {
			return nil, err
		}

		converter := card.NewTemplatedCardCreator(tmpl, c.EscapeUnderscores, logger)
		converter = card.NewCreatorLoggingMiddleware(
			logger.With(
				"template_file", c.TemplateFile,
				"escaped_underscores", c.EscapeUnderscores,
			),
			converter,
		)

		var r transport.Route
		r.RequestPath = c.RequestPath
		r.Service = service.NewSimpleService(converter, httpClient, c.WebhookURL, cfg.WebhookType)
		r.Service = service.NewLoggingService(logger, r.Service)
		routes = append(routes, r)
	}
	return routes, nil
}

func setupPrometheusExporter(_ *utility.Logger) (*ocprometheus.Exporter, error) {
	pe, err := ocprometheus.NewExporter(
		ocprometheus.Options{
			Registry: stdprometheus.DefaultRegisterer.(*stdprometheus.Registry),
		},
	)
	if err != nil {
		return nil, err
	}
	if err := view.Register(ocviews()...); err != nil {
		return nil, err
	}
	return pe, nil
}

func setupServer(logger *utility.Logger, routes []transport.Route, dRoutes []transport.DynamicRoute, tc PromTeamsConfig, pe *ocprometheus.Exporter) *echo.Echo {
	// Main app.
	handler := transport.NewServer(logger.GetLogger(), routes, dRoutes)
	// Prometheus metrics.
	handler.GET("/metrics", echo.WrapHandler(pe))
	// Pprof.
	handler.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	// Config.
	handler.GET("/config", func(c echo.Context) error {
		return c.JSON(200, tc.Connectors)
	})
	return handler
}
