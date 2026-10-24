package goralens

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http/client"
)

func Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if _, ok := base.(*watchedTransport); ok {
		return base
	}
	return &watchedTransport{base: base}
}

type watchedTransport struct {
	base http.RoundTripper
}

const maxClientBody = 64 * 1024

func (t *watchedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := Instance()
	if r == nil || !r.Recording() || !r.config.watching(TypeClientRequest) || skipped(req.Context()) {
		return t.base.RoundTrip(req)
	}
	var payload []byte
	if req.Body != nil && req.GetBody != nil {
		if body, err := req.GetBody(); err == nil {
			payload, _ = io.ReadAll(io.LimitReader(body, maxClientBody+1))
			_ = body.Close()
		}
	}
	start := time.Now()
	res, err := t.base.RoundTrip(req)
	content := map[string]any{
		"method":   req.Method,
		"uri":      redactURI(req.URL.String(), r.config.HiddenRequestParameters),
		"headers":  redactHeaders(req.Header, r.config.HiddenRequestHeaders),
		"payload":  clientBody(payload, req.Header.Get("Content-Type"), r),
		"duration": milliseconds(time.Since(start)),
	}
	if err != nil {
		content["error"] = err.Error()
	} else {
		body, readErr := io.ReadAll(io.LimitReader(res.Body, maxClientBody+1))

		res.Body = &replayBody{Reader: io.MultiReader(bytes.NewReader(body), res.Body), closer: res.Body}
		content["response_status"] = res.StatusCode
		content["response_headers"] = redactHeaders(res.Header, r.config.HiddenResponseHeaders)
		if readErr == nil {
			content["response"] = clientBody(body, res.Header.Get("Content-Type"), r)
		}
	}
	e := newEntry(TypeClientRequest, content)
	e.CreatedAt = start
	e.AddTags(req.URL.Host)
	r.Record(req.Context(), e)
	return res, err
}

func clientBody(body []byte, contentType string, r *Recorder) any {
	if len(body) > maxClientBody {
		return "Purged By Goralens: body is larger than 64 KB"
	}
	if strings.Contains(contentType, "x-www-form-urlencoded") {
		return redactForm(string(body), r.config.HiddenRequestParameters)
	}
	return responseContent(body, contentType, 0, r.config.HiddenRequestParameters)
}

type replayBody struct {
	io.Reader
	closer io.Closer
}

func (b *replayBody) Close() error { return b.closer.Close() }

func WrapHttp(f client.Factory) client.Factory {
	if f == nil {
		return nil
	}
	if _, ok := f.(*watchedHttp); ok {
		return f
	}
	return &watchedHttp{Factory: f}
}

type watchedHttp struct {
	client.Factory
	mu sync.Mutex
}

func (f *watchedHttp) Client(names ...string) client.Request {
	req := f.Factory.Client(names...)

	if hc := req.HttpClient(); hc != nil {
		f.mu.Lock()
		if _, ok := hc.Transport.(*watchedTransport); !ok {
			hc.Transport = Transport(hc.Transport)
		}
		f.mu.Unlock()
	}
	return req
}
