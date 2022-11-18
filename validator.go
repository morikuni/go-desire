package desire

import (
	"fmt"
	"reflect"
	"strconv"
)

type Fields map[string]any

func (fs Fields) Validate(ctx ValidationContext, got any) {
	gotRV := reflect.ValueOf(got)
	gotRV = reflect.Indirect(gotRV)
	switch gotRV.Kind() {
	case reflect.Map:
		if gotRV.Type().Key().Kind() != reflect.String {
			ctx.Rejectf("expected key type string but got %s", gotRV.Type().Key())
		}
		keySet := make(map[string]reflect.Value)
		for _, k := range gotRV.MapKeys() {
			keySet[k.String()] = k
		}
		for k := range fs {
			if _, ok := keySet[k]; !ok {
				keySet[k] = reflect.ValueOf(k)
			}
		}
		for k, krv := range keySet {
			gotIndex := gotRV.MapIndex(krv)
			v, ok := fs[k]
			if !ok {
				ctx.WithField(k).Rejectf("expected undefined but exists with value %v", gotIndex.Interface())
				continue
			}
			if !gotIndex.IsValid() {
				ctx.WithField(k).Rejectf("expected %v but undefined", v)
				continue
			}
			tv := gotIndex.Interface()
			Validate(ctx.WithField(k), tv, v)
		}
	case reflect.Struct:
		keySet := make(map[string]struct{})
		gotRT := gotRV.Type()
		for i, n := 0, gotRT.NumField(); i < n; i++ {
			keySet[gotRT.Field(i).Name] = struct{}{}
		}
		for k := range fs {
			if _, ok := keySet[k]; !ok {
				keySet[k] = struct{}{}
			}
		}
		for k := range keySet {
			gotIndex := gotRV.FieldByName(k)
			v, ok := fs[k]
			if !ok {
				ctx.WithField(k).Rejectf("expected undefined but exists with value %v", gotIndex.Interface())
				continue
			}
			if !gotIndex.IsValid() {
				ctx.WithField(k).Rejectf("expected %v but undefined", v)
				continue
			}
			tv := gotIndex.Interface()
			Validate(ctx.WithField(k), tv, v)
		}
	default:
		ctx.Rejectf("expected map or struct but got %s", gotRV.Kind())
	}
}

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
			Validate(ctx.WithField(fmt.Sprint(k)), tv, v)
		}
	case reflect.Struct:
		for k, v := range p {
			switch k.(type) {
			case string:
				ks := k.(string)
				fieldValue := gotRV.FieldByName(ks)
				if !fieldValue.IsValid() {
					ctx.WithField(ks).Rejectf("expected %v but undefined", v)
					continue
				}
				tv := fieldValue.Interface()
				Validate(ctx.WithField(ks), tv, v)
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
				Validate(ctx.WithField(strconv.Itoa(idx)), tv, v)
			default:
				ctx.Reject("key type of Partial must be int for slice or array")
			}
		}
	default:
		ctx.Rejectf("expected slice, array, map or struct but got %s", gotRV.Kind())
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
			Validate(ctx.WithField(strconv.Itoa(i)), tv, ls[i])
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
			rs := Desire(got, candidate)
			if len(rs) == 0 {
				return
			}
		}
		ctx.Rejectf("expected one of %v but got %v", candidates, got)
	})
}

func All(vs ...Validator) Validator {
	return ValidatorFunc(func(ctx ValidationContext, got any) {
		for _, v := range vs {
			v.Validate(ctx, got)
		}
	})
}

func Any() Validator {
	return ValidatorFunc(func(ctx ValidationContext, got any) {})
}

//
//func Struct(strct any, override ...OverWrite) Validator {
//	return ValidatorFunc(func(ctx ValidationContext, got any) {
//		strctRV := reflect.ValueOf(strct)
//		gotRV := reflect.ValueOf(got)
//		if got, expected := gotRV.Type(), strctRV.Type(); got != expected {
//			ctx.Rejectf("expected type %s but got %s", expected, got)
//			return
//		}
//		strctRV = reflect.Indirect(strctRV)
//		if strctRV.Kind() != reflect.Struct {
//			ctx.Rejectf("expected struct but got %s", strctRV.Kind())
//			return
//		}
//		gotRV = reflect.Indirect(gotRV)
//		strctType := strctRV.Type()
//		for i := 0; i < strctType.NumField(); i++ {
//			field := strctType.Field(i)
//			if field.PkgPath != "" {
//				continue
//			}
//			gotField := gotRV.FieldByName(field.Name)
//			if !gotField.IsValid() {
//				ctx.WithField(field.Name).Rejectf("expected %v but undefined", strctRV.Field(i).Interface())
//				continue
//			}
//			Validate(ctx.WithField(field.Name), gotField.Interface(), strctRV.Field(i).Interface())
//		}
//	})
//}
//
//type OverWrite struct {
//	Path      Path
//	Validator Validator
//}
