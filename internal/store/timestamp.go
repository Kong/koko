package store

import (
	"reflect"
	"time"
)

func setTSField(obj interface{}, field string, override bool) {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(field)
	if ts.Int() == 0 || override {
		ts.Set(reflect.ValueOf(int32(time.Now().Unix())))
		return
	}
}

func addTS(obj interface{}) {
	// TODO(hbagdi): make this configuration for update operations
	setTSField(obj, "CreatedAt", true)
	setTSField(obj, "UpdatedAt", true)
}
