package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beratkara/goralens"
	contractshttp "github.com/goravel/framework/contracts/http"

	"github.com/beratkara/goralens/examples/basic/app/facades"
	"github.com/beratkara/goralens/examples/basic/app/watchers"
	"github.com/beratkara/goralens/examples/basic/bootstrap"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "goralens-example")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("DB_DATABASE", filepath.Join(dir, "database.sqlite"))
	_ = os.Setenv("APP_ENV", "local")
	bootstrap.Boot()
	if err := facades.App().MakeArtisan().Call("migrate"); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func serve(t *testing.T, method, path, body string) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet && strings.HasPrefix(path, "/goralens/") {
		req.Header.Set("X-Goralens", "1")
	}
	req.Header.Set("Authorization", "Bearer should-be-hidden")
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

type entry struct {
	UUID    string         `json:"uuid"`
	BatchID string         `json:"batch_id"`
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
	Tags    []string       `json:"tags"`
}

func entries(t *testing.T, query string) []entry {
	t.Helper()
	goralens.Instance().Flush(5 * time.Second)
	code, body := serve(t, http.MethodGet, "/goralens/api/entries?"+query, "")
	if code != http.StatusOK {
		t.Fatalf("entries api: %d %s", code, body)
	}
	var data struct{ Entries []entry }
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatal(err)
	}
	return data.Entries
}

func detail(t *testing.T, id string) (entry, []entry) {
	t.Helper()
	code, body := serve(t, http.MethodGet, "/goralens/api/entries/"+id, "")
	if code != http.StatusOK {
		t.Fatalf("detail api: %d %s", code, body)
	}
	var data struct {
		Entry entry
		Batch []entry
	}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatal(err)
	}
	return data.Entry, data.Batch
}

func findRequest(t *testing.T, method, uri string) entry {
	t.Helper()
	for _, e := range entries(t, "type=request&take=200") {
		if e.Content["method"] == method && e.Content["uri"] == uri {
			return e
		}
	}
	t.Fatalf("no request entry for %s %s", method, uri)
	return entry{}
}

func types(batch []entry) map[string]int {
	out := map[string]int{}
	for _, e := range batch {
		out[e.Type]++
	}
	return out
}

func TestRequestBatchHoldsEverythingItDid(t *testing.T) {
	code, body := serve(t, http.MethodPost, "/users", `{"name":"Ada","email":"ada@example.com","password":"hunter2"}`)
	if code != http.StatusCreated {
		t.Fatalf("create user: %d %s", code, body)
	}
	req := findRequest(t, "POST", "/users")
	if req.Content["response_status"] != float64(201) {
		t.Fatalf("status: %v", req.Content["response_status"])
	}
	if req.Content["payload"].(map[string]any)["password"] != "********" {
		t.Fatalf("password stored: %v", req.Content["payload"])
	}
	if req.Content["headers"].(map[string]any)["authorization"] != "********" {
		t.Fatal("authorization header stored")
	}
	if req.Content["response"].(map[string]any)["token"] != "********" {
		t.Fatalf("response token stored: %v", req.Content["response"])
	}

	_, batch := detail(t, req.UUID)
	got := types(batch)
	for _, want := range []string{"query", "model", "event", "job", "log"} {
		if got[want] == 0 {
			t.Errorf("batch has no %s entry: %v", want, got)
		}
	}
	for _, e := range batch {
		if e.Type == "model" && e.Content["action"] == "created" {
			if e.Content["changes"].(map[string]any)["password"] != "********" {
				t.Errorf("model password stored: %v", e.Content["changes"])
			}
		}
		if e.Type == "query" && strings.Contains(e.Content["sql"].(string), "goralens_") {
			t.Errorf("Goralens recorded its own query: %v", e.Content["sql"])
		}
	}
}

func TestReadsCacheLogsAndDumps(t *testing.T) {
	if code, _ := serve(t, http.MethodGet, "/users", ""); code != http.StatusOK {
		t.Fatalf("list users: %d", code)
	}
	_, batch := detail(t, findRequest(t, "GET", "/users").UUID)
	got := types(batch)
	if got["cache"] == 0 || got["dump"] == 0 || got["log"] == 0 || got["query"] == 0 {
		t.Fatalf("batch: %v", got)
	}
	for _, e := range batch {
		if e.Type == "query" && e.Content["file"] == nil {
			t.Errorf("query without location: %v", e.Content)
		}
	}
}

