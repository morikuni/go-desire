package desire

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/go-cmp/cmp"
)

type Partial map[any]any

func (p Partial) Validate(ctx ValidationContext, got any) {
	gotRV := reflect.ValueOf(got)
	gotRV = reflect.Indirect(gotRV)
	switch gotRV.Kind() {
	case reflect.Map:
		for k, v := range p {
			keyRV := reflect.ValueOf(k)
			if !keyRV.Type().AssignableTo(gotRV.Type().Key()) {
				ctx.WithField(fmt.Sprint(k)).Rejectf("expected key type %s but got %s", keyRV.Type(), gotRV.Type().Key())
				continue
			}
			indexValue := gotRV.MapIndex(keyRV)
			if !indexValue.IsValid() {
				ctx.WithField(fmt.Sprint(k)).Rejectf("expected %v but undefined", v)
				continue
			}
			tv := indexValue.Interface()
			validate(ctx.WithField(fmt.Sprint(k)), tv, v)
		}
	case reflect.Struct:
		for k, v := range p {
			switch k.(type) {
			case string:
				ks := k.(string)
				tv := gotRV.FieldByName(ks).Interface()
				validate(ctx.WithField(ks), tv, v)
			default:
				ctx.Reject("key type of Partial must be string for struct")
			}
		}
	case reflect.Slice, reflect.Array:
		gotLen := gotRV.Len()
		for k, v := range p {
			switch k.(type) {
			case int:
				idx := k.(int)
				if idx < 0 || idx >= gotLen {
					ctx.WithField(fmt.Sprint(k)).Rejectf("index out of range for size %d", gotLen)
					continue
				}
				indexValue := gotRV.Index(idx)
				if !indexValue.IsValid() {
					ctx.WithField(fmt.Sprint(k)).Rejectf("expected %v but undefined", v)
					continue
				}
				tv := indexValue.Interface()
				validate(ctx.WithField(strconv.Itoa(idx)), tv, v)
			default:
				ctx.Reject("key type of Partial must be int for slice or array")
			}
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
