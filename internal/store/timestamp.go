package store

import (
	"reflect"
	"time"
)

const (
	fieldCreatedAt = "CreatedAt"
	fieldUpdatedAt = "UpdatedAt"
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

func getCreationTimestamp(obj interface{}) int32 {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return 0
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(fieldCreatedAt)
	return int32(ts.Int())
}

func setCreationTimestamp(obj interface{}, val int32) {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(fieldCreatedAt)
	ts.Set(reflect.ValueOf(val))
}

func addTS(obj interface{}) {
	// TODO(hbagdi): make this configuration for update operations
	setTSField(obj, fieldUpdatedAt, true)
	setTSField(obj, fieldCreatedAt, true)
}