func TestPanicsAndErrorLogsBecomeExceptions(t *testing.T) {
	if code, _ := serve(t, http.MethodGet, "/panic", ""); code != http.StatusInternalServerError {
		t.Fatalf("panic route returned %d", code)
	}
	serve(t, http.MethodGet, "/panic", "")
	serve(t, http.MethodGet, "/error-log", "")

	exceptions := entries(t, "type=exception")
	byClass := map[string]entry{}
	for _, e := range exceptions {
		byClass[e.Content["class"].(string)] = e
	}
	panicEntry, ok := byClass["string"]
	if !ok || panicEntry.Content["message"] != "something broke" {
		t.Fatalf("panic not recorded: %v", exceptions)
	}
	if panicEntry.Content["occurrences"] != float64(2) {
		t.Errorf("panic occurrences: %v", panicEntry.Content["occurrences"])
	}
	if !strings.HasSuffix(panicEntry.Content["file"].(string), "routes/web.go") {
		t.Errorf("panic location: %v", panicEntry.Content["file"])
	}
	if logged, ok := byClass["Log error"]; !ok {
		t.Errorf("error log not recorded as exception: %v", byClass)
	} else if file, _ := logged.Content["file"].(string); !strings.HasSuffix(file, "routes/web.go") {
		t.Errorf("error log location: %v", logged.Content["file"])
	}
	if req := findRequest(t, "GET", "/panic"); req.Content["response_status"] != float64(500) {
		t.Errorf("panic request status: %v", req.Content["response_status"])
	}
}

func TestGateChecks(t *testing.T) {
	serve(t, http.MethodGet, "/reports", "")
	for _, e := range entries(t, "type=gate") {
		if e.Content["ability"] == "view-reports" && e.Content["result"] == "denied" {
			return
		}
	}
	t.Fatal("gate check not recorded")
}

func TestCommandsAreRecorded(t *testing.T) {
	found := false
	for _, e := range entries(t, "type=command") {
		if e.Content["command"] == "migrate" {
			found = true
		}
	}
	if !found {
		t.Fatal("migrate command not recorded")
	}
}

func TestDashboardAndAuthorization(t *testing.T) {
	code, body := serve(t, http.MethodGet, "/goralens", "")
	if code != http.StatusOK || !strings.Contains(string(body), `src="/goralens/app.js"`) {
		t.Fatalf("dashboard: %d %s", code, body)
	}
	if code, _ := serve(t, http.MethodGet, "/goralens/app.js", ""); code != http.StatusOK {
		t.Fatalf("script: %d", code)
	}

	req := httptest.NewRequest(http.MethodDelete, "/goralens/api/entries", nil)
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("delete without header: %d", rec.Code)
	}

	goralens.Auth(func(ctx contractshttp.Context) bool { return ctx.Request().Header("X-Admin") == "yes" })
	defer goralens.Auth(nil)
	if code, _ := serve(t, http.MethodGet, "/goralens/api/status", ""); code != http.StatusForbidden {
		t.Fatalf("unauthorized status: %d", code)
	}
}

func TestMonitoringAndRecordingToggle(t *testing.T) {
	if code, body := serve(t, http.MethodPost, "/goralens/api/monitored-tags", `{"tags":["Auth:1"]}`); code != http.StatusOK || !strings.Contains(string(body), "Auth:1") {
		t.Fatalf("monitor: %d %s", code, body)
	}
	if code, _ := serve(t, http.MethodPost, "/goralens/api/recording", `{"recording":false}`); code != http.StatusOK {
		t.Fatal("pause failed")
	}
	before := len(entries(t, "type=request&take=200"))
	serve(t, http.MethodGet, "/users", "")
	if after := len(entries(t, "type=request&take=200")); after != before {
		t.Fatalf("recorded while paused: %d -> %d", before, after)
	}
	serve(t, http.MethodPost, "/goralens/api/recording", `{"recording":true}`)
}

func TestPruneCommand(t *testing.T) {
	if err := facades.App().MakeArtisan().Call("goralens:prune --hours=0"); err != nil {
		t.Fatal(err)
	}
	if left := entries(t, "take=200"); len(left) > 1 {
		t.Fatalf("%d entries left after pruning", len(left))
	}
}

func TestDashboardMiddlewareChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/goralens/api/status", nil)
	req.Header.Set("X-Block", "yes")
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("a refusing middleware must close the dashboard: %d", rec.Code)
	}
	if code, _ := serve(t, http.MethodGet, "/goralens/api/status", ""); code != http.StatusOK {
		t.Fatalf("without the header the chain lets the request through: %d", code)
	}
}

