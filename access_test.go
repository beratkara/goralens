package goralens

import (
	"testing"

	"github.com/goravel/framework/contracts/http"
)

func resolveEmail(email string) {
	ResolveUser(func(http.Context) *User {
		if email == "" {
			return nil
		}
		return &User{ID: 1, Email: email}
	})
}

func TestAllowEmailsMatchesTheSignedInUser(t *testing.T) {
	newTestRecorder(t)
	gate := AllowEmails("Admin@Example.com", " ops@example.com ")

	resolveEmail("admin@example.com")
	if !gate(nil) {
		t.Fatal("listed email (different case) was refused")
	}
	resolveEmail("ops@example.com")
	if !gate(nil) {
		t.Fatal("listed email with spaces in the config was refused")
	}
	resolveEmail("someone@example.com")
	if gate(nil) {
		t.Fatal("unlisted email was let in")
	}
	resolveEmail("")
	if gate(nil) {
		t.Fatal("guest was let in")
	}
}

func TestAuthorizeOrder(t *testing.T) {
	r, _ := newTestRecorder(t, func(c *Config) { c.Environment = "production" })
	authorize := Authorize()
	if authorize(nil) {
		t.Fatal("production without a callback or allowed emails must stay closed")
	}

	r.config.Environment = "local"
	if !authorize(nil) {
		t.Fatal("local without allowed emails is open")
	}

	r.config.AllowedUsers = map[string][]string{"email": {"admin@example.com"}}
	resolveEmail("intruder@example.com")
	if authorize(nil) {
		t.Fatal("allowed emails apply in local too")
	}
	resolveEmail("admin@example.com")
	if !authorize(nil) {
		t.Fatal("allowed email was refused")
	}

	Auth(func(http.Context) bool { return false })
	if authorize(nil) {
		t.Fatal("an Auth callback overrides the email list")
	}
}

func TestValueListAcceptsCommaSeparatedEnv(t *testing.T) {
	got := valueList(" a@example.com, ,b@example.com ")
	if len(got) != 2 || got[0] != "a@example.com" || got[1] != "b@example.com" {
		t.Fatalf("valueList: %#v", got)
	}
	if valueList(nil) != nil {
		t.Fatal("nil config should give no list")
	}
}

func TestAllowUsersByAnyField(t *testing.T) {
	newTestRecorder(t)
	ResolveUser(func(http.Context) *User {
		return &User{ID: 42, Name: "Berk", Email: "berk@example.com", Attributes: map[string]any{"role": "admin"}}
	})
	cases := []struct {
		field  string
		values []string
		want   bool
	}{
		{"name", []string{"berk"}, true},
		{"id", []string{"42"}, true},
		{"id", []string{"7"}, false},
		{"role", []string{"Admin"}, true},
		{"role", []string{"editor"}, false},
		{"team", []string{"core"}, false},
	}
	for _, c := range cases {
		if got := AllowUsers(c.field, c.values...)(nil); got != c.want {
			t.Errorf("AllowUsers(%q, %v) = %v, want %v", c.field, c.values, got, c.want)
		}
	}
}

func TestAllowedUsersConfig(t *testing.T) {
	got := allowedUsers(map[string]any{"name": []string{"Berk"}, "role": "admin, owner", "empty": []string{}})
	if len(got) != 2 || got["name"][0] != "Berk" || len(got["role"]) != 2 {
		t.Fatalf("allowedUsers: %#v", got)
	}
}
