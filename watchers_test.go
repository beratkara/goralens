package goralens

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/auth/access"
	frameworkcache "github.com/goravel/framework/cache"
	contractsaccess "github.com/goravel/framework/contracts/auth/access"
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsevent "github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/log"
	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksevent "github.com/goravel/framework/mocks/event"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	"github.com/goravel/framework/schedule"
	"github.com/stretchr/testify/mock"
)

type memoryCache struct{ *frameworkcache.Memory }

func (c memoryCache) Store(string) contractscache.Driver { return c.Memory }

func TestCacheWatcher(t *testing.T) {
	r, s := newTestRecorder(t, func(c *Config) { c.HiddenCacheKeys = []string{"secret*"} })
	cfg := mocksconfig.NewConfig(t)
	cfg.EXPECT().GetString("cache.prefix").Return("")
	mem, _ := frameworkcache.NewMemory(cfg)
	c := WrapCache(memoryCache{mem})

	c.Get("missing")
	_ = c.Put("user", "ada", time.Minute)
	c.Get("user")
	_ = c.Put("secret-key", "s", 0)
	c.Forget("user")

	var kinds []string
	for _, e := range stored(t, r, s, TypeCache) {
		kinds = append(kinds, e.Content["type"].(string)+":"+e.Content["key"].(string))
		if e.Content["key"] == "secret-key" && e.Content["value"] != hiddenValue {
			t.Fatalf("hidden key value stored: %v", e.Content["value"])
		}
	}
	want := "forget:user set:secret-key hit:user set:user missed:missing"
	if got := strings.Join(kinds, " "); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

type orderShipped struct{}

func (orderShipped) Handle(args []contractsevent.Arg) ([]contractsevent.Arg, error) { return args, nil }

func TestEventWatcher(t *testing.T) {
	r, s := newTestRecorder(t)
	ev := orderShipped{}
	task := mocksevent.NewTask(t)
	task.EXPECT().Dispatch().Return(nil)
	listener := mocksevent.NewListener(t)
	listener.EXPECT().Signature().Return("send_shipment_mail")
	listener.EXPECT().Queue().Return(contractsevent.Queue{Enable: true})
	inner := mocksevent.NewInstance(t)
	inner.EXPECT().Job(ev, mock.Anything).Return(task)
	inner.EXPECT().GetEvents().Return(map[contractsevent.Event][]contractsevent.Listener{ev: {listener}})

	if err := WrapEvent(inner).Job(ev, []contractsevent.Arg{{Type: "int", Value: 42}}).Dispatch(); err != nil {
		t.Fatal(err)
	}
	e := one(t, r, s, TypeEvent)
	if e.Content["name"] != "goralens.orderShipped" {
		t.Fatalf("name: %v", e.Content["name"])
	}
	listeners := e.Content["listeners"].([]map[string]any)
	if len(listeners) != 1 || listeners[0]["name"] != "send_shipment_mail" || listeners[0]["queued"] != true {
		t.Fatalf("listeners: %v", listeners)
	}
}

type sendInvoice struct{ fail bool }

func (sendInvoice) Signature() string { return "send_invoice" }
func (j sendInvoice) Handle(args ...any) error {
	Dump("inside job")
	if j.fail {
		return errors.New("smtp down")
	}
	return nil
}

func TestQueueWatcher(t *testing.T) {
	r, s := newTestRecorder(t)
	pending := mocksqueue.NewPendingJob(t)
	pending.EXPECT().OnQueue("mail").Return(pending)
	pending.EXPECT().Dispatch().Return(nil)
	inner := mocksqueue.NewQueue(t)
	var registered []contractsqueue.Job
	inner.EXPECT().Register(mock.Anything).Run(func(jobs []contractsqueue.Job) { registered = jobs })
	inner.EXPECT().Job(mock.Anything, mock.Anything).Return(pending)

	q := WrapQueue(inner)
	q.Register([]contractsqueue.Job{sendInvoice{fail: true}})
	if err := q.Job(sendInvoice{}, []contractsqueue.Arg{{Type: "string", Value: "inv-1"}}).OnQueue("mail").Dispatch(); err != nil {
		t.Fatal(err)
	}

	if err := registered[0].Handle("inv-1"); err == nil {
		t.Fatal("job error was swallowed")
	}

	statuses := map[string]*Entry{}
	for _, e := range stored(t, r, s, TypeJob) {
		statuses[e.Content["status"].(string)] = e
	}
	if statuses["queued"] == nil || statuses["queued"].Content["queue"] != "mail" {
		t.Fatalf("queued entry: %+v", statuses["queued"])
	}
	failed := statuses["failed"]
	if failed == nil || failed.Content["error"] != "smtp down" {
		t.Fatalf("failed entry: %+v", failed)
	}

	dump := one(t, r, s, TypeDump)
	if dump.BatchID != failed.BatchID {
		t.Fatal("job entries are not batched together")
	}
}

func TestMailWatcher(t *testing.T) {
	r, s := newTestRecorder(t)
	inner := mocksmail.NewMail(t)
	inner.EXPECT().To([]string{"ada@example.com"}).Return(inner)
	inner.EXPECT().Subject("Welcome").Return(inner)
	inner.EXPECT().Content(mock.Anything).Return(inner)
	inner.EXPECT().Send().Return(nil)

	err := WrapMail(inner).To([]string{"ada@example.com"}).Subject("Welcome").
		Content(contractsmail.Content{Html: "<h1>Hi</h1>"}).Send()
	if err != nil {
		t.Fatal(err)
	}
	e := one(t, r, s, TypeMail)
	if e.Content["subject"] != "Welcome" || e.Content["html"] != "<h1>Hi</h1>" || !e.HasTag("mail:ada@example.com") {
		t.Fatalf("mail entry: %+v", e)
	}
}

func TestGateWatcher(t *testing.T) {
	r, s := newTestRecorder(t, func(c *Config) { c.IgnoreAbilities = []string{"ignored"} })
	gate := access.NewGate(context.Background())
	gate.Define("edit-post", func(ctx context.Context, args map[string]any) contractsaccess.Response {
		return access.NewAllowResponse()
	})
	WatchGate(gate)
	if !gate.Allows("edit-post", map[string]any{"post": 1}) {
		t.Fatal("watcher changed the result")
	}
	gate.Denies("delete-post", nil)
	gate.Allows("ignored", nil)

	results := map[string]string{}
	for _, e := range stored(t, r, s, TypeGate) {
		results[e.Content["ability"].(string)] = e.Content["result"].(string)
	}
	if len(results) != 2 || results["edit-post"] != "allowed" || results["delete-post"] != "denied" {
		t.Fatalf("results: %v", results)
	}
}

func TestHttpClientWatcher(t *testing.T) {
	r, s := newTestRecorder(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":1,"access_token":"abc"}`)
	}))
	defer server.Close()

	client := &http.Client{Transport: Transport(nil)}
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/users", strings.NewReader(`{"password":"p"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer x")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if string(body) != `{"id":1,"access_token":"abc"}` {
		t.Fatalf("caller got a changed body: %s", body)
	}
	e := one(t, r, s, TypeClientRequest)
	if e.Content["response_status"] != 200 || e.Content["headers"].(map[string]string)["authorization"] != hiddenValue {
		t.Fatalf("entry: %+v", e.Content)
	}
	if e.Content["payload"].(map[string]any)["password"] != hiddenValue || e.Content["response"].(map[string]any)["access_token"] != hiddenValue {
		t.Fatalf("secrets stored: %+v", e.Content)
	}
}

