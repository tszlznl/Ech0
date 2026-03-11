package bus

import "reflect"

func safeTypeString(t reflect.Type) string {
	if t == nil {
		return ""
	}
	return t.String()
}
