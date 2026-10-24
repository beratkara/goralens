package goralens

import (
	"context"
	"embed"
	"html"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
)

//go:embed ui
var uiFiles embed.FS

type Dashboard struct {
	recorder *Recorder
}

func NewDashboard(r *Recorder) *Dashboard {
	return &Dashboard{recorder: r}
}

func (d *Dashboard) Serve(ctx http.Context, path string) {
	res := ctx.Response()
	res.Header("Cache-Control", "no-store")
	res.Header("X-Robots-Tag", "noindex, nofollow")
	method := ctx.Request().Method()
	if path == "/session" && method == nethttp.MethodPost {
		d.startSession(ctx)
		return
	}
	d.restoreSession(ctx)
	shell := path == "" || path == "/" || path == "/app.js" || path == "/app.css"
	if !(shell && d.recorder.config.TokenStorageKey != "") && !runMiddleware(ctx, d.recorder.config.Middleware) {
		return
	}

	if method != nethttp.MethodGet && ctx.Request().Header("X-Goralens") != "1" {
		_ = res.Json(nethttp.StatusForbidden, map[string]any{"message": "Missing X-Goralens header."}).Abort()
		return
	}

	switch {
	case path == "" || path == "/":
		d.file(ctx, "ui/index.html", "text/html; charset=utf-8")
	case path == "/app.js":
		d.file(ctx, "ui/app.js", "text/javascript; charset=utf-8")
	case path == "/app.css":
		d.file(ctx, "ui/app.css", "text/css; charset=utf-8")
	case path == "/api/status" && method == nethttp.MethodGet:
		d.status(ctx)
	case path == "/api/entries" && method == nethttp.MethodGet:
		d.list(ctx)
	case path == "/api/entries" && method == nethttp.MethodDelete:
		d.clear(ctx)
	case strings.HasPrefix(path, "/api/entries/") && method == nethttp.MethodGet:
		d.show(ctx, strings.TrimPrefix(path, "/api/entries/"))
	case path == "/api/session" && method == nethttp.MethodDelete:
		d.endSession(ctx)
	case path == "/api/recording" && method == nethttp.MethodPost:
		d.toggleRecording(ctx)
	case path == "/api/monitored-tags" && method == nethttp.MethodGet:
		d.monitored(ctx)
	case path == "/api/monitored-tags" && (method == nethttp.MethodPost || method == nethttp.MethodDelete):
		d.updateMonitored(ctx, method == nethttp.MethodPost)
	default:
		_ = res.Json(nethttp.StatusNotFound, map[string]any{"message": "Not found."}).Abort()
	}
}

func (d *Dashboard) file(ctx http.Context, name, contentType string) {
	data, err := uiFiles.ReadFile(name)
	if err != nil {
		_ = ctx.Response().Json(nethttp.StatusNotFound, map[string]any{"message": "Not found."}).Abort()
		return
	}
	if name == "ui/index.html" {
		page := strings.ReplaceAll(string(data), "{{BASE}}", "/"+d.recorder.config.Path)
		page = strings.ReplaceAll(page, "{{TOKEN_KEY}}", html.EscapeString(d.recorder.config.TokenStorageKey))
		data = []byte(page)
	}
	_ = ctx.Response().Data(nethttp.StatusOK, contentType, data).Abort()
}

func (d *Dashboard) storageCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithValue(context.Background(), skipKey{}, true), 15*time.Second)
}

func (d *Dashboard) fail(ctx http.Context, err error) {
	_ = ctx.Response().Json(nethttp.StatusInternalServerError, map[string]any{"message": err.Error()}).Abort()
}

func (d *Dashboard) status(ctx http.Context) {
	c, cancel := d.storageCtx()
	defer cancel()
	counts, err := d.recorder.storage.Counts(c)
	if err != nil {
		d.fail(ctx, err)
		return
	}
	watchers := map[EntryType]bool{}
	for _, t := range EntryTypes {
		watchers[t] = d.recorder.config.watching(t)
	}
	custom := customWatchers()
	for _, w := range custom {
		watchers[w.Type] = d.recorder.config.watching(w.Type)
	}
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{
		"recording": d.recorder.Recording(),
		"counts":    counts,
		"watchers":  watchers,
		"types":     EntryTypes,
		"custom":    custom,
		"dropped":   d.recorder.Dropped(),
	}).Abort()
}

