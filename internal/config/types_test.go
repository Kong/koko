package config

import (
	"reflect"
	"testing"
)

// This test ensures that no slices are added to configuration.
// Slices complicated environment variable based configuration and are not
// allowed.
func TestConfigNoArray(t *testing.T) {
	var c Config
	v := reflect.ValueOf(c)
	walk(v, t)
}

func walk(v reflect.Value, t *testing.T) {
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch typ.Field(i).Type.Kind() { //nolint:exhaustive
		case reflect.Slice:
			t.Error("slices are not allowed in configuration")
		case reflect.Array:
			t.Error("arrays are not allowed in configuration")
		case reflect.Interface:
			t.Error("interfaces are not allowed in configuration")
		case reflect.Struct:
			walk(v.Field(i), t)
		default:
		}
	}
}
