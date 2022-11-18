package desire

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type Validator interface {
	Validate(ValidationContext, any)
}

type ValidatorFunc func(ValidationContext, any)

func (f ValidatorFunc) Validate(ctx ValidationContext, got any) {
	f(ctx, got)
}

func Desire(got, desire any) []Rejection {
	root := newRootValidationContext()
	Validate(root, got, desire)
	return *root.rejections
}

func Validate(ctx ValidationContext, got, desire any) {
	switch desire := desire.(type) {
	case Validator:
		desire.Validate(ctx, got)
	default:
		r := &cmpReporter{}
		_ = cmp.Equal(desire, got, cmp.Reporter(r))
		addRejections(ctx, r.rejections)
	}
}

type Rejection struct {
	Path   Path
	Reason string
}

func (r Rejection) String() string {
	return fmt.Sprintf("%s: %s", r.Path, r.Reason)
}

type Path []string

func (p Path) String() string {
	return strings.Join(p, ".")
}
