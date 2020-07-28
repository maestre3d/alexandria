package mw

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"go.opencensus.io/trace"
)

type CategoryRepositoryTracing struct {
	Next domain.CategoryRepository
}

func (c CategoryRepositoryTracing) Save(ctx context.Context, category domain.Category) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_save")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "save"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "write row in cassandra",
			})
		}

		span.End()
	}()

	err = c.Next.Save(ctx, category)
	return
}

func (c CategoryRepositoryTracing) Fetch(ctx context.Context, params core.PaginationParams,
	filter core.FilterParams) (categories []*domain.Category, err error) {
	ctxT, span := trace.StartSpan(ctx, "category_fetch")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "fetch"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "read row in cassandra",
			})
		}

		span.End()
	}()

	categories, err = c.Next.Fetch(ctx, params, filter)
	return
}

func (c CategoryRepositoryTracing) FetchByID(ctx context.Context, id string, activeOnly bool) (category *domain.Category, err error) {
	ctxT, span := trace.StartSpan(ctx, "category_fetch_by_id")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "fetch_by_id"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "read row in cassandra",
			})
		}

		span.End()
	}()

	category, err = c.Next.FetchByID(ctx, id, activeOnly)
	return
}

func (c CategoryRepositoryTracing) Replace(ctx context.Context, category domain.Category) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_replace")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "replace"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "write row in cassandra",
			})
		}

		span.End()
	}()

	err = c.Next.Replace(ctx, category)
	return
}

func (c CategoryRepositoryTracing) Remove(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_remove")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "remove"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "write row in cassandra",
			})
		}

		span.End()
	}()

	err = c.Next.Remove(ctx, id)
	return
}

func (c CategoryRepositoryTracing) Restore(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_restore")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "restore"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "write row in cassandra",
			})
		}

		span.End()
	}()

	err = c.Next.Restore(ctx, id)
	return
}

func (c CategoryRepositoryTracing) HardRemove(ctx context.Context, id string) (err error) {
	ctxT, span := trace.StartSpan(ctx, "category_hard_remove")
	ctx = ctxT
	span.AddAttributes(trace.StringAttribute("operation", "hard_remove"), trace.StringAttribute("db.driver", "cassandra"))

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "write row in cassandra",
			})
		}

		span.End()
	}()

	err = c.Next.HardRemove(ctx, id)
	return
}
