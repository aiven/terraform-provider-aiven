package adapter

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type warningCollector struct {
	mu    sync.Mutex
	diags diag.Diagnostics
}

type warningCollectorKey struct{}

func withWarnings(ctx context.Context, target *diag.Diagnostics) (context.Context, func()) {
	if target == nil {
		return ctx, func() {}
	}

	collector := &warningCollector{}
	ctx = context.WithValue(ctx, warningCollectorKey{}, collector)

	return ctx, func() {
		collector.mu.Lock()
		defer collector.mu.Unlock()

		target.Append(collector.diags...)
		collector.diags = nil
	}
}

// AddWarning appends a warning diagnostic to the request-scoped collector stored in ctx.
// If the context has no collector, the warning is ignored.
func AddWarning(ctx context.Context, summary, detail string) {
	AddWarnings(ctx, diag.NewWarningDiagnostic(summary, detail))
}

// AddAttributeWarning appends an attribute warning diagnostic to the request-scoped collector stored in ctx.
// If the context has no collector, the warning is ignored.
func AddAttributeWarning(ctx context.Context, attributePath path.Path, summary, detail string) {
	AddWarnings(ctx, diag.NewAttributeWarningDiagnostic(attributePath, summary, detail))
}

// AddWarnings appends diagnostic warnings to the request-scoped collector stored in ctx.
// If the context has no collector, the diagnostics are ignored.
func AddWarnings(ctx context.Context, in ...diag.Diagnostic) {
	collector, ok := ctx.Value(warningCollectorKey{}).(*warningCollector)
	if !ok || collector == nil {
		return
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	collector.diags.Append(diag.Diagnostics(in).Warnings()...)
}
