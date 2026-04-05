package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/database"
	"github.com/mpa-forge/backend-api/internal/usersvc"
	"github.com/mpa-forge/platform-observability/backendobs"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

const (
	httpRequestCountMetricName    = "http.server.request.count"
	httpRequestDurationMetricName = "http.server.request.duration"
)

func TestRequestLoggerAddsTraceCorrelationFields(t *testing.T) {
	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	traceProvider := sdktrace.NewTracerProvider()
	t.Cleanup(func() {
		if err := traceProvider.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown() error = %v", err)
		}
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	tracer := traceProvider.Tracer("request-logger-test")
	instrumented := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "request-span")
		defer span.End()

		handlerWithLogging := requestLogger(logger, nil)(handler)
		handlerWithLogging.ServeHTTP(w, r.WithContext(ctx))
	})

	recorder := httptest.NewRecorder()
	instrumented.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("Unmarshal() error = %v; logs = %s", err, logs.String())
	}

	if _, ok := entry["trace_id"]; !ok {
		t.Fatalf("log entry = %#v, want trace_id", entry)
	}
	if _, ok := entry["span_id"]; !ok {
		t.Fatalf("log entry = %#v, want span_id", entry)
	}
	if sampled, ok := entry["trace_sampled"].(bool); !ok || !sampled {
		t.Fatalf("trace_sampled = %#v, want true", entry["trace_sampled"])
	}
}

func TestNewRouterWithDisabledObservabilityRuntimeStillServesRequests(t *testing.T) {
	obsRuntime, err := backendobs.Init(context.Background(), backendobs.Config{
		ServiceName: "backend-api",
		Environment: "test",
		Mode:        backendobs.ModeDisabled,
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer func() {
		if err := obsRuntime.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown() error = %v", err)
		}
	}()

	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		obsRuntime,
		staticVerifier{principal: auth.Principal{UserID: "user_123", Email: "user@example.com", DisplayName: "Example User", Role: auth.RoleUser}},
		usersvc.NewServer(&fakeProfileStore{
			getProfile: databaseUserProfile("user_123", "user@example.com", "Example User", "user"),
		}),
	)

	healthResp := httptest.NewRecorder()
	handler.ServeHTTP(healthResp, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if healthResp.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthResp.Code, http.StatusOK)
	}

	connectReq := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/GetCurrentUser", strings.NewReader("{}"))
	connectReq.Header.Set("Authorization", "Bearer valid-token")
	connectReq.Header.Set("Content-Type", "application/json")
	connectResp := httptest.NewRecorder()
	handler.ServeHTTP(connectResp, connectReq)
	if connectResp.Code != http.StatusOK {
		t.Fatalf("connect status = %d, want %d; body = %s", connectResp.Code, http.StatusOK, connectResp.Body.String())
	}
}

func TestPublicEndpointExportsServerTelemetry(t *testing.T) {
	capture := newOTLPCapture(t)
	obsRuntime := newTestObservabilityRuntime(t, capture.server.URL)

	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		obsRuntime,
		staticVerifier{},
		usersvc.NewServer(&fakeProfileStore{}),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}

	shutdownObservabilityRuntime(t, obsRuntime)

	span := capture.findSpanByName(t, "GET /healthz")
	if got := stringTraceAttr(span.Attributes, "http.request.method"); got != http.MethodGet {
		t.Fatalf("http.request.method = %q, want %q", got, http.MethodGet)
	}
	if got := stringTraceAttr(span.Attributes, "http.route"); got != "/healthz" {
		t.Fatalf("http.route = %q, want %q", got, "/healthz")
	}
	if got := intTraceAttr(span.Attributes, "http.response.status_code"); got != http.StatusOK {
		t.Fatalf("http.response.status_code = %d, want %d", got, http.StatusOK)
	}

	if !capture.hasMetricDatapoint(
		httpRequestCountMetricName,
		map[string]string{
			"http.request.method":        http.MethodGet,
			"http.route":                 "/healthz",
			"http.response.status_class": "2xx",
		},
		map[string]int64{"http.response.status_code": http.StatusOK},
	) {
		t.Fatalf("request count metric missing normalized attributes")
	}
	if !capture.hasMetricDatapoint(
		httpRequestDurationMetricName,
		map[string]string{
			"http.request.method":        http.MethodGet,
			"http.route":                 "/healthz",
			"http.response.status_class": "2xx",
		},
		map[string]int64{"http.response.status_code": http.StatusOK},
	) {
		t.Fatalf("request duration metric missing normalized attributes")
	}
}

