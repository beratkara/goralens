package goralens

import (
	"fmt"
	nethttp "net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
)

var hostname, _ = os.Hostname()

type Middleware struct {
	recorder  *Recorder
	dashboard *Dashboard
}

func NewMiddleware(r *Recorder) *Middleware {
	return &Middleware{recorder: r, dashboard: NewDashboard(r)}
}

func (m *Middleware) Signature() string { return "goralens" }

func (m *Middleware) Handle(ctx http.Context) {
	r := m.recorder
	path := "/" + strings.Trim(ctx.Request().Path(), "/")
	prefix := "/" + r.config.Path
	if path == prefix || strings.HasPrefix(path, prefix+"/") {
		defer r.Suppress()()
		m.dashboard.Serve(ctx, strings.TrimPrefix(path, prefix))
		return
	}
	if !r.Recording() || matchesAny(path, r.config.IgnorePaths) {
		ctx.Request().Next()
		return
	}

	start := time.Now()
	batch, end := r.StartBatch()
	ctx.WithValue(batchKey{}, batch)
	ctx.WithContext(WithBatch(ctx.Request().Origin().Context(), batch))
	payload := redactMap(ctx.Request().All(), r.config.HiddenRequestParameters)

	defer func() {
		recovered := recover()
		if recovered != nil {
			r.recordException(ctx, recovered, panicTrace(), map[string]any{"panic": true})
		}
		m.recordRequest(ctx, batch, path, payload, start, recovered != nil)
		end()
		if recovered != nil {
			panic(recovered)
		}
	}()
	ctx.Request().Next()
}

func (m *Middleware) recordRequest(ctx http.Context, batch *Batch, path string, payload any, start time.Time, panicked bool) {
	r := m.recorder
	var user *User
	if hook := currentHooks().user; hook != nil {
		if user = hook(ctx); user != nil && user.ID != nil {
			batch.Tag(fmt.Sprintf("Auth:%v", user.ID))
		}
	}
	if !r.config.watching(TypeRequest) {
		return
	}
	req := ctx.Request()
	status, body, responseHeaders := responseOf(ctx)
	if panicked && status < 500 {
		status = 500
	}
	if slices.Contains(r.config.IgnoreStatusCodes, status) {
		return
	}
	uri := path
	if q := req.Origin().URL.RawQuery; q != "" {
		uri += "?" + q
	}
	info := req.Info()
	content := map[string]any{
		"ip":                req.Ip(),
		"uri":               redactURI(uri, r.config.HiddenRequestParameters),
		"method":            req.Method(),
		"route":             info.Path,
		"route_name":        info.Name,
		"controller_action": info.Handler,
		"headers":           redactHeaders(req.Headers(), r.config.HiddenRequestHeaders),
		"payload":           payload,
		"response_status":   status,
		"response_headers":  redactHeaders(responseHeaders, r.config.HiddenResponseHeaders),
		"response":          responseContent(body, responseHeaders.Get("Content-Type"), r.config.ResponseSizeLimit, r.config.HiddenResponseParameters),
		"duration":          milliseconds(time.Since(start)),
		"hostname":          hostname,
	}
	if user != nil {
		content["user"] = user
	}
	e := newEntry(TypeRequest, content)
	e.CreatedAt = start
	r.Record(ctx, e)
}

func responseOf(ctx http.Context) (int, []byte, nethttp.Header) {
	res := ctx.Response()
	if origin := res.Origin(); origin != nil {
		var body []byte
		if b := origin.Body(); b != nil {
			body = b.Bytes()
		}
		return origin.Status(), body, origin.Header()
	}
	w := res.Writer()
	status := 200
	if s, ok := w.(interface{ Status() int }); ok {
		status = s.Status()
	}
	return status, nil, w.Header()
}
