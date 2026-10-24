package watchers

import (
	"context"
	"sync/atomic"

	"github.com/goravel/framework/contracts/foundation"

	"github.com/beratkara/goralens"
)

const PaymentType goralens.EntryType = "payment"

var paymentsRegistered, disabledRegistered atomic.Bool

type Payments struct{}

func (Payments) Name() string { return string(PaymentType) }

func (Payments) Register(foundation.Application) error {
	paymentsRegistered.Store(true)
	return nil
}

func (Payments) Describe() goralens.WatcherInfo {
	return goralens.WatcherInfo{
		Title: "Payments",
		Group: "Billing",
		Hint:  "Payments taken at checkout.",
		Columns: []goralens.Column{
			{Label: "Order", Field: "order"},
			{Label: "Amount", Field: "amount"},
			{Label: "Status", Field: "status"},
		},
	}
}

func RecordPayment(ctx context.Context, order string, amount float64, status string) {
	e := goralens.NewEntry(PaymentType, map[string]any{"order": order, "amount": amount, "status": status})
	e.AddTags("order:" + order)
	goralens.Record(ctx, e)
}

type Disabled struct{}

func (Disabled) Name() string { return "disabled_demo" }

func (Disabled) Register(foundation.Application) error {
	disabledRegistered.Store(true)
	return nil
}

func PaymentsRegistered() bool { return paymentsRegistered.Load() }

func DisabledRegistered() bool { return disabledRegistered.Load() }