func TestProtectedConnectRequestExportsProcedureTelemetry(t *testing.T) {
	capture := newOTLPCapture(t)
	obsRuntime := newTestObservabilityRuntime(t, capture.server.URL)

	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		obsRuntime,
		staticVerifier{principal: auth.Principal{UserID: "user_123", Email: "user@example.com", DisplayName: "Example User", Role: auth.RoleUser}},
		usersvc.NewServer(&fakeProfileStore{
			getProfile: databaseUserProfile("user_123", "stored@example.com", "Stored User", "user"),
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/GetCurrentUser", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	shutdownObservabilityRuntime(t, obsRuntime)

	span := capture.findSpanByName(t, "POST /blueprint.user.v1.UserService/GetCurrentUser")
	if got := stringTraceAttr(span.Attributes, "rpc.system"); got != "connect" {
		t.Fatalf("rpc.system = %q, want %q", got, "connect")
	}
	if got := stringTraceAttr(span.Attributes, "rpc.connect.procedure"); got != "/blueprint.user.v1.UserService/GetCurrentUser" {
		t.Fatalf("rpc.connect.procedure = %q, want %q", got, "/blueprint.user.v1.UserService/GetCurrentUser")
	}
	if got := stringTraceAttr(span.Attributes, "auth.result"); got != "authenticated" {
		t.Fatalf("auth.result = %q, want %q", got, "authenticated")
	}
	if !capture.hasMetricDatapoint(
		httpRequestCountMetricName,
		map[string]string{
			"http.request.method":        http.MethodPost,
			"http.route":                 "/blueprint.user.v1.UserService/GetCurrentUser",
			"http.response.status_class": "2xx",
		},
		map[string]int64{"http.response.status_code": http.StatusOK},
	) {
		t.Fatalf("connect request count metric missing procedure route attributes")
	}
}

func TestProtectedConnectFailuresAnnotateActiveSpan(t *testing.T) {
	tests := []struct {
		name                string
		verifier            auth.Verifier
		store               *fakeProfileStore
		authorizationHeader string
		wantStatus          int
		wantConnectCode     string
		wantAuthResult      string
	}{
		{
			name:            "missing bearer token",
			verifier:        staticVerifier{},
			store:           &fakeProfileStore{},
			wantStatus:      http.StatusUnauthorized,
			wantConnectCode: "unauthenticated",
			wantAuthResult:  "missing_bearer_token",
		},
		{
			name:                "handler failure",
			verifier:            staticVerifier{principal: auth.Principal{UserID: "user_123", Email: "user@example.com", DisplayName: "Example User", Role: auth.RoleUser}},
			store:               &fakeProfileStore{getErr: errors.New("boom")},
			authorizationHeader: "Bearer valid-token",
			wantStatus:          http.StatusInternalServerError,
			wantConnectCode:     "internal",
			wantAuthResult:      "authenticated",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture := newOTLPCapture(t)
			obsRuntime := newTestObservabilityRuntime(t, capture.server.URL)

			handler := newRouter(
				config.Config{AppEnv: "test"},
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				obsRuntime,
				test.verifier,
				usersvc.NewServer(test.store),
			)

			req := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/GetCurrentUser", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			if test.authorizationHeader != "" {
				req.Header.Set("Authorization", test.authorizationHeader)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}

			shutdownObservabilityRuntime(t, obsRuntime)

			span := capture.findSpanByName(t, "POST /blueprint.user.v1.UserService/GetCurrentUser")
			if got := stringTraceAttr(span.Attributes, "rpc.connect.code"); got != test.wantConnectCode {
				t.Fatalf("rpc.connect.code = %q, want %q", got, test.wantConnectCode)
			}
			if got := stringTraceAttr(span.Attributes, "auth.result"); got != test.wantAuthResult {
				t.Fatalf("auth.result = %q, want %q", got, test.wantAuthResult)
			}
			if span.Status == nil || span.Status.Code != tracepb.Status_STATUS_CODE_ERROR {
				t.Fatalf("span status = %#v, want error status code", span.Status)
			}
		})
	}
}

