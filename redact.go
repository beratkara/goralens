package goralens

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"
)

const hiddenValue = "********"

func redactHeaders(h http.Header, hidden []string) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		key := strings.ToLower(k)
		if containsFold(hidden, key) {
			out[key] = hiddenValue
		} else {
			out[key] = strings.Join(v, ", ")
		}
	}
	return out
}

func redactMap(v any, hidden []string) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			if containsFold(hidden, k) {
				out[k] = hiddenValue
			} else {
				out[k] = redactMap(item, hidden)
			}
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = redactMap(item, hidden)
		}
		return out
	default:
		return v
	}
}

func redactURI(raw string, hidden []string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		if _, has := u.User.Password(); has {
			u.User = url.UserPassword(u.User.Username(), hiddenValue)
		}
	}
	if u.RawQuery != "" {
		values := u.Query()
		changed := false
		for k := range values {
			if containsFold(hidden, k) {
				values[k] = []string{hiddenValue}
				changed = true
			}
		}
		if changed {
			u.RawQuery = values.Encode()
		}
	}
	return strings.ReplaceAll(u.String(), url.QueryEscape(hiddenValue), hiddenValue)
}

func redactForm(body string, hidden []string) any {
	values, err := url.ParseQuery(body)
	if err != nil {
		return body
	}
	out := make(map[string]any, len(values))
	for k, v := range values {
		switch {
		case containsFold(hidden, k):
			out[k] = hiddenValue
		case len(v) == 1:
			out[k] = v[0]
		default:
			out[k] = v
		}
	}
	return out
}

func containsFold(list []string, v string) bool {
	for _, item := range list {
		if strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}

func responseContent(body []byte, contentType string, limitKB int, hidden []string) any {
	if len(body) == 0 {
		return "Empty Response"
	}
	if limitKB > 0 && len(body) > limitKB*1024 {
		return "Purged By Goralens: response is larger than the size limit"
	}
	if strings.Contains(contentType, "json") {
		var decoded any
		if json.Unmarshal(body, &decoded) == nil {
			return redactMap(decoded, hidden)
		}
	}
	if strings.Contains(contentType, "text/html") {
		return "HTML Response"
	}
	if !utf8.Valid(body) {
		return fmt.Sprintf("Binary Response (%d bytes)", len(body))
	}
	return string(body)
}

func exportValue(v any) any {
	if v == nil {
		return nil
	}
	if b, err := json.Marshal(v); err == nil {
		var decoded any
		if json.Unmarshal(b, &decoded) == nil {
			if m, ok := decoded.(map[string]any); ok && len(m) == 0 && reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Struct {

				return fmt.Sprintf("%+v", v)
			}
			return decoded
		}
	}
	return fmt.Sprintf("%+v", v)
}

func typeName(v any) string {
	if v == nil {
		return "<nil>"
	}
	return strings.TrimPrefix(fmt.Sprintf("%T", v), "*")
}
