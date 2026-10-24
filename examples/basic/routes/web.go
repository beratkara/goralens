package routes

import (
	"context"
	"time"

	"github.com/beratkara/goralens"
	"github.com/goravel/framework/auth/access"
	contractsaccess "github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"

	"github.com/beratkara/goralens/examples/basic/app/facades"
	"github.com/beratkara/goralens/examples/basic/app/models"
	"github.com/beratkara/goralens/examples/basic/app/watchers"
)

type UserRegistered struct{}

func (UserRegistered) Handle(args []event.Arg) ([]event.Arg, error) { return args, nil }

type SendWelcome struct{}

func (SendWelcome) Signature() string             { return "send_welcome" }
func (SendWelcome) Queue(args ...any) event.Queue { return event.Queue{} }
func (SendWelcome) Handle(args ...any) error      { return nil }

type WelcomeJob struct{}

func (WelcomeJob) Signature() string { return "welcome_job" }
func (WelcomeJob) Handle(args ...any) error {
	facades.Log().Info("welcome job ran")
	return nil
}

func Web() {
	facades.Gate().Define("view-reports", func(ctx context.Context, args map[string]any) contractsaccess.Response {
		return access.NewDenyResponse("reports are for admins")
	})
	r := facades.Route()

	r.Get("/users", func(ctx http.Context) http.Response {
		var users []models.User
		if err := facades.Orm().Query().Find(&users); err != nil {
			return ctx.Response().Json(500, http.Json{"error": err.Error()})
		}
		count, _ := facades.Cache().Remember("users.count", time.Minute, func() (any, error) { return len(users), nil })
		facades.Log().Info("listing users")
		goralens.Dump(map[string]any{"count": count})
		return ctx.Response().Json(200, http.Json{"data": users})
	})

	r.Post("/users", func(ctx http.Context) http.Response {
		user := models.User{Name: ctx.Request().Input("name"), Email: ctx.Request().Input("email"), Password: ctx.Request().Input("password")}
		if err := facades.Orm().Query().Create(&user); err != nil {
			return ctx.Response().Json(500, http.Json{"error": err.Error()})
		}
		_ = facades.Event().Job(&UserRegistered{}, []event.Arg{{Type: "uint", Value: user.ID}}).Dispatch()
		_ = facades.Queue().Job(&WelcomeJob{}, []queue.Arg{{Type: "uint", Value: user.ID}}).Dispatch()
		return ctx.Response().Json(201, http.Json{"id": user.ID, "token": "secret-token"})
	})

	r.Put("/profile", func(ctx http.Context) http.Response {
		return ctx.Response().Json(http.StatusOK, map[string]any{"saved": true, "token": "issued-token"})
	})

	r.Post("/checkout", func(ctx http.Context) http.Response {
		watchers.RecordPayment(ctx, "A-1001", 49.9, "paid")
		return ctx.Response().Json(http.StatusOK, map[string]any{"paid": true})
	})

	r.Get("/reports", func(ctx http.Context) http.Response {
		allowed := facades.Gate().Allows("view-reports", nil)
		return ctx.Response().Json(200, http.Json{"allowed": allowed})
	})

	r.Get("/error-log", func(ctx http.Context) http.Response {
		facades.Log().Error("payment provider timed out")
		return ctx.Response().Json(502, http.Json{"error": "upstream"})
	})

	r.Get("/panic", func(ctx http.Context) http.Response {
		panic("something broke")
	})
}

func Events() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{&UserRegistered{}: {&SendWelcome{}}}
}

func Jobs() []queue.Job {
	return []queue.Job{&WelcomeJob{}}
}
