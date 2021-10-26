package util

import "bytes"

type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	var b bytes.Buffer
	b.WriteString("Errors:\n")
	for _, err := range m.Errors {
		b.WriteString("- " + err.Error() + "\n")
	}
	return b.String()
}