func newTestObservabilityRuntime(t *testing.T, endpoint string) *backendobs.Runtime {
	t.Helper()

	runtime, err := backendobs.Init(context.Background(), backendobs.Config{
		ServiceName:            "backend-api",
		ServiceVersion:         "test",
		Environment:            "test",
		Mode:                   backendobs.ModeDirectOTLP,
		Profile:                backendobs.ProfileBalanced,
		OTLPEndpoint:           endpoint,
		GrafanaCloudInstanceID: "1546554",
		GrafanaOTLPIngestToken: "token-value",
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	return runtime
}

func shutdownObservabilityRuntime(t *testing.T, runtime *backendobs.Runtime) {
	t.Helper()

	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

type otlpCapture struct {
	server         *httptest.Server
	mu             sync.Mutex
	traceRequests  []*coltracepb.ExportTraceServiceRequest
	metricRequests []*colmetricpb.ExportMetricsServiceRequest
}

func newOTLPCapture(t *testing.T) *otlpCapture {
	t.Helper()

	capture := &otlpCapture{}
	capture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := readRequestPayload(r)
		if err != nil {
			t.Errorf("readRequestPayload() error = %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		capture.mu.Lock()
		defer capture.mu.Unlock()

		switch r.URL.Path {
		case "/v1/traces":
			var req coltracepb.ExportTraceServiceRequest
			if err := proto.Unmarshal(payload, &req); err != nil {
				t.Errorf("trace proto.Unmarshal() error = %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			capture.traceRequests = append(capture.traceRequests, &req)
		case "/v1/metrics":
			var req colmetricpb.ExportMetricsServiceRequest
			if err := proto.Unmarshal(payload, &req); err != nil {
				t.Errorf("metric proto.Unmarshal() error = %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			capture.metricRequests = append(capture.metricRequests, &req)
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(capture.server.Close)

	return capture
}

func readRequestPayload(r *http.Request) ([]byte, error) {
	var reader io.ReadCloser = r.Body
	if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		reader = io.NopCloser(gzipReader)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (c *otlpCapture) findSpanByName(t *testing.T, name string) *tracepb.Span {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, req := range c.traceRequests {
		for _, resourceSpans := range req.ResourceSpans {
			for _, scopeSpans := range resourceSpans.ScopeSpans {
				for _, span := range scopeSpans.Spans {
					if span.Name == name {
						return span
					}
				}
			}
		}
	}

	t.Fatalf("span %q not found in %#v", name, c.traceRequests)
	return nil
}

func (c *otlpCapture) hasMetricDatapoint(name string, wantStrings map[string]string, wantInts map[string]int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, req := range c.metricRequests {
		for _, resourceMetrics := range req.ResourceMetrics {
			for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
				for _, metric := range scopeMetrics.Metrics {
					if metric.Name != name {
						continue
					}
					if metricDataMatches(metric, wantStrings, wantInts) {
						return true
					}
				}
			}
		}
	}

	return false
}

func metricDataMatches(metric *metricpb.Metric, wantStrings map[string]string, wantInts map[string]int64) bool {
	switch data := metric.Data.(type) {
	case *metricpb.Metric_Sum:
		for _, point := range data.Sum.DataPoints {
			if attrsMatch(point.Attributes, wantStrings, wantInts) {
				return true
			}
		}
	case *metricpb.Metric_Histogram:
		for _, point := range data.Histogram.DataPoints {
			if attrsMatch(point.Attributes, wantStrings, wantInts) {
				return true
			}
		}
	}

	return false
}

func attrsMatch(attrs []*commonpb.KeyValue, wantStrings map[string]string, wantInts map[string]int64) bool {
	for key, want := range wantStrings {
		if got := stringTraceAttr(attrs, key); got != want {
			return false
		}
	}
	for key, want := range wantInts {
		if got := intTraceAttr(attrs, key); int64(got) != want {
			return false
		}
	}

	return true
}

func stringTraceAttr(attrs []*commonpb.KeyValue, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value.GetStringValue()
		}
	}

	return ""
}

func intTraceAttr(attrs []*commonpb.KeyValue, key string) int {
	for _, attr := range attrs {
		if attr.Key == key {
			return int(attr.Value.GetIntValue())
		}
	}

	return 0
}

func databaseUserProfile(clerkUserID, email, displayName, role string) database.UserProfile {
	return database.UserProfile{
		ClerkUserID: clerkUserID,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	}
}
