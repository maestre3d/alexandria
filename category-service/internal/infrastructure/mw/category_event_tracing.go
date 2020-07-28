package mw

import (
	"context"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"go.opencensus.io/trace"
)

type CategoryEventTracing struct {
	Next domain.CategoryEventBus
}

func (c CategoryEventTracing) Created(ctx context.Context, category domain.Category) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_created")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryCreated), trace.StringAttribute("event.type", eventbus.EventDomain))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "created event sent",
			})
		}

		span.End()
	}()

	err = c.Next.Created(ctx, category)
	return
}

func (c CategoryEventTracing) Updated(ctx context.Context, category domain.Category) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_updated")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryUpdated), trace.StringAttribute("event.type", eventbus.EventDomain))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "updated event sent",
			})
		}

		span.End()
	}()

	err = c.Next.Updated(ctx, category)
	return
}

func (c CategoryEventTracing) Removed(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_removed")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryRemoved), trace.StringAttribute("event.type", eventbus.EventDomain))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "removed event sent",
			})
		}

		span.End()
	}()

	err = c.Next.Removed(ctx, id)
	return
}

func (c CategoryEventTracing) Restored(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_restored")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryRestored), trace.StringAttribute("event.type", eventbus.EventDomain))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "restored event sent",
			})
		}

		span.End()
	}()

	err = c.Next.Restored(ctx, id)
	return
}

func (c CategoryEventTracing) HardRemoved(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_hard_remove")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryHardRemoved), trace.StringAttribute("event.type", eventbus.EventDomain))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "hard_removed event sent",
			})
		}

		span.End()
	}()

	return c.Next.HardRemoved(ctx, id)
}
