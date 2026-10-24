package goralens

import (
	"testing"
	"time"
)

func TestSessionCookieRoundTrip(t *testing.T) {
	now := time.Now()
	value, err := signSession("app-key", &User{ID: 7, Email: "admin@example.com"}, time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	user := verifySession("app-key", value, now)
	if user == nil || user.Email != "admin@example.com" {
		t.Fatalf("valid cookie rejected: %+v", user)
	}
	if verifySession("other-key", value, now) != nil {
		t.Fatal("cookie signed with another key accepted")
	}
	if verifySession("app-key", value, now.Add(2*time.Hour)) != nil {
		t.Fatal("expired cookie accepted")
	}
	if verifySession("app-key", "x"+value, now) != nil {
		t.Fatal("tampered cookie accepted")
	}
	if verifySession("", value, now) != nil {
		t.Fatal("cookie accepted without a key")
	}
}
