package wrapper

import (
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/category-service/pkg/middleware"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"go.uber.org/ratelimit"
)

// HOC-like function to attach required observability (tracing, logging & metrics) and
// resiliency patterns to category's use case layer using chain-of-responsibility pattern
func WrapCategoryMiddleware(svcUnwrap service.Category, logger log.Logger) service.Category {
	var svc service.Category
	svc = svcUnwrap
	svc = middleware.CategoryLog{Logger: logger, Next: svc}
	svc = middleware.CategoryResiliency{
		Logger:      logger,
		RateLimiter: ratelimit.New(100),
		Next:        svc,
	}
	return svc
}