func TestCustomWatchers(t *testing.T) {
	if !watchers.PaymentsRegistered() {
		t.Fatal("the payments watcher was not registered")
	}
	if watchers.DisabledRegistered() {
		t.Fatal("a watcher turned off in the config must not be registered")
	}

	_, body := serve(t, http.MethodGet, "/goralens/api/status", "")
	var status struct {
		Watchers map[string]bool `json:"watchers"`
		Custom   []struct {
			Type    string `json:"type"`
			Title   string `json:"title"`
			Group   string `json:"group"`
			Columns []struct {
				Label string `json:"label"`
				Field string `json:"field"`
			} `json:"columns"`
		} `json:"custom"`
	}
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatal(err)
	}
	if len(status.Custom) != 1 || status.Custom[0].Type != "payment" || status.Custom[0].Title != "Payments" || status.Custom[0].Group != "Billing" || len(status.Custom[0].Columns) != 3 {
		t.Fatalf("custom watchers in status: %+v", status.Custom)
	}
	if status.Watchers["redis"] {
		t.Fatal(`"redis": false must turn the built-in watcher off`)
	}

	serve(t, http.MethodPost, "/checkout", "")
	goralens.Instance().Flush(time.Second)
	payments := entries(t, "type=payment")
	if len(payments) == 0 || payments[0].Content["order"] != "A-1001" || payments[0].Content["status"] != "paid" {
		t.Fatalf("payment entry: %+v", payments)
	}
}

func TestSensitiveDataIsMaskedOnEveryMethod(t *testing.T) {
	body := `{"user":{"name":"ali","password":"hunter2"},"password_confirmation":"hunter2","note":"visible"}`
	if code, _ := serve(t, http.MethodPut, "/profile?token=abc123&page=2", body); code != http.StatusOK {
		t.Fatalf("put: %d", code)
	}
	var found *entry
	for _, e := range entries(t, "type=request") {
		if e.Content["method"] == "PUT" {
			found = &e
			break
		}
	}
	if found == nil {
		t.Fatal("PUT request not recorded")
	}
	raw, _ := json.Marshal(found.Content)
	for _, secret := range []string{"hunter2", "abc123", "should-be-hidden", "issued-token"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("%q leaked into the entry: %s", secret, raw)
		}
	}
	payload := found.Content["payload"].(map[string]any)
	if payload["note"] != "visible" || payload["user"].(map[string]any)["name"] != "ali" {
		t.Fatalf("non-sensitive fields must stay readable: %v", payload)
	}
	if !strings.Contains(found.Content["uri"].(string), "page=2") {
		t.Fatalf("uri lost its other parameters: %v", found.Content["uri"])
	}
}

func adminOnly() func() {
	goralens.ResolveUser(func(ctx contractshttp.Context) *goralens.User {
		if ctx.Request().Header("Authorization") == "Bearer good-token" {
			return &goralens.User{ID: 1, Email: "admin@example.com"}
		}
		return nil
	})
	goralens.Auth(func(ctx contractshttp.Context) bool {
		u := goralens.CurrentUser(ctx)
		return u != nil && u.Email == "admin@example.com"
	})
	return func() {
		goralens.Auth(nil)
		goralens.ResolveUser(nil)
	}
}

func startSession(token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/goralens/session", strings.NewReader("token="+token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	return rec
}

func withCookie(path, cookie string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	return rec.Code
}

func TestSessionHandoffForBearerTokenApps(t *testing.T) {
	defer adminOnly()()

	if code := withCookie("/goralens/api/status", ""); code != http.StatusForbidden {
		t.Fatalf("no session: %d", code)
	}
	if rec := startSession("wrong-token"); rec.Code != http.StatusForbidden || rec.Header().Get("Set-Cookie") != "" {
		t.Fatalf("a wrong token must not start a session: %d %q", rec.Code, rec.Header().Get("Set-Cookie"))
	}

	rec := startSession("good-token")
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/goralens" {
		t.Fatalf("handoff: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "goralens_session=") || !strings.Contains(setCookie, "HttpOnly") || !strings.Contains(setCookie, "Path=/goralens") {
		t.Fatalf("session cookie: %q", setCookie)
	}
	cookie := strings.SplitN(setCookie, ";", 2)[0]
	if code := withCookie("/goralens/api/status", cookie); code != http.StatusOK {
		t.Fatalf("with the session cookie: %d", code)
	}
	if code := withCookie("/goralens/api/status", cookie+"x"); code != http.StatusForbidden {
		t.Fatalf("a tampered cookie must be refused: %d", code)
	}
}

func TestTokenStorageModeServesTheShellOnly(t *testing.T) {
	defer adminOnly()()

	req := httptest.NewRequest(http.MethodGet, "/goralens", nil)
	rec := httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `data-token-key="app_token"`) {
		t.Fatalf("shell: %d", rec.Code)
	}
	if code := withCookie("/goralens/api/entries", ""); code != http.StatusForbidden {
		t.Fatalf("data without a token: %d", code)
	}
	req = httptest.NewRequest(http.MethodGet, "/goralens/api/entries", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec = httptest.NewRecorder()
	facades.Route().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("data with the bearer token: %d", rec.Code)
	}
}
