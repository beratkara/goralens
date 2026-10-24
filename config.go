package goralens

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/spf13/cast"
)

type Config struct {
	Enabled bool

	Environment string

	Path string

	Driver string

	Connection string

	QueueSize int

	FlushInterval time.Duration

	IgnorePaths    []string
	IgnoreCommands []string

	HiddenRequestHeaders     []string
	HiddenResponseHeaders    []string
	HiddenRequestParameters  []string
	HiddenResponseParameters []string

	ResponseSizeLimit int

	Watchers map[EntryType]bool

	Middleware      []DashboardMiddleware
	Key             string
	TokenStorageKey string
	SessionLifetime time.Duration
	AllowedUsers    map[string][]string
	CustomWatchers  []Watcher

	IgnoreStatusCodes  []int
	SlowQuery          time.Duration
	QueryConnections   []string
	LogLevel           string
	ExceptionsFromLogs bool
	IgnoreAbilities    []string
	HiddenCacheKeys    []string
}

func defaultConfig() *Config {
	c := &Config{
		Enabled:                  true,
		Path:                     "goralens",
		Driver:                   "database",
		QueueSize:                1000,
		IgnorePaths:              []string{"/favicon.ico"},
		FlushInterval:            500 * time.Millisecond,
		HiddenRequestHeaders:     []string{"authorization", "cookie", "proxy-authorization", "x-csrf-token", "x-xsrf-token", "x-api-key"},
		HiddenResponseHeaders:    []string{"set-cookie"},
		HiddenRequestParameters:  []string{"password", "password_confirmation", "current_password", "new_password", "token", "access_token", "refresh_token", "secret", "_token"},
		HiddenResponseParameters: []string{"token", "access_token", "refresh_token"},
		ResponseSizeLimit:        64,
		Watchers:                 map[EntryType]bool{},
		Middleware:               []DashboardMiddleware{Authorize()},
		SessionLifetime:          2 * time.Hour,
		SlowQuery:                100 * time.Millisecond,
		LogLevel:                 "debug",
		ExceptionsFromLogs:       true,
	}
	for _, t := range EntryTypes {
		c.Watchers[t] = true
	}
	return c
}

func loadConfig(cfg config.Config) *Config {
	c := defaultConfig()
	if cfg == nil {
		return c
	}
	c.Enabled = cast.ToBool(setting(cfg, "goralens.enabled", "GORALENS_ENABLED", c.Enabled))
	c.Environment = cfg.GetString("app.env")
	c.Path = strings.Trim(cast.ToString(setting(cfg, "goralens.path", "GORALENS_PATH", c.Path)), "/")
	if c.Path == "" {
		c.Path = "goralens"
	}
	c.Driver = cast.ToString(setting(cfg, "goralens.storage.driver", "GORALENS_DRIVER", c.Driver))
	c.Connection = cast.ToString(setting(cfg, "goralens.storage.connection", "GORALENS_DB_CONNECTION", c.Connection))
	c.QueueSize = cfg.GetInt("goralens.queue_size", c.QueueSize)
	if ms := cfg.GetInt("goralens.flush_interval_ms", 0); ms > 0 {
		c.FlushInterval = time.Duration(ms) * time.Millisecond
	}
	c.IgnorePaths = stringSlice(cfg.Get("goralens.ignore_paths"), c.IgnorePaths)
	c.IgnoreCommands = stringSlice(cfg.Get("goralens.ignore_commands"), c.IgnoreCommands)
	c.HiddenRequestHeaders = stringSlice(cfg.Get("goralens.hidden.request_headers"), c.HiddenRequestHeaders)
	c.HiddenResponseHeaders = stringSlice(cfg.Get("goralens.hidden.response_headers"), c.HiddenResponseHeaders)
	c.HiddenRequestParameters = stringSlice(cfg.Get("goralens.hidden.request_parameters"), c.HiddenRequestParameters)
	c.HiddenResponseParameters = stringSlice(cfg.Get("goralens.hidden.response_parameters"), c.HiddenResponseParameters)
	if list, ok := cfg.Get("goralens.middleware").([]DashboardMiddleware); ok {
		c.Middleware = list
	}
	c.Key = cfg.GetString("app.key")
	c.TokenStorageKey = strings.TrimSpace(cast.ToString(setting(cfg, "goralens.token_storage_key", "GORALENS_TOKEN_STORAGE_KEY", "")))
	if minutes := cast.ToInt(setting(cfg, "goralens.session_minutes", "GORALENS_SESSION_MINUTES", 0)); minutes > 0 {
		c.SessionLifetime = time.Duration(minutes) * time.Minute
	}
	c.AllowedUsers = allowedUsers(cfg.Get("goralens.allowed_users"))
	if emails := valueList(setting(cfg, "goralens.allowed_emails", "GORALENS_ALLOWED_EMAILS", nil)); len(emails) > 0 {
		c.AllowedUsers["email"] = append(c.AllowedUsers["email"], emails...)
	}
	if list, ok := cfg.Get("goralens.watchers").([]Watcher); ok {
		applyWatchers(c, list)
	}
	return c
}

func setting(cfg config.Config, key, env string, fallback any) any {
	if v := cfg.Get(key); v != nil && v != "" {
		return v
	}
	if v := cfg.Env(env); v != nil && v != "" {
		return v
	}
	return fallback
}

func allowedUsers(v any) map[string][]string {
	out := map[string][]string{}
	m, ok := v.(map[string]any)
	if !ok {
		if typed, isTyped := v.(map[string][]string); isTyped {
			for k, vals := range typed {
				out[k] = valueList(vals)
			}
		}
		return out
	}
	for field, vals := range m {
		if list := valueList(vals); len(list) > 0 {
			out[field] = list
		}
	}
	return out
}

func valueList(v any) []string {
	var raw []string
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		raw = strings.Split(x, ",")
	default:
		raw = cast.ToStringSlice(x)
	}
	out := make([]string, 0, len(raw))
	for _, e := range raw {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}
	return out
}

func stringSlice(v any, fallback []string) []string {
	if v == nil {
		return fallback
	}
	return cast.ToStringSlice(v)
}

func (c *Config) watching(t EntryType) bool {
	return c.Enabled && c.Watchers[t]
}

func matchesAny(value string, patterns []string) bool {
	for _, p := range patterns {
		if strings.HasSuffix(p, "*") {
			if strings.HasPrefix(value, strings.TrimSuffix(p, "*")) {
				return true
			}
		} else if value == p {
			return true
		}
	}
	return false
}
