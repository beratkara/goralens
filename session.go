package goralens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
)

const (
	sessionCookie  = "goralens_session"
	sessionUserKey = "goralens.user"
)

type sessionPayload struct {
	User      *User `json:"user"`
	ExpiresAt int64 `json:"exp"`
}

func CurrentUser(ctx http.Context) *User {
	if ctx != nil {
		if u, ok := ctx.Value(sessionUserKey).(*User); ok && u != nil {
			return u
		}
	}
	if resolve := currentHooks().user; resolve != nil {
		return resolve(ctx)
	}
	return nil
}

func signSession(key string, user *User, ttl time.Duration, now time.Time) (string, error) {
	raw, err := json.Marshal(sessionPayload{User: user, ExpiresAt: now.Add(ttl).Unix()})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	return body + "." + sessionMAC(key, body), nil
}

func verifySession(key, value string, now time.Time) *User {
	body, mac, ok := strings.Cut(value, ".")
	if !ok || key == "" || !hmac.Equal([]byte(mac), []byte(sessionMAC(key, body))) {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil
	}
	var p sessionPayload
	if json.Unmarshal(raw, &p) != nil || p.User == nil || now.Unix() >= p.ExpiresAt {
		return nil
	}
	return p.User
}

func sessionMAC(key, body string) string {
	m := hmac.New(sha256.New, []byte("goralens-session:"+key))
	m.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (d *Dashboard) restoreSession(ctx http.Context) {
	cookie, err := ctx.Request().Origin().Cookie(sessionCookie)
	if err != nil {
		return
	}
	if user := verifySession(d.recorder.config.Key, cookie.Value, time.Now()); user != nil {
		ctx.WithValue(sessionUserKey, user)
	}
}

func (d *Dashboard) startSession(ctx http.Context) {
	req := ctx.Request().Origin()
	token := strings.TrimSpace(req.FormValue("token"))
	if token == "" {
		_ = ctx.Response().Json(nethttp.StatusUnprocessableEntity, map[string]any{"message": "Missing token."}).Abort()
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	user := CurrentUser(ctx)
	if user == nil {
		_ = ctx.Response().Json(nethttp.StatusForbidden, map[string]any{"message": "Goralens is not available."}).Abort()
		return
	}
	ctx.WithValue(sessionUserKey, user)
	if !runMiddleware(ctx, d.recorder.config.Middleware) {
		return
	}
	value, err := signSession(d.recorder.config.Key, user, d.recorder.config.SessionLifetime, time.Now())
	if err != nil || d.recorder.config.Key == "" {
		_ = ctx.Response().Json(nethttp.StatusInternalServerError, map[string]any{"message": "Goralens sessions need app.key."}).Abort()
		return
	}
	path := "/" + d.recorder.config.Path
	nethttp.SetCookie(ctx.Response().Writer(), &nethttp.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     path,
		MaxAge:   int(d.recorder.config.SessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: nethttp.SameSiteLaxMode,
	})
	_ = ctx.Response().Redirect(nethttp.StatusSeeOther, path).Abort()
}

func (d *Dashboard) endSession(ctx http.Context) {
	nethttp.SetCookie(ctx.Response().Writer(), &nethttp.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/" + d.recorder.config.Path,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
	})
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"message": "Signed out."}).Abort()
}
