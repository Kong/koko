package status

import "fmt"

type Code string

const (
	DPMissingPlugin = "DP001"
)

var messages = map[Code]string{
	DPMissingPlugin: "kong data-plane node missing plugin",
}

// MessageForCode returns message with additional context derived from code.
func MessageForCode(code Code, message string) string {
	codeMessage := messages[code]
	return fmt.Sprintf("%s[%s]: %s", codeMessage, code, message)
}
