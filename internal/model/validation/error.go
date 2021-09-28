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
	len := len(e.Fields)
	for i, field := range e.Fields {
		buf.WriteString(field.Error())
		if i < len-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}