func TestArtisanWatcher(t *testing.T) {
	r, s := newTestRecorder(t, func(c *Config) { c.IgnoreCommands = []string{"list"} })
	inner := mocksconsole.NewArtisan(t)
	inner.EXPECT().Call("report:send --daily").RunAndReturn(func(string) error {
		Dump("inside command")
		return errors.New("no recipients")
	})
	inner.EXPECT().Call("list").Return(nil)

	a := WrapArtisan(inner)
	if err := a.Call("report:send --daily"); err == nil {
		t.Fatal("error swallowed")
	}
	_ = a.Call("list")
	e := one(t, r, s, TypeCommand)
	if e.Content["command"] != "report:send" || e.Content["exit_code"] != 1 {
		t.Fatalf("command entry: %+v", e.Content)
	}
	if one(t, r, s, TypeDump).BatchID != e.BatchID {
		t.Fatal("command entries are not batched together")
	}
}

func TestScheduleWatcher(t *testing.T) {
	r, s := newTestRecorder(t)
	var registered []contractsschedule.Event
	inner := mocksschedule.NewSchedule(t)
	inner.EXPECT().Register(mock.Anything).Run(func(events []contractsschedule.Event) { registered = events })
	artisan := mocksconsole.NewArtisan(t)
	artisan.EXPECT().Call("cleanup").Return(nil)

	ran := false
	WrapSchedule(inner, artisan).Register([]contractsschedule.Event{
		schedule.NewCallbackEvent(func() { ran = true }).EveryMinute().Name("heartbeat"),
		schedule.NewCommandEvent("cleanup").Daily(),
	})
	for _, e := range registered {
		if e.GetCommand() != "" {
			t.Fatal("command events must run through the callback")
		}
		e.GetCallback()()
	}
	if !ran {
		t.Fatal("callback did not run")
	}
	names := map[string]string{}
	for _, e := range stored(t, r, s, TypeSchedule) {
		names[e.Content["name"].(string)] = e.Content["status"].(string)
	}
	if names["heartbeat"] != "processed" || names["cleanup"] != "processed" {
		t.Fatalf("schedule entries: %v", names)
	}
}

