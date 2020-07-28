package mw

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"strings"
)

type CategoryEventLog struct {
	Logger log.Logger
	Next   domain.CategoryEventBus
}

func (c CategoryEventLog) Created(ctx context.Context, category domain.Category) (err error) {
	defer func() {
		if err != nil {
			_ = level.Error(c.Logger).Log(
				"err", err,
			)
			return
		}
		_ = level.Info(c.Logger).Log(
			"msg", fmt.Sprintf("event %s successfully sent", domain.CategoryCreated),
			"event_name", domain.CategoryCreated,
			"kind", strings.ToLower(eventbus.EventDomain),
		)
	}()

	err = c.Next.Created(ctx, category)
	return
}

func (c CategoryEventLog) Updated(ctx context.Context, category domain.Category) (err error) {
	defer func() {
		if err != nil {
			_ = level.Error(c.Logger).Log(
				"err", err,
			)
			return
		}
		_ = level.Info(c.Logger).Log(
			"msg", fmt.Sprintf("event %s successfully sent", domain.CategoryUpdated),
			"event_name", domain.CategoryUpdated,
			"kind", strings.ToLower(eventbus.EventDomain),
		)
	}()

	err = c.Next.Updated(ctx, category)
	return
}

func (c CategoryEventLog) Removed(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			_ = level.Error(c.Logger).Log(
				"err", err,
			)
			return
		}
		_ = level.Info(c.Logger).Log(
			"msg", fmt.Sprintf("event %s successfully sent", domain.CategoryRemoved),
			"event_name", domain.CategoryRemoved,
			"kind", strings.ToLower(eventbus.EventDomain),
		)
	}()

	err = c.Next.Removed(ctx, id)
	return
}

func (c CategoryEventLog) Restored(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			_ = level.Error(c.Logger).Log(
				"err", err,
			)
			return
		}
		_ = level.Info(c.Logger).Log(
			"msg", fmt.Sprintf("event %s successfully sent", domain.CategoryRestored),
			"event_name", domain.CategoryRestored,
			"kind", strings.ToLower(eventbus.EventDomain),
		)
	}()

	err = c.Next.Restored(ctx, id)
	return
}

func (c CategoryEventLog) HardRemoved(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			_ = level.Error(c.Logger).Log(
				"err", err,
			)
			return
		}
		_ = level.Info(c.Logger).Log(
			"msg", fmt.Sprintf("event %s successfully sent", domain.CategoryHardRemoved),
			"event_name", domain.CategoryHardRemoved,
			"kind", strings.ToLower(eventbus.EventDomain),
		)
	}()

	err = c.Next.HardRemoved(ctx, id)
	return
}