func (d *Dashboard) list(ctx http.Context) {
	req := ctx.Request()
	before, _ := strconv.ParseInt(req.Query("before"), 10, 64)
	take, _ := strconv.Atoi(req.Query("take"))
	q := Query{
		Type:       EntryType(req.Query("type")),
		Tag:        strings.TrimSpace(req.Query("tag")),
		FamilyHash: req.Query("family_hash"),
		Before:     before,
		Limit:      take,
	}

	q.IndexOnly = q.Type == TypeException && q.FamilyHash == "" && q.Tag == ""
	c, cancel := d.storageCtx()
	defer cancel()
	entries, err := d.recorder.storage.Get(c, q)
	if err != nil {
		d.fail(ctx, err)
		return
	}
	if entries == nil {
		entries = []*Entry{}
	}
	if q.IndexOnly {
		for _, e := range entries {
			if e.FamilyHash != "" {
				n, _ := d.recorder.storage.Occurrences(c, e.FamilyHash)
				e.Content["occurrences"] = n
			}
		}
	}
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"entries": entries}).Abort()
}

func (d *Dashboard) show(ctx http.Context, id string) {
	c, cancel := d.storageCtx()
	defer cancel()
	entry, err := d.recorder.storage.Find(c, id)
	if err != nil {
		d.fail(ctx, err)
		return
	}
	if entry == nil {
		_ = ctx.Response().Json(nethttp.StatusNotFound, map[string]any{"message": "Entry not found."}).Abort()
		return
	}
	batch, err := d.recorder.storage.Get(c, Query{BatchID: entry.BatchID, Limit: 200})
	if err != nil {
		d.fail(ctx, err)
		return
	}
	related := make([]*Entry, 0, len(batch))
	for _, e := range batch {
		if e.UUID != entry.UUID {
			related = append(related, e)
		}
	}
	result := map[string]any{"entry": entry, "batch": related}
	if entry.FamilyHash != "" {
		n, _ := d.recorder.storage.Occurrences(c, entry.FamilyHash)
		result["occurrences"] = n
	}
	_ = ctx.Response().Json(nethttp.StatusOK, result).Abort()
}

func (d *Dashboard) clear(ctx http.Context) {
	c, cancel := d.storageCtx()
	defer cancel()
	if err := d.recorder.storage.Clear(c); err != nil {
		d.fail(ctx, err)
		return
	}
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"message": "Entries cleared."}).Abort()
}

func (d *Dashboard) toggleRecording(ctx http.Context) {
	d.recorder.SetRecording(ctx.Request().InputBool("recording"))
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"recording": d.recorder.Recording()}).Abort()
}

func (d *Dashboard) monitored(ctx http.Context) {
	c, cancel := d.storageCtx()
	defer cancel()
	tags, err := d.recorder.storage.MonitoredTags(c)
	if err != nil {
		d.fail(ctx, err)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"tags": tags}).Abort()
}

func (d *Dashboard) updateMonitored(ctx http.Context, add bool) {
	tags := ctx.Request().InputArray("tags")
	if len(tags) == 0 {
		_ = ctx.Response().Json(nethttp.StatusUnprocessableEntity, map[string]any{"message": "Send a non-empty tags list."}).Abort()
		return
	}
	c, cancel := d.storageCtx()
	defer cancel()
	var err error
	if add {
		err = d.recorder.storage.Monitor(c, tags)
	} else {
		err = d.recorder.storage.Unmonitor(c, tags)
	}
	if err != nil {
		d.fail(ctx, err)
		return
	}
	current, err := d.recorder.storage.MonitoredTags(c)
	if err != nil {
		d.fail(ctx, err)
		return
	}
	d.recorder.setMonitored(current)
	_ = ctx.Response().Json(nethttp.StatusOK, map[string]any{"tags": current}).Abort()
}
