package validation

import (
	"bytes"
	"fmt"
)

type FieldError struct {
	Name    string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
}

type Error struct {
	Fields []FieldError
}

func (e Error) Error() string {
	var buf bytes.Buffer
	errCount := len(e.Fields)
	for i, field := range e.Fields {
		buf.WriteString(field.Error())
		if i < errCount-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}
