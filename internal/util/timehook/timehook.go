package timehook

import (
	"reflect"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// NormalizeModelTimesToUTC 将模型中的 time.Time/*time.Time 字段归一为 UTC。
// 约定用于 GORM hook（BeforeCreate/BeforeUpdate）中调用。
func NormalizeModelTimesToUTC(model any) {
	if model == nil {
		return
	}
	v := reflect.ValueOf(model)
	normalizeValue(v)
}

// normalizeValue 递归遍历可写字段，仅将 time.Time 与 *time.Time 归一为 UTC。
// 说明：
// 1) 非结构体或不可写字段会直接跳过，避免反射 panic。
// 2) 指针会先解引用，再继续处理其指向的值。
// 3) 零值时间不做转换，保留 GORM 自动时间填充语义。
func normalizeValue(v reflect.Value) {
	if !v.IsValid() {
		return
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		// *time.Time 直接转换；其他指针类型继续递归解引用。
		if v.Type().Elem() == timeType {
			t := v.Elem().Interface().(time.Time)
			if !t.IsZero() && v.Elem().CanSet() {
				v.Elem().Set(reflect.ValueOf(t.UTC()))
			}
			return
		}
		normalizeValue(v.Elem())
		return
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		ft := field.Type()
		switch {
		case ft == timeType:
			// 将 time.Time 字段统一转换为 UTC。
			t := field.Interface().(time.Time)
			if !t.IsZero() {
				field.Set(reflect.ValueOf(t.UTC()))
			}
		case ft.Kind() == reflect.Pointer && ft.Elem() == timeType:
			// 将 *time.Time 指针指向的时间统一转换为 UTC。
			if !field.IsNil() {
				t := field.Elem().Interface().(time.Time)
				if !t.IsZero() {
					field.Elem().Set(reflect.ValueOf(t.UTC()))
				}
			}
		case field.Kind() == reflect.Struct:
			// 嵌套结构体继续递归处理其内部时间字段。
			normalizeValue(field)
		}
	}
}