func TestLogWatcher(t *testing.T) {
	r, s := newTestRecorder(t, func(c *Config) { c.LogLevel = "info" })
	handler, _ := LogDriver().Handle("goralens")
	if handler.Enabled(log.LevelDebug) {
		t.Fatal("debug enabled below the configured level")
	}
	entry := mockslog.NewEntry(t)
	entry.EXPECT().Context().Return(context.Background())
	entry.EXPECT().With().Return(map[string]any{"order": 5})
	entry.EXPECT().Data().Return(log.Data{})
	entry.EXPECT().Level().Return(log.LevelError)
	entry.EXPECT().Message().Return("payment failed")
	entry.EXPECT().Code().Return("")
	entry.EXPECT().Domain().Return("")
	entry.EXPECT().Hint().Return("")
	entry.EXPECT().User().Return(nil)
	entry.EXPECT().Request().Return(nil)
	entry.EXPECT().Time().Return(time.Now())
	entry.EXPECT().Tags().Return([]string{"billing"})
	entry.EXPECT().Trace().Return(nil)
	if err := handler.Handle(entry); err != nil {
		t.Fatal(err)
	}
	logEntry := one(t, r, s, TypeLog)
	if logEntry.Content["message"] != "payment failed" || !logEntry.HasTag("billing") {
		t.Fatalf("log entry: %+v", logEntry)
	}
	exception := one(t, r, s, TypeException)

	if exception.Content["class"] != "Log error" || exception.Content["message"] != "payment failed" {
		t.Fatalf("exception from log: %+v", exception.Content)
	}
}

func TestDumpRecordsLocation(t *testing.T) {
	r, s := newTestRecorder(t)
	Dump(struct{ Name string }{"ada"})
	e := one(t, r, s, TypeDump)
	if !strings.Contains(e.Content["dump"].(string), `"Name": "ada"`) {
		t.Fatalf("dump: %v", e.Content["dump"])
	}
}
