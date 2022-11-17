package desire

import (
	"fmt"
	"strings"
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
