package desire

import (
	"fmt"
)

type ValidationContext interface {
	Reject(reason string)
	Rejectf(format string, args ...interface{})
	WithField(field string) ValidationContext
}

type validationContext struct {
	parent     *validationContext
	path       string
	rejections *[]Rejection
}

func newRootValidationContext() validationContext {
	return validationContext{
		rejections: new([]Rejection),
	}
}

func pathOf(ctx validationContext, depth int) Path {
	if ctx.parent == nil {
		if depth == 0 {
			return nil
		}
		return make(Path, 0, depth)
	}
	return append(pathOf(*ctx.parent, depth+1), ctx.path)
}

func (ctx validationContext) Reject(reason string) {
	steps := pathOf(ctx, 0)
	*ctx.rejections = append(*ctx.rejections, Rejection{
		steps,
		reason,
	})
}

func (ctx validationContext) Rejectf(format string, args ...interface{}) {
	ctx.Reject(fmt.Sprintf(format, args...))
}

func (ctx validationContext) WithField(path string) ValidationContext {
	return validationContext{
		parent:     &ctx,
		path:       path,
		rejections: ctx.rejections,
	}
}

func addRejections(ctx ValidationContext, rs []Rejection) {
	for _, r := range rs {
		tmpCtx := ctx
		for _, step := range r.Path {
			tmpCtx = tmpCtx.WithField(step)
		}
		tmpCtx.Reject(r.Reason)
	}
}
