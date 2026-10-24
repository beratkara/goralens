package goralens

import (
	"net/http"
	"strings"
	"testing"
)

func TestRedaction(t *testing.T) {
	headers := redactHeaders(http.Header{"Authorization": {"Bearer x"}, "Accept": {"a", "b"}}, []string{"authorization"})
	if headers["authorization"] != hiddenValue || headers["accept"] != "a, b" {
		t.Fatalf("headers: %v", headers)
	}
	payload := redactMap(map[string]any{
		"password": "secret",
		"user":     map[string]any{"Password": "x", "name": "n"},
		"list":     []any{map[string]any{"token": "t"}},
	}, []string{"password", "token"}).(map[string]any)
	if payload["password"] != hiddenValue || payload["user"].(map[string]any)["Password"] != hiddenValue ||
		payload["list"].([]any)[0].(map[string]any)["token"] != hiddenValue || payload["user"].(map[string]any)["name"] != "n" {
		t.Fatalf("payload: %v", payload)
	}
}

func TestResponseContent(t *testing.T) {
	if got := responseContent([]byte(`{"token":"t","ok":true}`), "application/json", 64, []string{"token"}); got.(map[string]any)["token"] != hiddenValue {
		t.Fatalf("json: %v", got)
	}
	if got := responseContent([]byte("<html></html>"), "text/html", 64, nil); got != "HTML Response" {
		t.Fatalf("html: %v", got)
	}
	if got := responseContent([]byte(strings.Repeat("a", 2048)), "text/plain", 1, nil); !strings.HasPrefix(got.(string), "Purged") {
		t.Fatalf("limit: %v", got)
	}
	if got := responseContent(nil, "", 64, nil); got != "Empty Response" {
		t.Fatalf("empty: %v", got)
	}
}

func TestMatchesAny(t *testing.T) {
	if !matchesAny("/health/live", []string{"/health*"}) || matchesAny("/api", []string{"/health*", "/ap"}) {
		t.Fatal("pattern matching is wrong")
	}
}

func TestRedactURIAndForm(t *testing.T) {
	hidden := []string{"token", "password"}
	if got := redactURI("/reset?token=abc&page=2", hidden); got != "/reset?page=2&token=********" {
		t.Fatalf("redactURI query: %s", got)
	}
	if got := redactURI("https://user:secret@example.com/x", hidden); got != "https://user:********@example.com/x" {
		t.Fatalf("redactURI userinfo: %s", got)
	}
	form := redactForm("password=hunter2&name=ali&tags=a&tags=b", hidden).(map[string]any)
	if form["password"] != hiddenValue || form["name"] != "ali" || len(form["tags"].([]string)) != 2 {
		t.Fatalf("redactForm: %#v", form)
	}
}
