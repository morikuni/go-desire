package desire

import (
	"reflect"
	"strconv"

	"github.com/google/go-cmp/cmp"
)

type Partial map[string]any

func (p Partial) Validate(ctx ValidationContext, got any) {
	gotRV := reflect.ValueOf(got)
	gotRV = reflect.Indirect(gotRV)
	switch gotRV.Kind() {
	case reflect.Map:
		for k, v := range p {
			indexValue := gotRV.MapIndex(reflect.ValueOf(k))
			if !indexValue.IsValid() {
				ctx.WithField(k).Rejectf("expected %v but undefined", v)
				continue
			}
			tv := indexValue.Interface()
			validate(ctx.WithField(k), tv, v)
		}
	case reflect.Struct:
		for k, v := range p {
			tv := gotRV.FieldByName(k).Interface()
			validate(ctx.WithField(k), tv, v)
		}
	default:
		ctx.Rejectf("expected map or struct but got %s", gotRV.Kind())
	}
}

type List []any

func (ls List) Validate(ctx ValidationContext, got any) {
	gotRV := reflect.ValueOf(got)
	gotRV = reflect.Indirect(gotRV)
	switch gotRV.Kind() {
	case reflect.Array, reflect.Slice:
		min := gotRV.Len()
		if l := len(ls); l < min {
			min = l
		}
		for i := 0; i < min; i++ {
			tv := gotRV.Index(i).Interface()
			validate(ctx.WithField(strconv.Itoa(i)), tv, ls[i])
		}
		for i := min; i < len(ls); i++ {
			ctx.WithField(strconv.Itoa(i)).Rejectf("expected %v but undefined", ls[i])
		}
		for i := min; i < gotRV.Len(); i++ {
			tv := gotRV.Index(i).Interface()
			ctx.WithField(strconv.Itoa(i)).Rejectf("expected undefined but exists with value %v", tv)
		}
	default:
		ctx.Rejectf("expected array or slice but got %s", gotRV.Kind())
	}
}

func validate(ctx ValidationContext, got, desire any) {
	switch desire := desire.(type) {
	case Validator:
		desire.Validate(ctx, got)
	default:
		r := &cmpReporter{}
		_ = cmp.Equal(desire, got, cmp.Reporter(r))
		for i := range r.rejections {
			tmpCtx := ctx
			for _, path := range r.rejections[i].Path {
				tmpCtx = tmpCtx.WithField(path)
			}
			tmpCtx.Reject(r.rejections[i].Reason)
		}
	}
}

func NotZero() Validator {
	return ValidatorFunc(func(ctx ValidationContext, got any) {
		if reflect.ValueOf(got).IsZero() {
			ctx.Rejectf("expected non-zero value but got %v", got)
		}
	})
}

func NotZeroT[T comparable]() Validator {
	var zero T
	return ValidatorFunc(func(ctx ValidationContext, got any) {
		if v, ok := got.(T); ok {
			if v == zero {
				ctx.Rejectf("expected non-zero value but got %v", got)
			}
		} else {
			ctx.Rejectf("expected type %T but got %T", zero, got)
		}
	})
}

func OneOf(candidates ...any) Validator {
	return ValidatorFunc(func(ctx ValidationContext, got any) {
		for _, candidate := range candidates {
			if candidate == got {
				return
			}
		}
		ctx.Rejectf("expected one of %v but got %v", candidates, got)
	})
}
