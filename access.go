package goralens

import (
	nethttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"
)

type DashboardMiddleware func(ctx http.Context) bool

func Authorize() DashboardMiddleware {
	return func(ctx http.Context) bool {
		if auth := currentHooks().auth; auth != nil {
			return auth(ctx)
		}
		rec := Instance()
		if rec == nil {
			return false
		}
		if len(rec.config.AllowedUsers) > 0 {
			return userAllowed(ctx, rec.config.AllowedUsers)
		}
		return rec.config.Environment == "local"
	}
}

func AllowUsers(field string, values ...string) DashboardMiddleware {
	allowed := map[string][]string{field: values}
	return func(ctx http.Context) bool {
		return userAllowed(ctx, allowed)
	}
}

func AllowEmails(emails ...string) DashboardMiddleware {
	return AllowUsers("email", emails...)
}

func userAllowed(ctx http.Context, allowed map[string][]string) bool {
	user := CurrentUser(ctx)
	if user == nil {
		return false
	}
	for field, values := range allowed {
		value, ok := user.Field(field)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		for _, v := range values {
			if strings.EqualFold(strings.TrimSpace(v), value) {
				return true
			}
		}
	}
	return false
}

func runMiddleware(ctx http.Context, chain []DashboardMiddleware) bool {
	for _, m := range chain {
		if m == nil {
			continue
		}
		if !m(ctx) {
			if w, ok := ctx.Response().Writer().(interface{ Written() bool }); ok && w.Written() {
				ctx.Request().Abort()
				return false
			}
			_ = ctx.Response().Json(nethttp.StatusForbidden, map[string]any{"message": "Goralens is not available."}).Abort()
			return false
		}
	}
	return true
}
